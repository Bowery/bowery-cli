// +build linux darwin

// Copyright 2013-2014 Bowery, Inc.
package prompt

import (
	"bufio"
	"os"
	"syscall"
	"unsafe"

	"github.com/Bowery/bowery/errors"
)

var unsupported = []string{"", "dumb", "cons25"}

// supportedTerminal checks if the terminal supports ansi escapes.
func supportedTerminal() bool {
	term := os.Getenv("TERM")

	for _, t := range unsupported {
		if t == term {
			return false
		}
	}

	return true
}

// winsize contains the size for the terminal.
type winsize struct {
	rows   uint16
	cols   uint16
	xpixel uint16
	ypixel uint16
}

// TerminalSize retrieves the cols/rows for the terminal connected to out.
func TerminalSize(out *os.File) (int, int, error) {
	ws := new(winsize)

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, out.Fd(),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(ws)))
	if err != 0 {
		return 0, 0, errors.NewStackError(err)
	}

	return int(ws.cols), int(ws.rows), nil
}

// IsNotTerminal checks if an error is related to io not being a terminal.
func IsNotTerminal(err error) bool {
	if err == syscall.ENOTTY {
		return true
	}

	return false
}

// Terminal contains the state for raw terminal input.
type Terminal struct {
	In        *os.File
	Out       *os.File
	supported bool
	origMode  syscall.Termios
}

// NewTerminal creates a terminal and sets it to raw input mode.
func NewTerminal() (*Terminal, error) {
	if inReader == nil {
		inReader = bufio.NewReader(os.Stdin)
	}
	term := &Terminal{In: os.Stdin, Out: os.Stdout}
	if !supportedTerminal() {
		return term, nil
	}

	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, term.In.Fd(),
		uintptr(tcgets), uintptr(unsafe.Pointer(&term.origMode)))
	if err != 0 {
		if IsNotTerminal(err) {
			return term, nil
		}

		return nil, errors.NewStackError(err)
	}
	mode := term.origMode
	term.supported = true

	// Set new mode flags, for reference see cfmakeraw(3).
	mode.Iflag &^= (syscall.BRKINT | syscall.IGNBRK | syscall.ICRNL |
		syscall.INLCR | syscall.IGNCR | syscall.ISTRIP | syscall.IXON |
		syscall.PARMRK)

	mode.Oflag &^= syscall.OPOST

	mode.Lflag &^= (syscall.ECHO | syscall.ECHONL | syscall.ICANON |
		syscall.ISIG | syscall.IEXTEN)

	mode.Cflag &^= (syscall.CSIZE | syscall.PARENB)
	mode.Cflag |= syscall.CS8

	// Set controls; min num of bytes, and timeouts.
	mode.Cc[syscall.VMIN] = 1
	mode.Cc[syscall.VTIME] = 0

	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, term.In.Fd(),
		uintptr(tcsetsf), uintptr(unsafe.Pointer(&mode)))
	if err != 0 {
		return nil, errors.NewStackError(err)
	}

	return term, nil
}

// Prompt gets a line with the prefix and echos input.
func (term *Terminal) Prompt(prefix string) (string, error) {
	if !term.supported {
		return term.simplePrompt(prefix)
	}

	buf := NewBuffer(prefix, term.Out, true)
	return term.prompt(buf, term.In)
}

// Password gets a line with the prefix and doesn't echo input.
func (term *Terminal) Password(prefix string) (string, error) {
	if !term.supported {
		return term.simplePrompt(prefix)
	}

	buf := NewBuffer(prefix, term.Out, false)
	return term.password(buf, term.In)
}

// Close disables the terminals raw input.
func (term *Terminal) Close() error {
	if term.supported {
		_, _, err := syscall.Syscall(syscall.SYS_IOCTL, term.In.Fd(),
			uintptr(tcsets), uintptr(unsafe.Pointer(&term.origMode)))
		if err != 0 {
			return errors.NewStackError(err)
		}
	}

	return nil
}
