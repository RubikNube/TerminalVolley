package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terminalvolley/internal/render"
)

func main() {
	const (
		w   = 80
		h   = 24
		fps = 30
	)

	r := render.NewRenderer(os.Stdout, w, h)

	// Ensure terminal is restored on Ctrl+C.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		_ = r.ShowCursor()
		fmt.Fprint(os.Stdout, "\x1b[0m\x1b[2J\x1b[H")
		os.Exit(130)
	}()

	if err := r.HideCursor(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	ticker := time.NewTicker(time.Second / fps)
	defer ticker.Stop()

	// Demo positions (replace later with a game state).
	px1, px2 := 20, 60
	py := h - 3
	bx, by := w/2, 5
	vx, vy := 1, 1

	for range ticker.C {
		// Simple demo motion for the ball.
		bx += vx
		by += vy
		if bx <= 1 || bx >= w-2 {
			vx = -vx
		}
		if by <= 1 || by >= h-6 {
			vy = -vy
		}

		frame := render.NewFrame(w, h)

		// Draw scene.
		frame.Clear(' ')
		frame.DrawGround()
		frame.DrawNet()
		frame.DrawBlob(px1, py)
		frame.DrawBlob(px2, py)
		frame.DrawBall(bx, by)

		if err := r.Draw(frame); err != nil && !errors.Is(err, syscall.EPIPE) {
			fmt.Fprintln(os.Stderr, err)
			break
		}
	}

	_ = r.ShowCursor()
}
