package render

import "testing"

func TestNewFrame_InitializesAndClears(t *testing.T) {
	f := NewFrame(4, 3)
	if f.Width != 4 || f.Height != 3 {
		t.Fatalf("unexpected dimensions: got %dx%d", f.Width, f.Height)
	}
	if len(f.Cells) != 12 {
		t.Fatalf("unexpected Cells length: got %d want %d", len(f.Cells), 12)
	}
	for i, b := range f.Cells {
		if b != ' ' {
			t.Fatalf("cell %d: got %q want space", i, b)
		}
	}
}

func TestFrameClear_FillsAllCells(t *testing.T) {
	f := NewFrame(3, 2)
	f.Clear('x')
	for i, b := range f.Cells {
		if b != 'x' {
			t.Fatalf("cell %d: got %q want %q", i, b, 'x')
		}
	}
}

func TestFrameSet_InBounds(t *testing.T) {
	f := NewFrame(3, 2)
	f.Set(1, 1, 'A')
	if got := f.Cells[1*f.Width+1]; got != 'A' {
		t.Fatalf("expected set to write: got %q want %q", got, 'A')
	}
}

func TestFrameSet_OutOfBounds_DoesNotModify(t *testing.T) {
	f := NewFrame(3, 2)
	f.Clear('.')

	f.Set(-1, 0, 'X')
	f.Set(0, -1, 'X')
	f.Set(3, 0, 'X')
	f.Set(0, 2, 'X')

	for i, b := range f.Cells {
		if b != '.' {
			t.Fatalf("cell %d modified: got %q want %q", i, b, '.')
		}
	}
}

func TestDrawGround_DrawsUnderscoresOnLastRow(t *testing.T) {
	f := NewFrame(5, 4)
	f.Clear(' ')
	f.DrawGround()

	y := f.Height - 1
	for x := 0; x < f.Width; x++ {
		if got := f.Cells[y*f.Width+x]; got != '_' {
			t.Fatalf("ground at (%d,%d): got %q want %q", x, y, got, '_')
		}
	}
}

func TestDrawNet_DrawsVerticalLineAtMid(t *testing.T) {
	f := NewFrame(10, 12)
	f.Clear(' ')
	f.DrawNet()

	netX := f.Width / 2
	startY := f.Height - 2
	endY := f.Height - 8
	if endY < 0 {
		endY = 0
	}
	for y := startY; y >= endY; y-- {
		if got := f.Cells[y*f.Width+netX]; got != '|' {
			t.Fatalf("net at (%d,%d): got %q want %q", netX, y, got, '|')
		}
	}

	// Ensure it does not draw on the ground row (height-1)
	groundY := f.Height - 1
	if got := f.Cells[groundY*f.Width+netX]; got == '|' {
		t.Fatalf("net should not draw on ground row, but got %q at (%d,%d)", got, netX, groundY)
	}
}

func TestDrawBlob_Draws3x2Centered(t *testing.T) {
	f := NewFrame(7, 5)
	f.Clear(' ')
	f.DrawBlob(3, 3)

	points := [][2]int{
		{2, 2}, {3, 2}, {4, 2},
		{2, 3}, {3, 3}, {4, 3},
	}
	for _, p := range points {
		x, y := p[0], p[1]
		if got := f.Cells[y*f.Width+x]; got != 'O' {
			t.Fatalf("blob at (%d,%d): got %q want %q", x, y, got, 'O')
		}
	}
}

func TestDrawBlob_ClipsAtEdges(t *testing.T) {
	f := NewFrame(2, 2)
	f.Clear('.')

	// Center at (0,0) would normally attempt to draw negative coords;
	// only (0,0) and (1,0) and (0,1) and (1,1) are possible, but due to 3x2 shape
	// with offsets, we expect only positions that land in-bounds to be written.
	f.DrawBlob(0, 0)

	// In-bounds writes possible from:
	// y-1 = -1 => ignored
	// y   = 0  => x-1=-1 ignored, x=0 set, x+1=1 set
	if got := f.Cells[0*f.Width+0]; got != 'O' {
		t.Fatalf("expected in-bounds blob write at (0,0): got %q want %q", got, 'O')
	}
	if got := f.Cells[0*f.Width+1]; got != 'O' {
		t.Fatalf("expected in-bounds blob write at (1,0): got %q want %q", got, 'O')
	}
	// Bottom row should remain '.' because y=0 only affects y=0 for this call (y-1 ignored)
	if got := f.Cells[1*f.Width+0]; got != '.' {
		t.Fatalf("unexpected write at (0,1): got %q want %q", got, '.')
	}
	if got := f.Cells[1*f.Width+1]; got != '.' {
		t.Fatalf("unexpected write at (1,1): got %q want %q", got, '.')
	}
}

func TestDrawBall_DrawsAsterisk(t *testing.T) {
	f := NewFrame(3, 3)
	f.Clear(' ')
	f.DrawBall(1, 2)
	if got := f.Cells[2*f.Width+1]; got != '*' {
		t.Fatalf("ball at (1,2): got %q want %q", got, '*')
	}
}

func TestDrawGroundAndNet_DoNotPanicOnSmallFrames(t *testing.T) {
	f := NewFrame(1, 1)
	f.DrawGround()
	f.DrawNet() // should clip via Set without panic
}
