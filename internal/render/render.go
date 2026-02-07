// Package render provides a minimal ANSI terminal renderer built on a
// fixed-size character grid.
package render

import (
	"bufio"
	"fmt"
	"io"
)

type Renderer struct {
	out    io.Writer
	bw     *bufio.Writer
	width  int
	height int

	clearedOnce bool
}

func NewRenderer(out io.Writer, width, height int) *Renderer {
	return &Renderer{
		out:    out,
		bw:     bufio.NewWriterSize(out, 64*1024),
		width:  width,
		height: height,
	}
}

func (r *Renderer) HideCursor() error {
	if _, err := fmt.Fprint(r.bw, "\x1b[?25l"); err != nil {
		return err
	}
	return r.bw.Flush()
}

func (r *Renderer) ShowCursor() error {
	if _, err := fmt.Fprint(r.bw, "\x1b[?25h"); err != nil {
		return err
	}
	return r.bw.Flush()
}

func (r *Renderer) Draw(f *Frame) error {
	if f.Width != r.width || f.Height != r.height {
		return fmt.Errorf(
			"frame size %dx%d != renderer %dx%d",
			f.Width, f.Height, r.width, r.height,
		)
	}

	// Clear once to start from a clean screen, then just move cursor home.
	if !r.clearedOnce {
		if _, err := fmt.Fprint(r.bw, "\x1b[2J"); err != nil {
			return err
		}
		r.clearedOnce = true
	}
	if _, err := fmt.Fprint(r.bw, "\x1b[H"); err != nil {
		return err
	}

	// Write frame content. Avoid newline after last row to prevent scrolling.
	for y := 0; y < r.height; y++ {
		row := f.Cells[y*r.width : (y+1)*r.width]
		if _, err := r.bw.Write(row); err != nil {
			return err
		}
		if y != r.height-1 {
			if _, err := r.bw.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	return r.bw.Flush()
}
