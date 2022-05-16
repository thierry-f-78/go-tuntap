// +build linux

package tuntap

import (
	"os"
	"unsafe"
	"golang.org/x/sys/unix"
)

const (
	cIFF_TUN   = 0x0001
	cIFF_TAP   = 0x0002
	cIFF_NO_PI = 0x1000
)

type device struct {
	n string
	f *os.File
}

func (d *device) Name() string                { return d.n }
func (d *device) String() string              { return d.n }
func (d *device) Close() error                { return d.f.Close() }
func (d *device) Read(p []byte) (int, error)  { return d.f.Read(p) }
func (d *device) Write(p []byte) (int, error) { return d.f.Write(p) }

func newTUN(name string) (Interface, error) {
	file, err := createTuntapInterface(name, unix.IFF_TUN|unix.IFF_NO_PI)
	if err != nil {
		return nil, err
	}

	return &device{n: name, f: file}, nil
}

func newTAP(name string) (Interface, error) {
	file, err := createTuntapInterface(name, unix.IFF_TAP|unix.IFF_NO_PI)
	if err != nil {
		return nil, err
	}

	return &device{n: name, f: file}, nil
}

func createTuntapInterface(name string, flags uint16) (*os.File, error) {

	tunfd, err := unix.Open("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	var ifr [unix.IFNAMSIZ + 64]byte
	copy(ifr[:], []byte(name))
	*(*uint16)(unsafe.Pointer(&ifr[unix.IFNAMSIZ])) = flags
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(tunfd),
		uintptr(unix.TUNSETIFF),
		uintptr(unsafe.Pointer(&ifr[0])),
	)

	if errno != 0 {
		return nil, errno
	}
	unix.SetNonblock(tunfd, true)

	fd := os.NewFile(uintptr(tunfd), "/dev/net/tun")

	return fd, nil
}
