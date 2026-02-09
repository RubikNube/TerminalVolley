package main

import "math"

type Controls struct {
	P1Left, P1Right, P1Jump byte
	P2Left, P2Right, P2Jump byte
	ServeLeft, ServeRight   byte
}

type Game struct {
	W, H int

	// timing/tuning
	dt float64

	moveSpd    float64
	airMoveSpd float64
	jumpVel    float64
	grav       float64

	ballGrav      float64
	ballRestitute float64
	playerKick    float64
	playerCarry   float64

	groundY     int
	groundBallY float64

	netX         int
	netTopY      int
	netBottomY   int
	ballR        float64
	blobR        float64
	waitingServe bool
	serveToLeft  bool

	// players
	p1x, p2x               float64
	p1y, p2y               float64
	p1vy, p2vy             float64
	p1OnGround, p2OnGround bool

	p1LeftHeld, p1RightHeld bool
	p2LeftHeld, p2RightHeld bool

	p1JumpReq, p2JumpReq bool

	p1prevX, p2prevX float64
	p1vx, p2vx       float64

	// ball
	bx, by float64
	vx, vy float64

	// score
	p1Score, p2Score int

	controls Controls
}

func NewGame(w, h int, fps int, c Controls) *Game {
	g := &Game{
		W:  w,
		H:  h,
		dt: 1.0 / float64(fps),

		moveSpd:    20.0,
		airMoveSpd: 12.0,
		jumpVel:    -28.0,
		grav:       32.0,

		ballGrav:      18.0,
		ballRestitute: 0.78,
		playerKick:    10.0,
		playerCarry:   0.30,

		groundY:     h - 2,
		groundBallY: float64(h - 2),

		netX:       w / 2,
		netTopY:    h - 8,
		netBottomY: h - 2,

		ballR: 1.05,
		blobR: 1.7,

		waitingServe: true,
		serveToLeft:  true,

		controls: c,
	}

	// initial players
	g.p1x, g.p2x = 20.0, 60.0
	g.p1y, g.p2y = float64(g.groundY), float64(g.groundY)
	g.p1OnGround, g.p2OnGround = true, true
	g.p1prevX, g.p2prevX = g.p1x, g.p2x

	// initial ball position (frozen by waitingServe)
	g.bx, g.by, g.vx, g.vy = resetServe(w, true)
	g.vx, g.vy = 0, 0

	return g
}

func (g *Game) WaitingServe() bool { return g.waitingServe }
func (g *Game) Score() (int, int)  { return g.p1Score, g.p2Score }
func (g *Game) Ball() (float64, float64, float64, float64) {
	return g.bx, g.by, g.vx, g.vy
}

func (g *Game) PressKey(k byte) {
	// serve start
	switch k {
	case g.controls.ServeLeft, toUpper(g.controls.ServeLeft):
		if g.waitingServe {
			g.serveToLeft = true
			g.bx, g.by, g.vx, g.vy = resetServe(g.W, g.serveToLeft)
			g.waitingServe = false
		}
	case g.controls.ServeRight, toUpper(g.controls.ServeRight):
		if g.waitingServe {
			g.serveToLeft = false
			g.bx, g.by, g.vx, g.vy = resetServe(g.W, g.serveToLeft)
			g.waitingServe = false
		}
	}

	// player 1
	switch k {
	case g.controls.P1Left, toUpper(g.controls.P1Left):
		g.p1LeftHeld = true
		g.p1RightHeld = false
	case g.controls.P1Right, toUpper(g.controls.P1Right):
		g.p1RightHeld = true
		g.p1LeftHeld = false
	case g.controls.P1Jump, toUpper(g.controls.P1Jump):
		g.p1JumpReq = true
	}

	// player 2
	switch k {
	case g.controls.P2Left, toUpper(g.controls.P2Left):
		g.p2LeftHeld = true
		g.p2RightHeld = false
	case g.controls.P2Right, toUpper(g.controls.P2Right):
		g.p2RightHeld = true
		g.p2LeftHeld = false
	case g.controls.P2Jump, toUpper(g.controls.P2Jump):
		g.p2JumpReq = true
	}
}

func toUpper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - 32
	}
	return b
}

func (g *Game) hitPlayer(cx, cy, pvx float64) {
	dx := g.bx - cx
	dy := g.by - cy
	d2 := dx*dx + dy*dy
	rr := (g.blobR + g.ballR) * (g.blobR + g.ballR)
	if d2 <= rr && d2 > 0.0001 {
		d := math.Sqrt(d2)
		nx, ny := dx/d, dy/d

		penetration := (g.blobR + g.ballR) - d
		g.bx += nx * penetration
		g.by += ny * penetration

		dot := g.vx*nx + g.vy*ny
		if dot < 0 {
			g.vx = g.vx - 2*dot*nx
			g.vy = g.vy - 2*dot*ny
		}

		g.vx += nx * g.playerKick
		g.vy += ny * g.playerKick

		g.vx += pvx * g.playerCarry

		g.vx *= 0.98
		g.vy *= 0.98
	}
}

