package processes

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
		wantDiff bool
	}{
		{
			name:     "Check valid file Text",
			filePath: "Checksum.csv",
			want:     "75ea73570e0b8b7558304d292594017afa3ff4deef02e1dad40e8bb81863ac14",
			wantDiff: false,
		},
		{
			name:     "Check valid file PNG",
			filePath: "small_down_arrow_icon_test.png",
			want:     "c516da9a5ac7e341975b908821c24e5b0ce19304916a3b18d4ceccd1db642ad5",
			wantDiff: false,
		},
		{
			name:     "Check not valid changed file Text",
			filePath: "changed_Checksum.csv",
			want:     "75ea73570e0b8b7558304d292594017afa3ff4deef02e1dad40e8bb81863ac14",
			wantDiff: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := readFixture(t, tt.filePath)
			got, err := CalculateChecksum(path)
			assert.NoError(t, err)
			if tt.wantDiff {
				assert.NotEqual(t, tt.want, got)
			} else {
				diff := cmp.Diff(tt.want, got)
				if diff != "" {
					t.Fatal("CalculateChecksum() mismatch\n", diff)
				}
			}
		})
	}
}

func readFixture(t *testing.T, name string) string {
	t.Helper()

	_, curFile, _, ok := runtime.Caller(0)
	require.True(t, ok)

	path := filepath.Join(filepath.Dir(curFile), "testdata", name)

	return path
}
