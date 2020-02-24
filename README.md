# serialport-go

A simple serial port package written in Go.

Different platforms have different ways to access the serial port. This package depends on the compiler environment. Therefore, if you need to compile across platforms, modify `GOOS` and `GOARCH`.

# Usage

`serialport` package provides two functions for opening a serial port:

```
// Open for open a serial port according to the specified name and args.
// args[0]: Baudrate, default B9600
// args[1]: Timeout, default 100
// args[2]: Parity, default NoParity
// args[3]: DataBits, default 8
// args[4]: StopBits, default 1
func Open(name string, args ...int) (p *Port, err error)

// OpenByConfig for open a serial port according to the specified Config.
func OpenByConfig(cfg *Config) (p *Port, err error)
```

After the function succeeds, it will return a `Port` pointer, and then you can use the `Port` method set to access the serial port.

# Example

A simple serial echo:

```go
package main

import (
	"fmt"

	"github.com/dsyx/serialport-go"
)

func main() {
	port, _ := serialport.Open("/dev/ttyS0")
	buf := make([]byte, 256)
	for {
		n, _ := port.Read(buf)
		fmt.Printf("read(%d): %s\n", n, string(buf[:n]))
		port.Write(buf[:n])
	}
}
```
