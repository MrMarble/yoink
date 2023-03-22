package du

import (
	"syscall"
	"unsafe"
)

func Available(path string) uint64 {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	var available int64

	utfPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0
	}

	c.Call(
		uintptr(unsafe.Pointer(utfPtr)),
		uintptr(unsafe.Pointer(&available)))

	return uint64(available)
}
