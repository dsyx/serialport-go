package serialport

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

const deciseconds = time.Millisecond * 100 // 1/10 second

// A SerialPort is a serial port. This must be instantiated by calling Open() and not manually.
type SerialPort struct {
	fd int
}

// Open opens a serial port.
func Open(name string, cfg Config) (sp *SerialPort, err error) {
	fd, err := unix.Open(name, unix.O_RDWR|unix.O_NOCTTY, 0666)
	if err != nil {
		return
	}
	sp = &SerialPort{fd: fd}

	if err = sp.SetConfig(cfg); err != nil {
		sp.Close()
	}

	return
}

// Close close the serial port.
func (sp *SerialPort) Close() error {
	return unix.Close(sp.fd)
}

// Read reads up to len(b) bytes from the serial port.
// It returns the number of bytes (0 <= n <= len(b)) read from the serial port and any errors encountered.
// Note:
//     Timeout < 100 ms: Read blocks until at least one byte is readable;
//     Timeout > 100 ms: Read blocks until at least one byte is read or timeout.
func (sp *SerialPort) Read(b []byte) (n int, err error) {
	return unix.Read(sp.fd, b)
}

// Write writes len(b) bytes to the serial port.
// It returns the number of bytes (0 <= n <= len(b)) written to the serial port and any errors encountered.
func (sp *SerialPort) Write(b []byte) (n int, err error) {
	return unix.Write(sp.fd, b)
}

// Flush flushes both data received but not read, and data written but not transmitted.
func (sp *SerialPort) Flush() error {
	return unix.IoctlSetInt(sp.fd, unix.TCFLSH, unix.TCIOFLUSH)
}

// Config returns the configuration of the serial port.
func (sp *SerialPort) Config() (cfg Config, err error) {
	termios, err := unix.IoctlGetTermios(sp.fd, unix.TCGETS2)
	if err != nil {
		return
	}

	cfg.BaudRate = int(termios.Ospeed)

	switch {
	case termios.Cflag&unix.CS5 > 0:
		cfg.DataBits = DB5
	case termios.Cflag&unix.CS6 > 0:
		cfg.DataBits = DB6
	case termios.Cflag&unix.CS7 > 0:
		cfg.DataBits = DB7
	case termios.Cflag&unix.CS8 > 0:
		cfg.DataBits = DB8
	}

	if termios.Cflag&unix.CSTOPB == 0 {
		cfg.StopBits = SB1
	} else {
		cfg.StopBits = SB2
	}

	if termios.Cflag&unix.PARENB == 0 {
		cfg.Parity = PN
	} else if termios.Cflag&unix.PARODD > 0 {
		cfg.Parity = PO
	} else {
		cfg.Parity = PE
	}

	cfg.Timeout = time.Duration(termios.Cc[unix.VTIME]) * deciseconds

	return
}

func checkConfigParam(cfg Config) error {
	if cfg.BaudRate < 0 {
		return fmt.Errorf("serialport: Config.BaudRate cannot be negative %v", cfg.BaudRate)
	}

	if cfg.DataBits != DB5 && cfg.DataBits != DB6 && cfg.DataBits != DB7 && cfg.DataBits != DB8 {
		return fmt.Errorf("serialport: invalid Config.DataBits %v", cfg.DataBits)
	}

	if cfg.StopBits != SB1 && cfg.StopBits != SB2 {
		return fmt.Errorf("serialport: invalid Config.StopBits %v", cfg.StopBits)
	}

	if cfg.Parity != PN && cfg.Parity != PO && cfg.Parity != PE {
		return fmt.Errorf("serialport: invalid Config.Parity %v", cfg.Parity)
	}

	return nil
}

// SetConfig Set the serial port according to Config.
func (sp *SerialPort) SetConfig(cfg Config) error {
	if err := checkConfigParam(cfg); err != nil {
		return err
	}

	termios2 := unix.Termios{}
	termios2.Cflag |= unix.CREAD | unix.CLOCAL | unix.BOTHER

	termios2.Ispeed = uint32(cfg.BaudRate)
	termios2.Ospeed = uint32(cfg.BaudRate)

	// CSIZE  Character size mask.  Values are CS5, CS6, CS7, or CS8.
	switch cfg.DataBits {
	case DB5:
		termios2.Cflag |= unix.CS5
	case DB6:
		termios2.Cflag |= unix.CS6
	case DB7:
		termios2.Cflag |= unix.CS7
	case DB8:
		termios2.Cflag |= unix.CS8
	}

	// CSTOPB Set two stop bits, rather than one.
	switch cfg.StopBits {
	case SB1:
	case SB2:
		termios2.Cflag |= unix.CSTOPB
	}

	// PARENB Enable parity generation on output and parity checking for input.
	// PARODD If set, then parity for input and output is odd; otherwise even parity is used.
	// INPCK  Enable input parity checking.
	switch cfg.Parity {
	case PN:
	case PO:
		termios2.Cflag |= unix.PARENB | unix.PARODD
		termios2.Iflag |= unix.INPCK
	case PE:
		termios2.Cflag |= unix.PARENB
		termios2.Iflag |= unix.INPCK
	}

	// VMIN   Minimum number of characters for noncanonical read (MIN).
	// VTIME  Timeout in t for noncanonical read (TIME).
	t := uint8(cfg.Timeout / deciseconds)
	if t > 0 {
		termios2.Cc[unix.VMIN] = 0
		termios2.Cc[unix.VTIME] = t
	} else {
		termios2.Cc[unix.VMIN] = 1
		termios2.Cc[unix.VTIME] = 0
	}

	return unix.IoctlSetTermios(sp.fd, unix.TCSETS2, &termios2)
}
