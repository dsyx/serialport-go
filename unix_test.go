// +build linux

package serialport

import (
	"fmt"
	"testing"
)

func TestEcho(*testing.T) {
	//config := &Config{
	//	Name:     "/dev/ttyS0",
	//	Baudrate: B115200,
	//	DataBits: 8,
	//	StopBits: 1,
	//	Parity:   NoParity,
	//	Timeout:  100,
	//}
	//port, _ := OpenByConfig(config)
	port, _ := Open("/dev/ttyS0")
	buf := make([]byte, 256)
	for {
		n, _ := port.Read(buf)
		//port.Flush()
		fmt.Printf("read(%d): %s\n", n, string(buf[:n]))
		port.Write(buf[:n])
	}
}
