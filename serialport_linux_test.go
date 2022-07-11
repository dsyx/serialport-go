package serialport

import (
	"testing"
	"time"
)

func TestHelloWorld(t *testing.T) {
	sp, err := Open("/dev/pts/3", DefaultConfig())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer sp.Close()

	n, err := sp.Write([]byte("Hello, World"))
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	t.Logf("Write %v bytes to serial port", n)
}

func TestEcho(t *testing.T) {
	cfg := DefaultConfig()
	//cfg.Timeout = 0
	cfg.Timeout = 1000 * time.Millisecond
	sp, err := Open("/dev/pts/3", cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer sp.Close()

	buf := make([]byte, 64)
	for {
		n, err := sp.Read(buf)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		t.Logf("Read(%v): %v", n, buf[:n])

		n, err = sp.Write(buf[:n])
		if err != nil {
			t.Fatalf("Write: %v", err)
		}
		t.Logf("Write(%v): %v", n, buf[:n])
	}
}

func TestFlush(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Timeout = 12 * time.Second
	sp, err := Open("/dev/pts/3", cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer sp.Close()

	buf := make([]byte, 2)
	n, err := sp.Read(buf)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	t.Logf("Read %v bytes: %v", n, string(buf[:n]))

	sp.Flush()
	for {
		n, err = sp.Read(buf)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if n == 0 {
			break
		}
		t.Logf("Read %v bytes: %v", n, string(buf[:n]))
	}
}
