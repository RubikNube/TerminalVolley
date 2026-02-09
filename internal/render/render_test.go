package render

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewRenderer_SetsDimensions(t *testing.T) {
	var out bytes.Buffer
	r := NewRenderer(&out, 10, 5)
	if r.width != 10 || r.height != 5 {
		t.Fatalf("expected width/height 10/5, got %d/%d", r.width, r.height)
	}
	if r.bw == nil {
		t.Fatalf("expected buffered writer to be initialized")
	}
}

func TestRenderer_HideShowCursor_WritesEscapesAndFlushes(t *testing.T) {
	var out bytes.Buffer
	r := NewRenderer(&out, 1, 1)

	if err := r.HideCursor(); err != nil {
		t.Fatalf("HideCursor error: %v", err)
	}
	if got := out.String(); got != "\x1b[?25l" {
		t.Fatalf("HideCursor output mismatch: %q", got)
	}

	out.Reset()
	r = NewRenderer(&out, 1, 1)

	if err := r.ShowCursor(); err != nil {
		t.Fatalf("ShowCursor error: %v", err)
	}
	if got := out.String(); got != "\x1b[?25h" {
		t.Fatalf("ShowCursor output mismatch: %q", got)
	}
}

func TestRenderer_Draw_SizeMismatch(t *testing.T) {
	var out bytes.Buffer
	r := NewRenderer(&out, 4, 2)

	f := &Frame{
		Width:  5,
		Height: 2,
		Cells:  make([]byte, 10),
	}

	err := r.Draw(f)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "frame size") {
		t.Fatalf("expected size mismatch error, got: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("expected no output on error, got %q", out.String())
	}
}

func TestRenderer_Draw_ClearsOnceAndHomesEveryDraw(t *testing.T) {
	var out bytes.Buffer
	r := NewRenderer(&out, 3, 2)

	f := &Frame{
		Width:  3,
		Height: 2,
		Cells:  []byte("abcdef"),
	}

	if err := r.Draw(f); err != nil {
		t.Fatalf("Draw(1) error: %v", err)
	}
	got1 := out.String()
	want1 := "\x1b[2J" + "\x1b[H" + "abc\r\ndef"
	if got1 != want1 {
		t.Fatalf("Draw(1) output mismatch:\n got: %q\nwant: %q", got1, want1)
	}

	out.Reset()
	if err := r.Draw(f); err != nil {
		t.Fatalf("Draw(2) error: %v", err)
	}
	got2 := out.String()
	want2 := "\x1b[H" + "abc\r\ndef"
	if got2 != want2 {
		t.Fatalf("Draw(2) output mismatch:\n got: %q\nwant: %q", got2, want2)
	}
}

func TestRenderer_Draw_NoTrailingNewlineOnLastRow(t *testing.T) {
	var out bytes.Buffer
	r := NewRenderer(&out, 2, 2)

	f := &Frame{
		Width:  2,
		Height: 2,
		Cells:  []byte("wxyz"),
	}

	if err := r.Draw(f); err != nil {
		t.Fatalf("Draw error: %v", err)
	}
	// Should include CRLF only between rows, not after last row.
	wantSuffix := "wx\r\nyz"
	if !strings.HasSuffix(out.String(), wantSuffix) {
		t.Fatalf("expected output to end with %q, got %q", wantSuffix, out.String())
	}
	if strings.HasSuffix(out.String(), "\r\n") {
		t.Fatalf("expected no trailing CRLF, got %q", out.String())
	}
}