func (g *Game) Step() {
	dt := g.dt

	// movement command from held state
	p1hx := 0.0
	p1spd := g.moveSpd
	if !g.p1OnGround {
		p1spd = g.airMoveSpd
	}
	if g.p1LeftHeld {
		p1hx -= p1spd
	}
	if g.p1RightHeld {
		p1hx += p1spd
	}

	p2hx := 0.0
	p2spd := g.moveSpd
	if !g.p2OnGround {
		p2spd = g.airMoveSpd
	}
	if g.p2LeftHeld {
		p2hx -= p2spd
	}
	if g.p2RightHeld {
		p2hx += p2spd
	}

	// integrate horizontal
	g.p1x += p1hx * dt
	g.p2x += p2hx * dt

	// clamp to halves (don’t cross net). Blob is 3 chars wide, keep margin.
	if g.p1x < 2 {
		g.p1x = 2
	}
	if g.p1x > float64(g.netX-2) {
		g.p1x = float64(g.netX - 2)
	}
	if g.p2x < float64(g.netX+2) {
		g.p2x = float64(g.netX + 2)
	}
	if g.p2x > float64(g.W-3) {
		g.p2x = float64(g.W - 3)
	}

	// jump
	if g.p1JumpReq && g.p1OnGround {
		g.p1vy = g.jumpVel
		g.p1OnGround = false
	}
	if g.p2JumpReq && g.p2OnGround {
		g.p2vy = g.jumpVel
		g.p2OnGround = false
	}
	g.p1JumpReq, g.p2JumpReq = false, false

	// vertical integration
	if !g.p1OnGround {
		g.p1vy += g.grav * dt
		g.p1y += g.p1vy * dt
		if g.p1y >= float64(g.groundY) {
			g.p1y = float64(g.groundY)
			g.p1vy = 0
			g.p1OnGround = true
		}
	}
	if !g.p2OnGround {
		g.p2vy += g.grav * dt
		g.p2y += g.p2vy * dt
		if g.p2y >= float64(g.groundY) {
			g.p2y = float64(g.groundY)
			g.p2vy = 0
			g.p2OnGround = true
		}
	}

	// player horizontal velocity estimate
	g.p1vx = (g.p1x - g.p1prevX) / dt
	g.p2vx = (g.p2x - g.p2prevX) / dt
	g.p1prevX, g.p2prevX = g.p1x, g.p2x

	// ---- Ball physics ----
	if g.waitingServe {
		return
	}

	// gravity + integrate
	g.vy += g.ballGrav * dt
	g.bx += g.vx * dt
	g.by += g.vy * dt

	// walls
	if g.bx <= 1+g.ballR {
		g.bx = 1 + g.ballR
		g.vx = -g.vx * g.ballRestitute
	} else if g.bx >= float64(g.W-2)-g.ballR {
		g.bx = float64(g.W-2) - g.ballR
		g.vx = -g.vx * g.ballRestitute
	}

	// ceiling
	if g.by <= 1+g.ballR {
		g.by = 1 + g.ballR
		g.vy = -g.vy * g.ballRestitute
	}

	// net collision
	if int(math.Round(g.by)) >= g.netTopY && int(math.Round(g.by)) <= g.netBottomY {
		if math.Abs(g.bx-float64(g.netX)) < 0.6+g.ballR {
			if g.bx < float64(g.netX) {
				g.bx = float64(g.netX) - (1 + g.ballR)
			} else {
				g.bx = float64(g.netX) + (1 + g.ballR)
			}
			g.vx = -g.vx * g.ballRestitute
		}
	}

	// player collisions
	p1cx, p1cy := g.p1x, g.p1y-0.5
	p2cx, p2cy := g.p2x, g.p2y-0.5

	g.hitPlayer(p1cx, p1cy, g.p1vx)
	g.hitPlayer(p2cx, p2cy, g.p2vx)

	// ground/scoring
	if g.by >= g.groundBallY {
		if g.bx < float64(g.netX) {
			g.p2Score++
			g.waitingServe = true
			g.serveToLeft = false
			g.bx, g.by, g.vx, g.vy = resetServe(g.W, g.serveToLeft)
			g.vx, g.vy = 0, 0
		} else {
			g.p1Score++
			g.waitingServe = true
			g.serveToLeft = true
			g.bx, g.by, g.vx, g.vy = resetServe(g.W, g.serveToLeft)
			g.vx, g.vy = 0, 0
		}
	}
}
