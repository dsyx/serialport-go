// Package serialport provides some basic access to the serial port.
package serialport

import "fmt"

// Baudrate
const (
	B0      = 0
	B300    = 300
	B600    = 600
	B1200   = 1200
	B2400   = 2400
	B4800   = 4800
	B9600   = 9600
	B19200  = 19200
	B38400  = 38400
	B57600  = 57600
	B115200 = 115200
)

// Parity
const (
	NoParity = iota
	OddParity
	EvenParity
)

// Config describes the configuration of the serial port.
// Name field is the serial port name or its path.
// Baudrate field is the baudrate. Only Baudrate constant values are supported.
// DataBits field is the data bit size. Only 6, 7, 8 bits are supported.
// StopBits field is the stop bit size. Only 1, 2 bits are supported.
// Parity field is parity mode. Only Parity constant values are supported.
// Timeout field is the wait timeout (ms). Read will return immediately (<= 0), or wait until timeout (> 0) after receiving a byte.
type Config struct {
	Name     string
	Baudrate int
	DataBits int
	StopBits int
	Parity   int
	Timeout  int
}

// OpenError indicates an error that occurred while opening the serial port.
type OpenError struct {
	param string
	cause string
}

func (e *OpenError) Error() string {
	return fmt.Sprintf("[%s] %s", e.param, e.cause)
}

// Index of Open args.
const (
	baudrateIndex = iota
	timeoutIndex
	parityIndex
	dataBitsIndex
	stopBitsIndex
)

// Open for open a serial port according to the specified name and args.
// args[0]: Baudrate, default B9600
// args[1]: Timeout, default 100
// args[2]: Parity, default NoParity
// args[3]: DataBits, default 8
// args[4]: StopBits, default 1
func Open(name string, args ...int) (p *Port, err error) {
	cfg := &Config{
		Name:     name,
		Baudrate: B9600,
		Timeout:  100,
		Parity:   NoParity,
		DataBits: 8,
		StopBits: 1,
	}
	for i := 0; i < len(args); i++ {
		switch i {
		case baudrateIndex:
			cfg.Baudrate = args[i]
		case timeoutIndex:
			cfg.Timeout = args[i]
		case parityIndex:
			cfg.Parity = args[i]
		case dataBitsIndex:
			cfg.DataBits = args[i]
		case stopBitsIndex:
			cfg.StopBits = args[i]
		}
	}
	return open(cfg)
}

// OpenByConfig for open a serial port according to the specified Config.
func OpenByConfig(cfg *Config) (p *Port, err error) {
	return open(cfg)
}
