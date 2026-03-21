# Secret Rotation Procedures

## SSO_ENCRYPTION_KEY

The `SSO_ENCRYPTION_KEY` encrypts OIDC client secrets stored in the database. Rotation requires re-encrypting existing secrets.

### Steps

1. Generate a new key:
   ```bash
   openssl rand -hex 32
   ```

2. Set both old and new keys on the API (when the app supports dual-key decryption):
   ```bash
   SSO_ENCRYPTION_KEY=<new-key>
   SSO_ENCRYPTION_KEY_OLD=<old-key>
   ```

3. Restart the API. It will re-encrypt all stored client secrets with the new key.

4. Once re-encryption is confirmed, remove `SSO_ENCRYPTION_KEY_OLD` and restart.

### Until dual-key support is implemented

1. Generate a new key.
2. Decrypt all OIDC client secrets using the current key (via admin API or direct DB access).
3. Update `SSO_ENCRYPTION_KEY` in `.env` with the new key.
4. Re-encrypt all client secrets with the new key.
5. Restart the API.

## VAPID Keys

VAPID key rotation invalidates all existing push subscriptions. Users must re-subscribe.

### Steps

1. Generate a new keypair:
   ```bash
   npx web-push generate-vapid-keys
   ```

2. Update `VAPID_PUBLIC_KEY` and `VAPID_PRIVATE_KEY` in `.env`.

3. Restart the API.

4. The frontend will detect the key change and prompt users to re-subscribe.

## JWT_SECRET

Rotating `JWT_SECRET` invalidates all active sessions. Users must re-login.

1. Generate: `openssl rand -hex 32`
2. Update in `.env` and restart.
