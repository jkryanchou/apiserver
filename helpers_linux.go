// +build linux

package main

import (
	"net"
	"os"
	"reflect"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func RedirectStderrTo(file *os.File) error {
	return syscall.Dup3(int(file.Fd()), 2, 0)
}

func SetBindNoPortSockopts(c syscall.RawConn) error {
	const IP_BIND_ADDRESS_NO_PORT = 24
	return c.Control(func(fd uintptr) {
		syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, IP_BIND_ADDRESS_NO_PORT, 1)
	})
}

func ReusePortListen(network, address string) (net.Listener, error) {
	const SO_REUSEPORT = 15
	control := func(network string, address net.Addr, conn syscall.RawConn) error {
		return conn.Control(func(fd uintptr) {
			syscall.SetsockoptInt(int(fd), unix.SOL_SOCKET, SO_REUSEPORT, 1)
		})
	}
	return net.ListenControl(network, address, control)
}

func ReusePortListenUDP(network string, laddr *net.UDPAddr) (*net.UDPConn, error) {
	const SO_REUSEPORT = 15
	control := func(network string, address net.Addr, conn syscall.RawConn) error {
		return conn.Control(func(fd uintptr) {
			syscall.SetsockoptInt(int(fd), unix.SOL_SOCKET, SO_REUSEPORT, 1)
		})
	}
	return net.ListenUDPControl(network, laddr, control)
}

func SetProcessName(name string) error {
	argv0str := (*reflect.StringHeader)(unsafe.Pointer(&os.Args[0]))
	argv0 := (*[1 << 30]byte)(unsafe.Pointer(argv0str.Data))[:len(name)+1]

	n := copy(argv0, name+string(0))
	if n < len(argv0) {
		argv0[n] = 0
	}
	return nil
}
