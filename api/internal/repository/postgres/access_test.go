package postgres

import (
	"testing"

	"github.com/google/uuid"
	"github.com/omnir/crm-api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestSharingRulesForReplaceConvertsLegacyGrants(t *testing.T) {
	sourceID := uuid.New()
	roleID := uuid.New()
	groupID := uuid.New()
	advancedRuleID := uuid.New()
	legacyGrantID := uuid.New()

	rules := sharingRulesForReplace(domain.ACLSharingModuleRule{
		Module: domain.ACLModuleContacts,
		AdvancedRules: []domain.ACLSharingRule{
			{
				ID:          advancedRuleID,
				SourceType:  domain.SharingPrincipalUser,
				SourceID:    &sourceID,
				TargetType:  domain.SharingPrincipalRole,
				TargetID:    roleID,
				AccessLevel: domain.SharingAccessRead,
			},
		},
		Grants: []domain.ACLSharingGrant{
			{
				ID:          legacyGrantID,
				Module:      domain.ACLModuleContacts,
				GranteeType: domain.SharingGranteeRole,
				GranteeID:   roleID,
				AccessLevel: domain.SharingAccessWrite,
			},
			{
				Module:      domain.ACLModuleContacts,
				GranteeType: domain.SharingGranteeGroup,
				GranteeID:   groupID,
				AccessLevel: domain.SharingAccessRead,
			},
		},
	})

	require.Len(t, rules, 3)
	require.Equal(t, advancedRuleID, rules[0].ID)
	require.Equal(t, domain.SharingPrincipalUser, rules[0].SourceType)

	require.Equal(t, uuid.Nil, rules[1].ID)
	require.Equal(t, domain.ACLModuleContacts, rules[1].Module)
	require.Equal(t, domain.SharingPrincipalAll, rules[1].SourceType)
	require.Nil(t, rules[1].SourceID)
	require.Equal(t, domain.SharingPrincipalRole, rules[1].TargetType)
	require.Equal(t, roleID, rules[1].TargetID)
	require.Equal(t, domain.SharingAccessWrite, rules[1].AccessLevel)

	require.Equal(t, uuid.Nil, rules[2].ID)
	require.Equal(t, domain.ACLModuleContacts, rules[2].Module)
	require.Equal(t, domain.SharingPrincipalAll, rules[2].SourceType)
	require.Nil(t, rules[2].SourceID)
	require.Equal(t, domain.SharingPrincipalGroup, rules[2].TargetType)
	require.Equal(t, groupID, rules[2].TargetID)
	require.Equal(t, domain.SharingAccessRead, rules[2].AccessLevel)
}
