package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terminalvolley/internal/input"
	"terminalvolley/internal/render"
)

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func main() {
	const (
		w   = 80
		h   = 24
		fps = 30

		groundY = h - 2 // blob "feet" y (since ground is at h-1)
		moveSpd = 2
		jumpVel = -6.5
		grav    = 0.6
	)

	r := render.NewRenderer(os.Stdout, w, h)

	term, err := input.MakeRaw()
	if err != nil {
		fmt.Fprintln(os.Stderr, "raw mode:", err)
		return
	}
	defer func() { _ = term.Restore() }()

	keys := input.StartKeyReader()

	// Ensure terminal is restored on Ctrl+C.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		_ = term.Restore()
		_ = r.ShowCursor()
		fmt.Fprint(os.Stdout, "\x1b[0m\x1b[2J\x1b[H")
		os.Exit(130)
	}()

	if err := r.HideCursor(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	// Players
	p1x, p2x := 20, 60
	p1y, p2y := groundY, groundY
	p1vy, p2vy := 0.0, 0.0
	p1OnGround, p2OnGround := true, true

	// Ball demo
	bx, by := w/2, 5
	vx, vy := 1, 1

	// Simple "pressed this frame" jump latch
	var p1JumpReq, p2JumpReq bool

	for range ticker.C {
		// Drain keys available this frame.
		for {
			select {
			case k, ok := <-keys:
				if !ok {
					return
				}
				switch k {
				case 'q', 'Q':
					_ = r.ShowCursor()
					return

				// Player 1: A/D move, W jump
				case 'a', 'A':
					p1x -= moveSpd
				case 'd', 'D':
					p1x += moveSpd
				case 'w', 'W':
					p1JumpReq = true

				// Player 2: J/L move, I jump
				case 'j', 'J':
					p2x -= moveSpd
				case 'l', 'L':
					p2x += moveSpd
				case 'i', 'I':
					p2JumpReq = true
				}
			default:
				goto keysDone
			}
		}
	keysDone:

		// Clamp to halves (don’t cross net). Blob is 3 chars wide, keep margin.
		netX := w / 2
		p1x = clamp(p1x, 2, netX-2)
		p2x = clamp(p2x, netX+2, w-3)

		// Jump if requested and on ground.
		if p1JumpReq && p1OnGround {
			p1vy = jumpVel
			p1OnGround = false
		}
		if p2JumpReq && p2OnGround {
			p2vy = jumpVel
			p2OnGround = false
		}
		p1JumpReq, p2JumpReq = false, false

		// Integrate vertical motion.
		if !p1OnGround {
			p1vy += grav
			p1y = int(math.Round(float64(p1y) + p1vy))
			if p1y >= groundY {
				p1y = groundY
				p1vy = 0
				p1OnGround = true
			}
		}
		if !p2OnGround {
			p2vy += grav
			p2y = int(math.Round(float64(p2y) + p2vy))
			if p2y >= groundY {
				p2y = groundY
				p2vy = 0
				p2OnGround = true
			}
		}

		// Ball demo motion.
		bx += vx
		by += vy
		if bx <= 1 || bx >= w-2 {
			vx = -vx
		}
		if by <= 1 || by >= h-6 {
			vy = -vy
		}

		frame := render.NewFrame(w, h)
		frame.Clear(' ')
		frame.DrawGround()
		frame.DrawNet()
		frame.DrawBlob(p1x, p1y)
		frame.DrawBlob(p2x, p2y)
		frame.DrawBall(bx, by)

		if err := r.Draw(frame); err != nil && !errors.Is(err, syscall.EPIPE) {
			fmt.Fprintln(os.Stderr, err)
			break
		}
	}

	_ = r.ShowCursor()
}
