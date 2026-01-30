package crawler

import (
	"testing"
)

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
			ReadConfig()
			gotSelector := Conf.Spiders["douban-movie"].Rules["home"][0].Selector
			if gotSelector != tt.want {
				t.Errorf("readConfig() = %v, want %v", gotSelector, tt.want)
			}
		})
	}
}
