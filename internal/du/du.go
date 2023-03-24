//go:build !windows
// +build !windows

package du

import "syscall"

func Available(path string) uint64 {
	var stat syscall.Statfs_t
	syscall.Statfs(path, &stat) //nolint:errcheck

	return stat.Bavail * uint64(stat.Bsize)
}
