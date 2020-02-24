// +build linux

package serialport

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Port represents an open serial port.
type Port struct {
	file *os.File
}

// Read reads up to len(b) bytes from the Port.
// It returns the number of bytes read and any error encountered.
// At end of file, Read returns 0, io.EOF.
func (p *Port) Read(b []byte) (n int, err error) {
	return p.file.Read(b)
}

// Write writes len(b) bytes to the Port.
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (p *Port) Write(b []byte) (n int, err error) {
	return p.file.Write(b)
}

// Close closes the Port, and makes it unusable.
// It return an error if it has already been called.
func (p *Port) Close() error {
	return p.file.Close()
}

// Flush flushes both data received but not read, and data written but not transmitted.
func (p *Port) Flush() error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(p.file.Fd()),
		uintptr(unix.TCFLSH),
		uintptr(unix.TCIOFLUSH),
	)
	return errno
}

func open(cfg *Config) (p *Port, err error) {
	fd, err := unix.Open(cfg.Name, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		err = &OpenError{"Name", err.Error()}
		return
	}
	// enable receiver, local line - do not change "owner" of port
	var cFlag uint32 = unix.CREAD | unix.CLOCAL
	// baudrate
	var baudrate uint32
	switch cfg.Baudrate {
	case B0:
		baudrate = unix.B0
	case B300:
		baudrate = unix.B300
	case B600:
		baudrate = unix.B600
	case B1200:
		baudrate = unix.B1200
	case B2400:
		baudrate = unix.B2400
	case B4800:
		baudrate = unix.B4800
	case B9600:
		baudrate = unix.B9600
	case B19200:
		baudrate = unix.B19200
	case B38400:
		baudrate = unix.B38400
	case B57600:
		baudrate = unix.B57600
	case B115200:
		baudrate = unix.B115200
	default:
		err = &OpenError{"Baudrate", fmt.Sprintf("not support %d", cfg.Baudrate)}
		return
	}
	cFlag |= baudrate
	// data bits
	switch cfg.DataBits {
	case 6:
		cFlag |= unix.CS6
	case 7:
		cFlag |= unix.CS7
	case 8:
		cFlag |= unix.CS8
	default:
		err = &OpenError{"DataBits", fmt.Sprintf("not support %d", cfg.DataBits)}
		return
	}
	// stop bits
	switch cfg.StopBits {
	case 1:
		// default
	case 2:
		cFlag |= unix.CSTOPB
	default:
		err = &OpenError{"StopBits", fmt.Sprintf("not support %d", cfg.StopBits)}
		return
	}
	// parity
	var iFlag uint32
	switch cfg.Parity {
	case NoParity:
		// default
	case OddParity:
		cFlag |= unix.PARENB
		cFlag |= unix.PARODD
		iFlag |= unix.INPCK
	case EvenParity:
		cFlag |= unix.PARENB
		iFlag |= unix.INPCK
	default:
		err = &OpenError{"Parity", fmt.Sprintf("not support %d", cfg.Parity)}
		return
	}
	// termios
	termios := unix.Termios{
		Cflag:  cFlag,
		Iflag:  iFlag,
		Ispeed: baudrate,
		Ospeed: baudrate,
	}
	// timeout [100ms ~ 25500ms]
	if cfg.Timeout >= 100 {
		if cfg.Timeout > 25500 {
			cfg.Timeout = 25500
		}
		termios.Cc[unix.VMIN] = 1
		termios.Cc[unix.VTIME] = uint8(cfg.Timeout / 100)
	}
	// ioctl
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(unix.TCSETS),
		uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		err = &OpenError{"IOCTL", fmt.Sprintf("errno %v", errno)}
		return
	}
	// NewFile
	file := os.NewFile(uintptr(fd), cfg.Name)
	if file == nil {
		err = &OpenError{"NewFile", "fd is not a valid file descriptor"}
		return
	}
	return &Port{file: file}, nil
}
