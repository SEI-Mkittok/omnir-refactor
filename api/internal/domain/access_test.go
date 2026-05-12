package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccessContextHasPermissionRequiresExplicitExport(t *testing.T) {
	access := &AccessContext{
		Permissions: map[ACLModule]map[ACLAction]bool{
			ACLModuleExport: {
				ACLActionRead: true,
			},
		},
	}

	require.False(t, access.HasPermission(ACLModuleExport, ACLActionExport))

	access.Permissions[ACLModuleExport][ACLActionExport] = true
	require.True(t, access.HasPermission(ACLModuleExport, ACLActionExport))
}

func TestAccessContextHasPermissionAdminStillGrantsExport(t *testing.T) {
	access := &AccessContext{
		Permissions: map[ACLModule]map[ACLAction]bool{
			ACLModuleExport: {
				ACLActionAdmin: true,
			},
		},
	}

	require.True(t, access.HasPermission(ACLModuleExport, ACLActionExport))
}
