package handler

import (
	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
)

// buildContactChanges compares old contact state against a patch and returns
// the set of fields that changed with their before/after values.
func buildContactChanges(old *domain.Contact, patch domain.ContactPatch) domain.AuditChanges {
	changes := domain.AuditChanges{}
	if patch.FirstName != nil && *patch.FirstName != old.FirstName {
		changes["first_name"] = domain.FieldChange{From: old.FirstName, To: *patch.FirstName}
	}
	if patch.LastName != nil && *patch.LastName != old.LastName {
		changes["last_name"] = domain.FieldChange{From: old.LastName, To: *patch.LastName}
	}
	if patch.Email != nil && !ptrStrEq(old.Email, patch.Email) {
		changes["email"] = domain.FieldChange{From: ptrStrVal(old.Email), To: ptrStrVal(patch.Email)}
	}
	if patch.Phone != nil && !ptrStrEq(old.Phone, patch.Phone) {
		changes["phone"] = domain.FieldChange{From: ptrStrVal(old.Phone), To: ptrStrVal(patch.Phone)}
	}
	if patch.AccountID != nil && !ptrUUIDEq(old.AccountID, patch.AccountID) {
		changes["account_id"] = domain.FieldChange{From: ptrUUIDVal(old.AccountID), To: ptrUUIDVal(patch.AccountID)}
	}
	if patch.OwnerID != nil && *patch.OwnerID != old.OwnerID {
		changes["owner_id"] = domain.FieldChange{From: old.OwnerID.String(), To: patch.OwnerID.String()}
	}
	if patch.LeadSource != nil && !ptrStrEq(old.LeadSource, patch.LeadSource) {
		changes["lead_source"] = domain.FieldChange{From: ptrStrVal(old.LeadSource), To: ptrStrVal(patch.LeadSource)}
	}
	if patch.Stage != nil && *patch.Stage != old.Stage {
		changes["stage"] = domain.FieldChange{From: string(old.Stage), To: string(*patch.Stage)}
	}
	if len(changes) == 0 {
		return nil
	}
	return changes
}

// buildAccountChanges compares old account state against a patch.
func buildAccountChanges(old *domain.Account, patch domain.AccountPatch) domain.AuditChanges {
	changes := domain.AuditChanges{}
	if patch.Name != nil && *patch.Name != old.Name {
		changes["name"] = domain.FieldChange{From: old.Name, To: *patch.Name}
	}
	if patch.Domain != nil && !ptrStrEq(old.Domain, patch.Domain) {
		changes["domain"] = domain.FieldChange{From: ptrStrVal(old.Domain), To: ptrStrVal(patch.Domain)}
	}
	if patch.Industry != nil && !ptrStrEq(old.Industry, patch.Industry) {
		changes["industry"] = domain.FieldChange{From: ptrStrVal(old.Industry), To: ptrStrVal(patch.Industry)}
	}
	if patch.Size != nil {
		var oldSize, newSize interface{}
		if old.Size != nil {
			oldSize = string(*old.Size)
		}
		newSize = string(*patch.Size)
		if oldSize != newSize {
			changes["size"] = domain.FieldChange{From: oldSize, To: newSize}
		}
	}
	if patch.OwnerID != nil && *patch.OwnerID != old.OwnerID {
		changes["owner_id"] = domain.FieldChange{From: old.OwnerID.String(), To: patch.OwnerID.String()}
	}
	if len(changes) == 0 {
		return nil
	}
	return changes
}

