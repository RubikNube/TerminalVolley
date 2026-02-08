package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"terminalvolley/config"
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

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}

func resetServe(w int, groundBallY float64, toLeft bool) (bx, by, vx, vy float64) {
	// Serve from the middle to make the first hit easier.
	bx = float64(w) * 0.5
	by = 4

	// Slow, playable serve speed (cells/sec).
	vx = 6.0
	if toLeft {
		vx = -vx
	}
	vy = 0
	return
}

func main() {
	const (
		w   = 80
		h   = 24
		fps = 200

		groundY = h - 2 // blob "feet" y (since ground is at h-1)
	)

	controls, err := config.LoadControls("config/controls.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "load controls:", err)
		return
	}

	dt := 1.0 / float64(fps)

	// Player tuning (units: cells/sec and cells/sec^2)
	const (
		moveSpd    = 20.0  // cells/sec
		airMoveSpd = 12.0  // cells/sec
		jumpVel    = -28.0 // cells/sec
		grav       = 32.0  // cells/sec^2
	)

	// Ball physics tuning (time-based; units: cells/sec and cells/sec^2)
	const (
		ballGrav      = 18.0 // cells/sec^2
		ballRestitute = 0.78
		playerKick    = 10.0 // cells/sec impulse on hit (approx)
		playerCarry   = 0.30
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

	// Players (positions in cells; velocities in cells/sec)
	p1x, p2x := 20.0, 60.0
	p1y, p2y := float64(groundY), float64(groundY)
	p1vy, p2vy := 0.0, 0.0
	p1OnGround, p2OnGround := true, true

	// Held input state (for smooth air control). Note: raw stdin gives no key-up
	// events; last pressed direction keeps applying until opposite is pressed.
	var p1LeftHeld, p1RightHeld bool
	var p2LeftHeld, p2RightHeld bool

	// Horizontal velocity command (cells/sec)
	p1hx, p2hx := 0.0, 0.0

	// Track player horizontal velocity estimate (cells/sec) for ball carry.
	p1prevX, p2prevX := p1x, p2x
	p1vx, p2vx := 0.0, 0.0

	// Score
	p1Score, p2Score := 0, 0

	// Ball (position in cells; velocity in cells/sec)
	groundBallY := float64(h - 2) // treat ball "ground" one above '_' line
	bx, by, vx, vy := resetServe(w, groundBallY, true)

	// Serve gating: ball is frozen until a player starts the serve.
	waitingServe := true
	serveToLeft := true // direction of the next serve

	// Simple "pressed this frame" jump latch
	var p1JumpReq, p2JumpReq bool

	netX := w / 2
	netTopY := h - 8
	netBottomY := h - 2

	for range ticker.C {
		// Drain keys available this frame.
		for {
			select {
			case k, ok := <-keys:
				if !ok {
					return
				}

				// Compare against config (case-insensitive by normalizing to upper).
				kU := string([]byte{byte(k)})
				if k >= 'a' && k <= 'z' {
					kU = string([]byte{byte(k - 32)})
				}

				if kU == controls.Quit {
					_ = r.ShowCursor()
					return
				}

				// Serve start: P1 uses 'S', P2 uses 'K' (still hardcoded)
				switch k {
				case 's', 'S':
					if waitingServe {
						serveToLeft = true
						bx, by, vx, vy = resetServe(w, groundBallY, serveToLeft)
						waitingServe = false
					}
				case 'k', 'K':
					if waitingServe {
						serveToLeft = false
						bx, by, vx, vy = resetServe(w, groundBallY, serveToLeft)
						waitingServe = false
					}
				}

				// Player 1
				if kU == controls.Player1.Left {
					p1LeftHeld = true
					p1RightHeld = false
				} else if kU == controls.Player1.Right {
					p1RightHeld = true
					p1LeftHeld = false
				} else if kU == controls.Player1.Jump {
					p1JumpReq = true
				}

				// Player 2
				if kU == controls.Player2.Left {
					p2LeftHeld = true
					p2RightHeld = false
				} else if kU == controls.Player2.Right {
					p2RightHeld = true
					p2LeftHeld = false
				} else if kU == controls.Player2.Jump {
					p2JumpReq = true
				}

			default:
				goto keysDone
			}
		}
	keysDone:

		// Derive horizontal velocity command from held state (cells/sec).
		p1hx = 0
		p1spd := moveSpd
		if !p1OnGround {
			p1spd = airMoveSpd
		}
		if p1LeftHeld {
			p1hx -= p1spd
		}
		if p1RightHeld {
			p1hx += p1spd
		}

		p2hx = 0
		p2spd := moveSpd
		if !p2OnGround {
			p2spd = airMoveSpd
		}
		if p2LeftHeld {
			p2hx -= p2spd
		}
		if p2RightHeld {
			p2hx += p2spd
		}

		// Apply horizontal movement every frame (ground or air).
		p1x += p1hx * dt
		p2x += p2hx * dt

		// Clamp to halves (don’t cross net). Blob is 3 chars wide, keep margin.
		if p1x < 2 {
			p1x = 2
		}
		if p1x > float64(netX-2) {
			p1x = float64(netX - 2)
		}
		if p2x < float64(netX+2) {
			p2x = float64(netX + 2)
		}
		if p2x > float64(w-3) {
			p2x = float64(w - 3)
		}

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
			p1vy += grav * dt
			p1y += p1vy * dt
			if p1y >= float64(groundY) {
				p1y = float64(groundY)
				p1vy = 0
				p1OnGround = true
			}
		}
		if !p2OnGround {
			p2vy += grav * dt
			p2y += p2vy * dt
			if p2y >= float64(groundY) {
				p2y = float64(groundY)
				p2vy = 0
				p2OnGround = true
			}
		}

		// Update player horizontal velocity estimate (cells/sec).
		p1vx = (p1x - p1prevX) / dt
		p2vx = (p2x - p2prevX) / dt
		p1prevX, p2prevX = p1x, p2x

		// ---- Ball physics ----
		const ballR = 1.05 // slightly forgiving vs 2x2 sprite

		if !waitingServe {
			// Gravity
			vy += ballGrav * dt

			// Integrate
			bx += vx * dt
			by += vy * dt

			// Walls (keep 1-cell margin for borderless grid)
			if bx <= 1+ballR {
				bx = 1 + ballR
				vx = -vx * ballRestitute
			} else if bx >= float64(w-2)-ballR {
				bx = float64(w-2) - ballR
				vx = -vx * ballRestitute
			}

			// Ceiling
			if by <= 1+ballR {
				by = 1 + ballR
				vy = -vy * ballRestitute
			}

			// Net collision: vertical segment at netX from netTopY..netBottomY
			if int(math.Round(by)) >= netTopY && int(math.Round(by)) <= netBottomY {
				if math.Abs(bx-float64(netX)) < 0.6+ballR {
					if bx < float64(netX) {
						bx = float64(netX) - (1 + ballR)
					} else {
						bx = float64(netX) + (1 + ballR)
					}
					vx = -vx * ballRestitute
				}
			}

			// Player collisions: approximate blobs as circles.
			// Blob center: (px, py-0.5). Radius tuned for 3x2 glyph.
			const blobR = 1.7
			p1cx, p1cy := p1x, p1y-0.5
			p2cx, p2cy := p2x, p2y-0.5

			hitPlayer := func(cx, cy, pvx float64) {
				dx := bx - cx
				dy := by - cy
				d2 := dx*dx + dy*dy
				rr := (blobR + ballR) * (blobR + ballR)
				if d2 <= rr && d2 > 0.0001 {
					d := math.Sqrt(d2)
					nx, ny := dx/d, dy/d

					// Push ball out of the blob.
					penetration := (blobR + ballR) - d
					bx += nx * penetration
					by += ny * penetration

					// Reflect velocity along normal, then add "kick" and carry.
					dot := vx*nx + vy*ny
					if dot < 0 {
						vx = vx - 2*dot*nx
						vy = vy - 2*dot*ny
					}

					// Apply additional impulse away from player.
					vx += nx * playerKick
					vy += ny * playerKick

					// Horizontal carry from player movement.
					vx += pvx * playerCarry

					// Dampen slightly (avoid infinite energy).
					vx *= 0.98
					vy *= 0.98
				}
			}

			hitPlayer(p1cx, p1cy, p1vx)
			hitPlayer(p2cx, p2cy, p2vx)

			// Ground / scoring: if ball hits ground, award point and reset.
			if by >= groundBallY-ballR {
				if bx < float64(netX) {
					p2Score++
					waitingServe = true
					serveToLeft = false
					bx, by, vx, vy = resetServe(w, groundBallY, serveToLeft)
					vx, vy = 0, 0
				} else {
					p1Score++
					waitingServe = true
					serveToLeft = true
					bx, by, vx, vy = resetServe(w, groundBallY, serveToLeft)
					vx, vy = 0, 0
				}
			}
		}

		// ---- Render ----
		frame := render.NewFrame(w, h)
		frame.Clear(' ')
		frame.DrawGround()
		frame.DrawNet()
		frame.DrawBlob(int(math.Round(p1x)), int(math.Round(p1y)))
		frame.DrawBlob(int(math.Round(p2x)), int(math.Round(p2y)))
		frame.DrawBall(int(math.Round(bx)), int(math.Round(by)))

		// Top row UI.
		score := fmt.Sprintf("P1 %d : %d P2", p1Score, p2Score)
		for i := 0; i < len(score) && i < w; i++ {
			frame.Set(i, 0, score[i])
		}
		if waitingServe {
			msg := "  (S = serve left, K = serve right)"
			for i := 0; i < len(msg) && i+len(score) < w; i++ {
				frame.Set(len(score)+i, 0, msg[i])
			}
		}

		if err := r.Draw(frame); err != nil && !errors.Is(err, syscall.EPIPE) {
			fmt.Fprintln(os.Stderr, err)
			break
		}
	}

	_ = r.ShowCursor()
}
