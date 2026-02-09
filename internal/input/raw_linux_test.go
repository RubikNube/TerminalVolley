package input

import (
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestMakeRaw_NonTTYReturnsError(t *testing.T) {
	// Force stdin to not be a TTY by redirecting it to /dev/null.
	old := os.Stdin
	defer func() { os.Stdin = old }()

	f, err := os.Open("/dev/null")
	if err != nil {
		t.Fatalf("open /dev/null: %v", err)
	}
	defer f.Close()
	os.Stdin = f

	rt, err := MakeRaw()
	if err == nil {
		// Best-effort cleanup if it somehow succeeded.
		_ = rt.Restore()
		t.Fatalf("expected error when stdin is not a TTY, got nil")
	}
}

func TestRawTerminal_Restore_InvalidFD(t *testing.T) {
	// Ensure Restore returns an error for an invalid fd.
	// Termios value doesn't matter; the syscall should fail first.
	rt := &RawTerminal{
		fd:    -1,
		state: &unix.Termios{},
	}
	if err := rt.Restore(); err == nil {
		t.Fatalf("expected error restoring invalid fd, got nil")
	}
}