// buildDealChanges compares old deal state against a patch.
func buildDealChanges(old *domain.Deal, patch domain.DealPatch) domain.AuditChanges {
	changes := domain.AuditChanges{}
	if patch.Title != nil && *patch.Title != old.Title {
		changes["title"] = domain.FieldChange{From: old.Title, To: *patch.Title}
	}
	if patch.Stage != nil && *patch.Stage != old.Stage {
		changes["stage"] = domain.FieldChange{From: string(old.Stage), To: string(*patch.Stage)}
	}
	if patch.ValueCents != nil && *patch.ValueCents != old.ValueCents {
		changes["value_cents"] = domain.FieldChange{From: old.ValueCents, To: *patch.ValueCents}
	}
	if patch.Currency != nil && *patch.Currency != old.Currency {
		changes["currency"] = domain.FieldChange{From: old.Currency, To: *patch.Currency}
	}
	if patch.Probability != nil && *patch.Probability != old.Probability {
		changes["probability"] = domain.FieldChange{From: old.Probability, To: *patch.Probability}
	}
	if patch.ExpectedCloseDate != nil {
		oldDate := ""
		if old.ExpectedCloseDate != nil {
			oldDate = old.ExpectedCloseDate.Format("2006-01-02")
		}
		newDate := patch.ExpectedCloseDate.Format("2006-01-02")
		if newDate != oldDate {
			changes["expected_close_date"] = domain.FieldChange{From: oldDate, To: newDate}
		}
	}
	if patch.OwnerID != nil && *patch.OwnerID != old.OwnerID {
		changes["owner_id"] = domain.FieldChange{From: old.OwnerID.String(), To: patch.OwnerID.String()}
	}
	if patch.AccountID != nil && !ptrUUIDEq(old.AccountID, patch.AccountID) {
		changes["account_id"] = domain.FieldChange{From: ptrUUIDVal(old.AccountID), To: ptrUUIDVal(patch.AccountID)}
	}
	if patch.ContactID != nil && !ptrUUIDEq(old.ContactID, patch.ContactID) {
		changes["contact_id"] = domain.FieldChange{From: ptrUUIDVal(old.ContactID), To: ptrUUIDVal(patch.ContactID)}
	}
	if len(changes) == 0 {
		return nil
	}
	return changes
}

// buildLeadChanges compares old lead state against a patch.
func buildLeadChanges(old *domain.Lead, patch domain.LeadPatch) domain.AuditChanges {
	changes := domain.AuditChanges{}
	if patch.FirstName != nil && *patch.FirstName != old.FirstName {
		changes["first_name"] = domain.FieldChange{From: old.FirstName, To: *patch.FirstName}
	}
	if patch.LastName != nil && *patch.LastName != old.LastName {
		changes["last_name"] = domain.FieldChange{From: old.LastName, To: *patch.LastName}
	}
	if patch.Email != nil && !ptrStrEq(old.Email, patch.Email) {
		changes["email"] = domain.FieldChange{From: ptrStrVal(old.Email), To: ptrStrVal(patch.Email)}
	}
	if patch.Phone != nil && !ptrStrEq(old.Phone, patch.Phone) {
		changes["phone"] = domain.FieldChange{From: ptrStrVal(old.Phone), To: ptrStrVal(patch.Phone)}
	}
	if patch.Company != nil && !ptrStrEq(old.Company, patch.Company) {
		changes["company"] = domain.FieldChange{From: ptrStrVal(old.Company), To: ptrStrVal(patch.Company)}
	}
	if patch.LeadSource != nil && !ptrStrEq(old.LeadSource, patch.LeadSource) {
		changes["lead_source"] = domain.FieldChange{From: ptrStrVal(old.LeadSource), To: ptrStrVal(patch.LeadSource)}
	}
	if patch.Status != nil && *patch.Status != old.Status {
		changes["status"] = domain.FieldChange{From: string(old.Status), To: string(*patch.Status)}
	}
	if patch.OwnerID != nil && !ptrUUIDEq(old.OwnerID, patch.OwnerID) {
		changes["owner_id"] = domain.FieldChange{From: ptrUUIDVal(old.OwnerID), To: ptrUUIDVal(patch.OwnerID)}
	}
	if len(changes) == 0 {
		return nil
	}
	return changes
}

// buildUserChanges compares old user state against a patch.
func buildUserChanges(old *domain.User, patch domain.UserPatch) domain.AuditChanges {
	changes := domain.AuditChanges{}
	if patch.Name != nil && *patch.Name != old.Name {
		changes["name"] = domain.FieldChange{From: old.Name, To: *patch.Name}
	}
	if patch.Email != nil && *patch.Email != old.Email {
		changes["email"] = domain.FieldChange{From: old.Email, To: *patch.Email}
	}
	if patch.Role != nil && *patch.Role != old.Role {
		changes["role"] = domain.FieldChange{From: string(old.Role), To: string(*patch.Role)}
	}
	if len(changes) == 0 {
		return nil
	}
	return changes
}

// ptrStrEq returns true if both pointer strings are equal.
func ptrStrEq(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// ptrUUIDEq returns true if both UUID pointers are equal.
func ptrUUIDEq(a, b *uuid.UUID) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// ptrStrVal returns nil or the string value for use in AuditChanges.
func ptrStrVal(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

// ptrUUIDVal returns nil or the UUID string for use in AuditChanges.
func ptrUUIDVal(u *uuid.UUID) interface{} {
	if u == nil {
		return nil
	}
	return u.String()
}
