package input

import (
	"os"

	"golang.org/x/sys/unix"
)

type RawTerminal struct {
	fd    int
	state *unix.Termios
}

func MakeRaw() (*RawTerminal, error) {
	fd := int(os.Stdin.Fd())

	oldState, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}
	newState := *oldState

	// Rough equivalent of cfmakeraw.
	newState.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	newState.Oflag &^= unix.OPOST
	newState.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	newState.Cflag &^= unix.CSIZE | unix.PARENB
	newState.Cflag |= unix.CS8

	// Read returns as soon as 1 byte is available.
	newState.Cc[unix.VMIN] = 1
	newState.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &newState); err != nil {
		return nil, err
	}

	return &RawTerminal{fd: fd, state: oldState}, nil
}

func (t *RawTerminal) Restore() error {
	return unix.IoctlSetTermios(t.fd, unix.TCSETS, t.state)
}
