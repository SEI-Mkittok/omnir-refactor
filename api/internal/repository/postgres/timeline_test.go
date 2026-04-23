package postgres

import "testing"

func TestTimelineTotalFromInt64(t *testing.T) {
	t.Parallel()

	maxInt := int64(^uint(0) >> 1)

	tests := []struct {
		name    string
		input   int64
		want    int
		wantErr bool
	}{
		{
			name:  "negative totals clamp to zero",
			input: -1,
			want:  0,
		},
		{
			name:  "zero stays zero",
			input: 0,
			want:  0,
		},
		{
			name:  "positive totals within int range convert safely",
			input: 42,
			want:  42,
		},
		{
			name:    "totals above int range return error",
			input:   maxInt + 1,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := timelineTotalFromInt64(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %d", tc.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}
