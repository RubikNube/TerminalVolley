// Package render provides a minimal ANSI terminal renderer built on a
// fixed-size character grid.
package render

const (
	BlobWidth  = 3
	BlobHeight = 2
)

// Frame is a fixed-size 2D grid stored as a flat byte slice (row-major).
// Cells must contain exactly Width*Height bytes.
type Frame struct {
	Width  int
	Height int
	Cells  []byte
}

func NewFrame(width, height int) *Frame {
	f := &Frame{
		Width:  width,
		Height: height,
		Cells:  make([]byte, width*height),
	}
	f.Clear(' ')
	return f
}

func (f *Frame) Clear(ch byte) {
	for i := range f.Cells {
		f.Cells[i] = ch
	}
}

func (f *Frame) Set(x, y int, ch byte) {
	if x < 0 || y < 0 || x >= f.Width || y >= f.Height {
		return
	}
	f.Cells[y*f.Width+x] = ch
}

func (f *Frame) DrawGround() {
	y := f.Height - 1
	for x := 0; x < f.Width; x++ {
		f.Set(x, y, '_')
	}
}

func (f *Frame) DrawNet() {
	netX := f.Width / 2
	// Draw a simple vertical net rising from the ground.
	for y := f.Height - 2; y >= f.Height-8 && y >= 0; y-- {
		f.Set(netX, y, '|')
	}
}

func (f *Frame) DrawBlob(x, y int) {
	// Simple 3x2 blob.
	f.Set(x-1, y-1, 'O')
	f.Set(x, y-1, 'O')
	f.Set(x+1, y-1, 'O')
	f.Set(x-1, y, 'O')
	f.Set(x, y, 'O')
	f.Set(x+1, y, 'O')
}

func (f *Frame) DrawBall(x, y int) {
	// 1x1 ball sprite at (x, y)
	f.Set(x, y, '*')
}
