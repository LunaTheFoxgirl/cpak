package libcpak

import "syscall"

func (pk Package) Install() {
	sep := PATH_SEP
	root := PATH_ROOT

	syscall.Chroot("")
}
