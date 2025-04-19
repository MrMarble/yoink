//go:build !windows
// +build !windows

package du

import "syscall"

func Available(path string) uint64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0
	}

	if stat.Bsize < 0 {
		return 0
	}

	return stat.Bavail * uint64(stat.Bsize)
}
