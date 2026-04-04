package domain_test

import (
	"testing"

	"github.com/omnir/crm-api/internal/domain"
)

func TestScoreForStatus(t *testing.T) {
	tests := []struct {
		status domain.LeadStatus
		want   int
	}{
		{domain.LeadStatusNew, 0},
		{domain.LeadStatusContacted, 20},
		{domain.LeadStatusQualified, 40},
		{domain.LeadStatusUnqualified, 0},
		{domain.LeadStatusConverted, 100},
	}
	for _, tt := range tests {
		got := domain.ScoreForStatus(tt.status)
		if got != tt.want {
			t.Errorf("ScoreForStatus(%q) = %d, want %d", tt.status, got, tt.want)
		}
	}
}

func TestSnapScoreToStep(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 0},
		{1, 0},
		{4, 0},
		{5, 10},
		{10, 10},
		{14, 10},
		{15, 20},
		{23, 20},
		{25, 30},
		{50, 50},
		{95, 100},
		{99, 100},
		{100, 100},
		{-5, 0},
		{105, 100},
	}
	for _, tt := range tests {
		got := domain.SnapScoreToStep(tt.input)
		if got != tt.want {
			t.Errorf("SnapScoreToStep(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
