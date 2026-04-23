package handler

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository"
)

func validateContactAccountPair(ctx context.Context, contacts repository.ContactRepository, contactID, accountID *uuid.UUID) error {
	if contacts == nil || contactID == nil || accountID == nil {
		return nil
	}

	related, err := contacts.IsRelatedToAccount(ctx, *contactID, *accountID)
	if err != nil {
		return err
	}
	if related {
		return nil
	}
	return fmt.Errorf("%w: contact_id is not related to account_id", domain.ErrValidation)
}

func validateContactDealPair(ctx context.Context, contacts repository.ContactRepository, deals repository.DealRepository, contactID, dealID *uuid.UUID) error {
	if contacts == nil || deals == nil || contactID == nil || dealID == nil {
		return nil
	}

	deal, err := deals.GetByID(ctx, *dealID)
	if err != nil {
		return err
	}
	if deal.AccountID == nil {
		if deal.ContactID != nil && *deal.ContactID == *contactID {
			return nil
		}
		for _, c := range deal.Contacts {
			if c.ID == *contactID {
				return nil
			}
		}
		return fmt.Errorf("%w: contact_id is not related to deal_id", domain.ErrValidation)
	}

	related, err := contacts.IsRelatedToAccount(ctx, *contactID, *deal.AccountID)
	if err != nil {
		return err
	}
	if !related {
		return fmt.Errorf("%w: contact_id is not related to deal_id account", domain.ErrValidation)
	}
	return nil
}
