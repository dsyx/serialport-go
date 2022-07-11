# serialport-go

A Go package that allows you to easily access serial ports.

## Usage

`serialport` package provides some common serial port operations: `Open`, `Close`, `Read` and `Write`, etc.

First you should prepare a serial port configuration `Config`:

```go
type Config struct {
	BaudRate int
	DataBits int
	StopBits int
	Parity   int
	Timeout  time.Duration
}
```

If you don't know much about serial port configuration, you can use the default configuration `DefaultConfig`:

```go
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
```

Then, call `Open` to open a serial port:

```go
func Open(name string, cfg Config) (sp *SerialPort, err error) {
	/* Code ... */
}
```

When `Open` is successful, it will return a `*SerialPort`, which you can use to access the serial port.

## Example

A simple serial port echo application:

```go
package main

import (
	"fmt"
	"log"

	"github.com/dsyx/serialport-go"
)

func main() {
	sp, err := serialport.Open("COM1", DefaultConfig())
	if err != nil {
		log.Fatalln(err)
	}
	defer sp.Close()

	buf := make([]byte, 64)
	for {
		n, _ := sp.Read(buf)
		fmt.Printf("read(%d): %s\n", n, string(buf[:n]))
		sp.Write(buf[:n])
	}
}
```
