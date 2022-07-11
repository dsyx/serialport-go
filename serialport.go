// Package serialport allows you to easily access serial ports
package serialport

import "time"

// Config for serial port configuration:
//     BaudRate is the baud rate of serial transmission
//     DataBits is the number of bits per character
//     StopBits is the number of stop bits
//     Parity is a method of detecting errors in transmission
//     Timeout is the serial port Read() timeout
type Config struct {
	BaudRate int
	DataBits int
	StopBits int
	Parity   int
	Timeout  time.Duration
}

// BaudRate
const (
	BR110    = 110    // 110 bps
	BR300    = 300    // 300 bps
	BR600    = 600    // 600 bps
	BR1200   = 1200   // 1200 bps
	BR2400   = 2400   // 2400 bps
	BR4800   = 4800   // 4800 bps
	BR9600   = 9600   // 9600 bps
	BR14400  = 14400  // 14400 bps
	BR19200  = 19200  // 19200 bps
	BR38400  = 38400  // 38400 bps
	BR57600  = 57600  // 57600 bps
	BR115200 = 115200 // 115200 bps
	BR128000 = 128000 // 128000 bps
	BR256000 = 256000 // 256000 bps
)

// DataBits
const (
	DB5 = 5 // 5 data bits
	DB6 = 6 // 6 data bits
	DB7 = 7 // 7 data bits
	DB8 = 8 // 8 data bits
)

// StopBits
const (
	SB1   = 1  // 1 stop bit
	SB1_5 = 15 // 1.5 stop bits
	SB2   = 2  // 2 stop bits
)

// Parity
const (
	PN = 0 // No parity
	PO = 1 // Odd parity
	PE = 2 // Even parity
	PM = 3 // Mark parity
	PS = 4 // Space parity
)

// DefaultConfig returns a default serial port configuration:
//     115200 bps baudrate
//     8 data bits
//     1 stop bit
//     no parity
//     100 ms timeout
func DefaultConfig() Config {
	return Config{
		BaudRate: BR115200,
		DataBits: DB8,
		StopBits: SB1,
		Parity:   PN,
		Timeout:  100 * time.Millisecond,
	}
}
