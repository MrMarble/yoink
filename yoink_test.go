// Package yoink provides utilities to manage freeleech
// downloads automatically
package yoink

import (
	"reflect"
	"testing"

	"github.com/mrmarble/yoink/pkg/prowlarr"
)

func Test_filterTorrentsByDiskSize(t *testing.T) {
	type args struct {
		config    *Config
		totalSize uint64
		torrents  []prowlarr.SearchResult
	}
	tests := []struct {
		name string
		args args
		want []prowlarr.SearchResult
	}{
		{"", args{&Config{TotalFreelechSize: 5}, 100, []prowlarr.SearchResult{{Size: 100}}}, []prowlarr.SearchResult{{Size: 100}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterTorrentsByDiskSize(tt.args.config, tt.args.totalSize, tt.args.torrents); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterTorrentsByDiskSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
