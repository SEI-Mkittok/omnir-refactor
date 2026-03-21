package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFromName(t *testing.T) {
	tests := []struct {
		name          string
		rawFrom       string
		email         string
		wantFirstName string
		wantLastName  string
	}{
		{
			name:          "display name with first and last",
			rawFrom:       "Alice Smith <alice@example.com>",
			email:         "alice@example.com",
			wantFirstName: "Alice",
			wantLastName:  "Smith",
		},
		{
			name:          "display name with only first name",
			rawFrom:       "Bob <bob@example.com>",
			email:         "bob@example.com",
			wantFirstName: "Bob",
			wantLastName:  "",
		},
		{
			name:          "no display name — falls back to local part",
			rawFrom:       "carol@example.com",
			email:         "carol@example.com",
			wantFirstName: "carol",
			wantLastName:  "",
		},
		{
			name:          "display name with multiple spaces treated as first + rest",
			rawFrom:       "Dave Van Der Berg <dave@example.com>",
			email:         "dave@example.com",
			wantFirstName: "Dave",
			wantLastName:  "Van Der Berg",
		},
		{
			name:          "invalid RFC 5322 — falls back to local part",
			rawFrom:       "not-an-email",
			email:         "not-an-email",
			wantFirstName: "not-an-email",
			wantLastName:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			first, last := parseFromName(tc.rawFrom, tc.email)
			assert.Equal(t, tc.wantFirstName, first)
			assert.Equal(t, tc.wantLastName, last)
		})
	}
}

func TestParseEmailAddress(t *testing.T) {
	tests := []struct {
		name  string
		from  string
		want  string
	}{
		{
			name: "RFC 5322 format returns bare address",
			from: "Alice Smith <Alice@Example.COM>",
			want: "alice@example.com",
		},
		{
			name: "bare address is lowercased",
			from: "BOB@EXAMPLE.COM",
			want: "bob@example.com",
		},
		{
			name: "plain lowercase address unchanged",
			from: "carol@example.com",
			want: "carol@example.com",
		},
		{
			name: "invalid address returns empty string",
			from: "not-an-email-at-all",
			want: "",
		},
		{
			name: "empty string returns empty string",
			from: "",
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseEmailAddress(tc.from)
			assert.Equal(t, tc.want, got)
		})
	}
}
