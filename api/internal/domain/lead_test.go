package domain_test

import (
	"testing"

	"github.com/omnir/crm-api/internal/domain"
)

func TestScoreForLeadStatus(t *testing.T) {
	cases := []struct {
		status domain.LeadStatus
		want   int
	}{
		{domain.LeadStatusNew, 0},
		{domain.LeadStatusContacted, 20},
		{domain.LeadStatusQualified, 40},
		{domain.LeadStatusUnqualified, 0},
		{domain.LeadStatusConverted, 100},
		{domain.LeadStatus("unknown"), 0},
	}
	for _, c := range cases {
		got := domain.ScoreForLeadStatus(c.status)
		if got != c.want {
			t.Errorf("ScoreForLeadStatus(%q) = %d, want %d", c.status, got, c.want)
		}
	}
}
