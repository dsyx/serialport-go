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

var (
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procGetCommState = modkernel32.NewProc("GetCommState")
	procSetCommState = modkernel32.NewProc("SetCommState")
)

var dcbByteSizeMap = map[int]uint8{
	DB5: 5,
	DB6: 6,
	DB7: 7,
	DB8: 8,
}

var dcbStopBitsMap = map[int]uint8{
	SB1:   0,
	SB1_5: 1,
	SB2:   2,
}

var dcbParityMap = map[int]uint8{
	PN: 0,
	PO: 1,
	PE: 2,
	PM: 3,
	PS: 4,
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

func open(name string) (sp *SerialPort, err error) {
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(name),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		0,
		0)
	if err != nil {
		windows.CloseHandle(handle)
	}
	sp = &SerialPort{handle: handle}
	return
}

// Open opens a serial port.
func Open(name string, cfg Config) (sp *SerialPort, err error) {
	sp, err = open(name)
	if err != nil {
		return
	}
	err = sp.SetConfig(cfg)
	return
}

// Close close the serial port.
func (sp *SerialPort) Close() error {
	return windows.CloseHandle(sp.handle)
}

// Read reads up to len(b) bytes from the serial port.
// It returns the number of bytes (0 <= n <= len(b)) read from the serial port and any errors encountered.
func (sp *SerialPort) Read(b []byte) (n int, err error) {
	var done uint32
	err = windows.ReadFile(sp.handle, b, &done, nil)
	n = int(done)
	return
}

// Write writes len(b) bytes to the serial port.
// It returns the number of bytes (0 <= n <= len(b)) written to the serial port and any errors encountered.
func (sp *SerialPort) Write(b []byte) (n int, err error) {
	var done uint32
	err = windows.WriteFile(sp.handle, b, &done, nil)
	n = int(done)
	return
}

func findMapKey(m map[int]uint8, value uint8) int {
	for k, v := range m {
		if v == value {
			return k
		}
	}
	return int(value)
}

// Config returns the configuration of the serial port.
func (sp *SerialPort) Config() (Config, error) {
	dcb := win32DCB{DCBlength: uint32(unsafe.Sizeof(win32DCB{}))}
	if err := win32GetCommState(sp.handle, &dcb); err != nil {
		return Config{}, err
	}
	timeouts := windows.CommTimeouts{}
	if err := windows.GetCommTimeouts(sp.handle, &timeouts); err != nil {
		return Config{}, err
	}

	baudrate := int(dcb.BaudRate)
	databits := findMapKey(dcbByteSizeMap, dcb.ByteSize)
	stopbits := findMapKey(dcbStopBitsMap, dcb.StopBits)
	parity := findMapKey(dcbParityMap, dcb.Parity)
	timeout := time.Duration(timeouts.ReadTotalTimeoutConstant) * time.Millisecond
	return Config{
		BaudRate: baudrate,
		DataBits: databits,
		StopBits: stopbits,
		Parity:   parity,
		Timeout:  timeout,
	}, nil
}

func checkConfigParam(cfg Config) error {
	if cfg.BaudRate < 0 {
		return fmt.Errorf("serialport: Config.BaudRate cannot be negative %v", cfg.BaudRate)
	}
	if _, ok := dcbByteSizeMap[cfg.DataBits]; !ok {
		return fmt.Errorf("serialport: invalid Config.DataBits %v", cfg.DataBits)
	}
	if _, ok := dcbStopBitsMap[cfg.StopBits]; !ok {
		return fmt.Errorf("serialport: invalid Config.StopBits %v", cfg.StopBits)
	}
	if _, ok := dcbParityMap[cfg.Parity]; !ok {
		return fmt.Errorf("serialport: invalid Config.Parity %v", cfg.Parity)
	}
	if cfg.Timeout != 0 && cfg.Timeout.Milliseconds() == 0 {
		return fmt.Errorf("serialport: Config.Timeout on windows in milliseconds %v", cfg.Timeout)
	}

	return nil
}

// SetConfig Set the serial port according to Config.
func (sp *SerialPort) SetConfig(cfg Config) error {
	if err := checkConfigParam(cfg); err != nil {
		return err
	}

	baudrate := uint32(cfg.BaudRate)
	bytesize := dcbByteSizeMap[cfg.DataBits]
	parity := dcbParityMap[cfg.Parity]
	stopbits := dcbStopBitsMap[cfg.StopBits]
	dcb := win32DCB{
		DCBlength: uint32(unsafe.Sizeof(win32DCB{})),
		BaudRate:  baudrate,
		ByteSize:  bytesize,
		Parity:    parity,
		StopBits:  stopbits,
	}
	if err := win32SetCommState(sp.handle, &dcb); err != nil {
		return err
	}

	timeoutMs := uint32(cfg.Timeout.Milliseconds())
	var commTimeouts windows.CommTimeouts
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
