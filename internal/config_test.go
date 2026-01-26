package crawler

import "testing"

func Test_readConfig(t *testing.T) {
	tests := []struct {
		name    string // description of this test case
		want    string
		wantErr bool
	}{
		{
			name:    "test read douban movie config",
			want:    "#nowplaying .list-item",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := readConfig()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("readConfig() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("readConfig() succeeded unexpectedly")
			}
			gotSelector := got.Rules["douban-movie"].Spiders["Home"][0].Selector
			t.Fatalf("%v", got.Rules["douban-movie"].Spiders)
			if gotSelector != tt.want {
				t.Errorf("readConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
