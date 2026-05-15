package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/omnir/crm-api/internal/auth"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

const backupCodeCount = 8

// TwoFAHandler handles TOTP 2FA setup, verification, and disabling.
type TwoFAHandler struct {
	totpRepo   repository.TOTPRepository
	users      repository.UserRepository
	jwtSvc     *auth.JWTService
	encryptKey string
	auditor    Auditor
}

func NewTwoFAHandler(
	totpRepo repository.TOTPRepository,
	users repository.UserRepository,
	jwtSvc *auth.JWTService,
	encryptKey string,
) *TwoFAHandler {
	return &TwoFAHandler{totpRepo: totpRepo, users: users, jwtSvc: jwtSvc, encryptKey: encryptKey}
}

func (h *TwoFAHandler) WithAuditLog(r repository.AuditLogRepository) *TwoFAHandler {
	h.auditor = newAuditor(r)
	return h
}

// Router returns routes for authenticated user 2FA management.
func (h *TwoFAHandler) Router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.Status)
	r.Post("/setup", h.Setup)
	r.Post("/verify", h.Verify)
	r.Post("/disable", h.Disable)
	return r
}

// LoginRouter returns routes for 2FA verification during the auth flow.
func (h *TwoFAHandler) LoginRouter() chi.Router {
	r := chi.NewRouter()
	r.Post("/verify", h.VerifyLogin)
	return r
}

// Status reports whether TOTP 2FA is enabled for the current user.
// GET /api/v1/users/me/2fa
func (h *TwoFAHandler) Status(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)

	enabled, err := h.totpRepo.IsEnabled(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"enabled": enabled})
}

// Setup generates a TOTP secret and returns the otpauth:// URI + QR PNG data URL.
// POST /api/v1/users/me/2fa/setup
func (h *TwoFAHandler) Setup(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeError(w, http.StatusInternalServerError, "user not found")
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Omnir",
		AccountName: user.Email,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate TOTP key")
		return
	}

	// Encrypt secret before storing.
	encSecret, err := auth.Encrypt(h.encryptKey, key.Secret())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encryption error")
		return
	}
	if err := h.totpRepo.SetSecret(r.Context(), claims.UserID, encSecret); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Generate QR PNG → data URL.
	img, err := key.Image(200, 200)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "QR generation error")
		return
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		writeError(w, http.StatusInternalServerError, "QR encoding error")
		return
	}
	qrDataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	writeJSON(w, http.StatusOK, domain.TOTPSetupResult{
		OTPAuthURI: key.URL(),
		QRDataURL:  qrDataURL,
	})
}

// Verify confirms a TOTP code, activates 2FA, and returns 8 backup codes.
// POST /api/v1/users/me/2fa/verify
func (h *TwoFAHandler) Verify(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	encSecret, err := h.totpRepo.GetSecret(r.Context(), claims.UserID)
	if err != nil || encSecret == "" {
		writeError(w, http.StatusBadRequest, "2FA setup not started")
		return
	}

	secret, err := auth.Decrypt(h.encryptKey, encSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "decryption error")
		return
	}

	if !totp.Validate(req.Code, secret) {
		writeError(w, http.StatusUnprocessableEntity, "invalid TOTP code")
		return
	}

	// Generate 8 backup codes, bcrypt-hash them.
	plainCodes := make([]string, backupCodeCount)
	hashedCodes := make([]string, backupCodeCount)
	for i := range plainCodes {
		code, err := randomBackupCode()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "backup code generation failed")
			return
		}
		plainCodes[i] = code
		hashed, err := bcrypt.GenerateFromPassword([]byte(plainCodes[i]), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "backup code generation failed")
			return
		}
		hashedCodes[i] = string(hashed)
	}

	if err := h.totpRepo.Activate(r.Context(), claims.UserID, hashedCodes); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, domain.TOTPSetupResult{BackupCodes: plainCodes})
}

// Disable turns off 2FA after password confirmation.
// POST /api/v1/users/me/2fa/disable
func (h *TwoFAHandler) Disable(w http.ResponseWriter, r *http.Request) {
	claims := mustClaims(r)

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeError(w, http.StatusInternalServerError, "user not found")
		return
	}
	_, hash, err := h.users.FindByEmail(r.Context(), user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if hash == "" || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid password")
		return
	}

	if err := h.totpRepo.Disable(r.Context(), claims.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// VerifyLogin validates a TOTP code (or backup code) during the auth flow.
// POST /api/v1/auth/2fa/verify
func (h *TwoFAHandler) VerifyLogin(w http.ResponseWriter, r *http.Request) {
	// This endpoint is called with a partial JWT that carries user_id but 2FA is not yet confirmed.
	// For simplicity, we accept a bearer token with the same claims.
	claims := mustClaims(r)

	var req struct {
		Code       string `json:"code"`
		BackupCode string `json:"backup_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.Code != "" {
		// Validate TOTP code.
		encSecret, err := h.totpRepo.GetSecret(r.Context(), claims.UserID)
		if err != nil || encSecret == "" {
			writeError(w, http.StatusBadRequest, "2FA not configured")
			return
		}
		secret, err := auth.Decrypt(h.encryptKey, encSecret)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "decryption error")
			return
		}
		if !totp.Validate(req.Code, secret) {
			writeError(w, http.StatusUnauthorized, "invalid 2FA code")
			return
		}
	} else if req.BackupCode != "" {
		// Try backup codes.
		codes, err := h.totpRepo.FindUnusedBackupCode(r.Context(), claims.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		var matched *domain.TOTPBackupCode
		for _, c := range codes {
			if bcrypt.CompareHashAndPassword([]byte(c.CodeHash), []byte(req.BackupCode)) == nil {
				matched = c
				break
			}
		}
		if matched == nil {
			writeError(w, http.StatusUnauthorized, "invalid backup code")
			return
		}
		if err := h.totpRepo.MarkBackupCodeUsed(r.Context(), matched.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
	} else {
		writeError(w, http.StatusBadRequest, "code or backup_code required")
		return
	}

	// Issue fresh full JWTs.
	user, err := h.users.GetByID(r.Context(), claims.UserID)
	if err != nil || user == nil {
		writeError(w, http.StatusInternalServerError, "user not found")
		return
	}
	newClaims := auth.Claims{UserID: user.ID, OrgID: user.OrgID, Role: string(user.Role)}
	accessToken, err := h.jwtSvc.Issue(newClaims, accessTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	refreshToken, err := h.jwtSvc.Issue(newClaims, refreshTokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	secure := r.TLS != nil
	setAccessCookie(w, accessToken, secure)
	setRefreshCookie(w, refreshToken, secure)
	h.auditor.logLogin(r, user)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func randomBackupCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
