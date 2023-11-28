package processes

import (
	"testing"

	"karma8/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestSplitFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		partsIDs []int64
		want     []models.BucketItem
		wantErr  string
	}{
		{
			name:     "Good file split",
			filePath: "Checksum.csv",
			partsIDs: []int64{1, 2, 3},
			want: []models.BucketItem{
				{ID: 1, Source: []byte(`ip_address,country_code,country,city,latitude,longitude,mystery_value
200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
160.103.7.140,CZ,Nicaragua,New Neva,`)},
				{ID: 2, Source: []byte(`-68.31023296602508,-37.62435199624531,7301823115
70.95.73.73,TL,Saudi Arabia,Gradymouth,-49.16675918861615,-86.05920084416894,2559997162
,PY,Falkland Islands (Malvinas),,75.41685191518815,-14`)},
				{ID: 3, Source: []byte(`4.6943217219469,0
125.159.20.54,LI,Guyana,Port Karson,-78.2274228596799,-163.26218895343357,1337885276
not your IP address,HN,Benin,Fredyshire,-70.41275040993187,60.19866111663936,2040256925
`)}},
			wantErr: "",
		},
		{
			name:     "Empty file",
			filePath: "empty_file.txt",
			partsIDs: []int64{1, 2, 3},
			want:     nil,
			wantErr:  "file is empty",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := readFixture(t, tt.filePath)
			got, err := SplitFile(path, tt.partsIDs)
			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Fatal("SplitFile() mismatch\n", diff)
			}
		})
	}
}

func TestMergeFile(t *testing.T) {
	tests := []struct {
		name        string
		bucketItems []models.BucketItem
		want        []byte
	}{
		{
			name: "Good merge",
			bucketItems: []models.BucketItem{
				{ID: 2, Source: []byte(`-68.31023296602508,-37.62435199624531,7301823115
70.95.73.73,TL,Saudi Arabia,Gradymouth,-49.16675918861615,-86.05920084416894,2559997162
,PY,Falkland Islands (Malvinas),,75.41685191518815,-14`)},
				{ID: 1, Source: []byte(`ip_address,country_code,country,city,latitude,longitude,mystery_value
200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
160.103.7.140,CZ,Nicaragua,New Neva,`)},
				{ID: 3, Source: []byte(`4.6943217219469,0
125.159.20.54,LI,Guyana,Port Karson,-78.2274228596799,-163.26218895343357,1337885276
not your IP address,HN,Benin,Fredyshire,-70.41275040993187,60.19866111663936,2040256925
`)}},
			want: []byte(`ip_address,country_code,country,city,latitude,longitude,mystery_value
200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
160.103.7.140,CZ,Nicaragua,New Neva,-68.31023296602508,-37.62435199624531,7301823115
70.95.73.73,TL,Saudi Arabia,Gradymouth,-49.16675918861615,-86.05920084416894,2559997162
,PY,Falkland Islands (Malvinas),,75.41685191518815,-144.6943217219469,0
125.159.20.54,LI,Guyana,Port Karson,-78.2274228596799,-163.26218895343357,1337885276
not your IP address,HN,Benin,Fredyshire,-70.41275040993187,60.19866111663936,2040256925
`),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MergeFile(tt.bucketItems)
			diff := cmp.Diff(tt.want, got)
			if diff != "" {
				t.Fatal("MergeFile() mismatch\n", diff)
			}
		})
	}
}
