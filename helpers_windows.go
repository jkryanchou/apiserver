// +build windows

package main

import (
	"net"
	"os"
	"syscall"
)

func SetStdHandle(stdhandle int32, handle syscall.Handle) error {
	procSetStdHandle := syscall.MustLoadDLL("kernel32.dll").MustFindProc("SetStdHandle")
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}

func RedirectStderrTo(file *os.File) error {
	err := SetStdHandle(syscall.STD_ERROR_HANDLE, syscall.Handle(file.Fd()))
	if err != nil {
		return err
	}

	os.Stderr = file

	return nil
}

func SetBindNoPortSockopts(c syscall.RawConn) error {
	return nil
}

func ReusePortListen(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

func ReusePortListenUDP(network string, laddr *net.UDPAddr) (*net.UDPConn, error) {
	return net.ListenUDP(network, laddr)
}

func SetProcessName(name string) error {
	return nil
}
