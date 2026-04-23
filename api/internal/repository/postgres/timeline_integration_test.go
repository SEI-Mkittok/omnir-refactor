//go:build integration

package postgres_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omnir/crm-api/internal/domain"
	"github.com/omnir/crm-api/internal/repository/postgres"
)

func TestTimelineRepo_List_TotalCountScan(t *testing.T) {
	pool, ctx := setupDB(t)
	ownerID := seedUser(t, pool)

	activityRepo := postgres.NewActivityRepo(pool)
	timelineRepo := postgres.NewTimelineRepo(pool)

	first, err := activityRepo.Create(ctx, &domain.Activity{
		Type:    domain.ActivityTypeCall,
		Subject: "First timeline event",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	second, err := activityRepo.Create(ctx, &domain.Activity{
		Type:    domain.ActivityTypeEmail,
		Subject: "Second timeline event",
		OwnerID: ownerID,
	})
	require.NoError(t, err)

	filter := domain.TimelineFilter{
		Page:  1,
		Limit: 1,
	}

	events, total, err := timelineRepo.List(ctx, filter)
	require.NoError(t, err)
	require.Len(t, events, 1)

	assert.Greater(t, total, 0)
	assert.Equal(t, 2, total)
	assert.Equal(t, "activity.created", events[0].EventType)
	assert.Contains(t, []uuid.UUID{first.ID, second.ID}, events[0].EventID)
}
