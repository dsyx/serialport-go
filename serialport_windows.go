package serialport

import (
	"fmt"
	"math"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Reference https://docs.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-dcb:
// typedef struct _DCB {
//   DWORD DCBlength;
//   DWORD BaudRate;
//   DWORD fBinary : 1;
//   DWORD fParity : 1;
//   DWORD fOutxCtsFlow : 1;
//   DWORD fOutxDsrFlow : 1;
//   DWORD fDtrControl : 2;
//   DWORD fDsrSensitivity : 1;
//   DWORD fTXContinueOnXoff : 1;
//   DWORD fOutX : 1;
//   DWORD fInX : 1;
//   DWORD fErrorChar : 1;
//   DWORD fNull : 1;
//   DWORD fRtsControl : 2;
//   DWORD fAbortOnError : 1;
//   DWORD fDummy2 : 17;
//   WORD  wReserved;
//   WORD  XonLim;
//   WORD  XoffLim;
//   BYTE  ByteSize;
//   BYTE  Parity;
//   BYTE  StopBits;
//   char  XonChar;
//   char  XoffChar;
//   char  ErrorChar;
//   char  EofChar;
//   char  EvtChar;
//   WORD  wReserved1;
// } DCB, *LPDCB;
//
// But Go does not support bit field.
type win32DCB struct {
	DCBlength  uint32
	BaudRate   uint32
	fxxxxBits  uint32
	wReserved  uint16
	XonLim     uint16
	XoffLim    uint16
	ByteSize   uint8
	Parity     uint8
	StopBits   uint8
	XonChar    int8
	XoffChar   int8
	ErrorChar  int8
	EofChar    int8
	EvtChar    int8
	wReserved1 uint16
}

const (
	win32ONESTOPBIT   = 0
	win32ONE5STOPBITS = 1
	win32TWOSTOPBITS  = 2
)

var (
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetCommState = modkernel32.NewProc("GetCommState")
	procSetCommState = modkernel32.NewProc("SetCommState")
)

// serialport stopbits to win32 stopbits
var spToWinStopBitsMap = map[int]uint8{
	SB1:   win32ONESTOPBIT,
	SB1_5: win32ONE5STOPBITS,
	SB2:   win32TWOSTOPBITS,
}

// win32 stopbits to serialport stopbits
var winToSpStopBitsMap = map[uint8]int{
	win32ONESTOPBIT:   SB1,
	win32ONE5STOPBITS: SB1_5,
	win32TWOSTOPBITS:  SB2,
}

func win32GetCommState(handle windows.Handle, dcb *win32DCB) error {
	r1, _, err := syscall.Syscall(procGetCommState.Addr(), 2, uintptr(handle), uintptr(unsafe.Pointer(dcb)), 0)
	if r1 == 0 {
		return err
	}
	return nil
}

func win32SetCommState(handle windows.Handle, dcb *win32DCB) error {
	r1, _, err := syscall.Syscall(procSetCommState.Addr(), 2, uintptr(handle), uintptr(unsafe.Pointer(dcb)), 0)
	if r1 == 0 {
		return err
	}
	return nil
}

// A SerialPort is a serial port. This must be instantiated by calling Open() and not manually.
type SerialPort struct {
	handle windows.Handle
}

// Open opens a serial port.
func Open(name string, cfg Config) (sp *SerialPort, err error) {
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(name),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		0,
		0)
	if err != nil {
		return
	}
	sp = &SerialPort{handle: handle}

	if err = sp.SetConfig(cfg); err != nil {
		sp.Close()
	}

	return
}

// Close close the serial port.
func (sp *SerialPort) Close() error {
	return windows.CloseHandle(sp.handle)
}

// Read reads up to len(b) bytes from the serial port.
// It returns the number of bytes (0 <= n <= len(b)) read from the serial port and any errors encountered.
// Note:
//     Timeout < 1 ms: Read blocks until len(b) bytes are readable;
//     Timeout > 1 ms: Read blocks until at least one byte is read or timeout.
func (sp *SerialPort) Read(b []byte) (n int, err error) {
	return windows.Read(sp.handle, b)
}

// Write writes len(b) bytes to the serial port.
// It returns the number of bytes (0 <= n <= len(b)) written to the serial port and any errors encountered.
func (sp *SerialPort) Write(b []byte) (n int, err error) {
	return windows.Write(sp.handle, b)
}

// Config returns the configuration of the serial port.
func (sp *SerialPort) Config() (cfg Config, err error) {
	dcb := win32DCB{DCBlength: uint32(unsafe.Sizeof(win32DCB{}))}
	if err = win32GetCommState(sp.handle, &dcb); err != nil {
		return
	}
	timeouts := windows.CommTimeouts{}
	if err = windows.GetCommTimeouts(sp.handle, &timeouts); err != nil {
		return
	}

	cfg = Config{
		BaudRate: int(dcb.BaudRate),
		DataBits: int(dcb.ByteSize),
		StopBits: winToSpStopBitsMap[dcb.StopBits],
		Parity:   int(dcb.Parity),
		Timeout:  time.Duration(timeouts.ReadTotalTimeoutConstant) * time.Millisecond,
	}

	return
}

func checkConfigParam(cfg Config) error {
	if cfg.BaudRate < 0 {
		return fmt.Errorf("serialport: Config.BaudRate cannot be negative %v", cfg.BaudRate)
	}

	if cfg.DataBits != DB5 && cfg.DataBits != DB6 && cfg.DataBits != DB7 && cfg.DataBits != DB8 {
		return fmt.Errorf("serialport: invalid Config.DataBits %v", cfg.DataBits)
	}

	if cfg.StopBits != SB1 && cfg.StopBits != SB1_5 && cfg.StopBits != SB2 {
		return fmt.Errorf("serialport: invalid Config.StopBits %v", cfg.StopBits)
	}

	if cfg.Parity != PN && cfg.Parity != PO && cfg.Parity != PE && cfg.Parity != PM && cfg.Parity != PS {
		return fmt.Errorf("serialport: invalid Config.Parity %v", cfg.Parity)
	}

	return nil
}

// SetConfig Set the serial port according to Config.
func (sp *SerialPort) SetConfig(cfg Config) error {
	if err := checkConfigParam(cfg); err != nil {
		return err
	}

	dcb := win32DCB{
		DCBlength: uint32(unsafe.Sizeof(win32DCB{})),
		BaudRate:  uint32(cfg.BaudRate),
		ByteSize:  uint8(cfg.DataBits),
		Parity:    uint8(cfg.Parity),
		StopBits:  spToWinStopBitsMap[cfg.StopBits],
	}
	if err := win32SetCommState(sp.handle, &dcb); err != nil {
		return err
	}

	var commTimeouts windows.CommTimeouts
	timeoutMs := uint32(cfg.Timeout.Milliseconds())
	if timeoutMs > 0 {
		commTimeouts = windows.CommTimeouts{
			ReadIntervalTimeout:        math.MaxUint32,
			ReadTotalTimeoutMultiplier: math.MaxUint32,
			ReadTotalTimeoutConstant:   timeoutMs,
			WriteTotalTimeoutConstant:  timeoutMs,
		}
	} else {
		commTimeouts = windows.CommTimeouts{}
	}
	if err := windows.SetCommTimeouts(sp.handle, &commTimeouts); err != nil {
		return err
	}

	return nil
}
