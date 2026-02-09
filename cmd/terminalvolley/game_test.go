package main

import "testing"

func TestGame_WaitingServe_BallFrozenUntilServe(t *testing.T) {
	g := NewGame(80, 24, 200, Controls{
		P1Left:     'A',
		P1Right:    'D',
		P1Jump:     'W',
		P2Left:     'J',
		P2Right:    'L',
		P2Jump:     'I',
		ServeLeft:  'S',
		ServeRight: 'K',
	})

	if !g.WaitingServe() {
		t.Fatalf("expected waitingServe=true at start")
	}

	bx1, by1, vx1, vy1 := g.Ball()
	g.Step()
	bx2, by2, vx2, vy2 := g.Ball()

	if bx1 != bx2 || by1 != by2 || vx1 != vx2 || vy1 != vy2 {
		t.Fatalf("expected ball frozen while waitingServe; before=%v after=%v",
			[]float64{bx1, by1, vx1, vy1}, []float64{bx2, by2, vx2, vy2})
	}

	g.PressKey('S')
	if g.WaitingServe() {
		t.Fatalf("expected waitingServe=false after serve key")
	}

	_, _, vx3, _ := g.Ball()
	if vx3 == 0 {
		t.Fatalf("expected non-zero ball vx after serve")
	}
}

func TestGame_Scoring_LeftSideAwardsP2AndResetsToWaitingServe(t *testing.T) {
	g := NewGame(80, 24, 200, Controls{
		P1Left:     'A',
		P1Right:    'D',
		P1Jump:     'W',
		P2Left:     'J',
		P2Right:    'L',
		P2Jump:     'I',
		ServeLeft:  'S',
		ServeRight: 'K',
	})

	// start serve so physics runs
	g.PressKey('S')

	// force ball to "hit ground" on left side
	g.bx = float64(g.netX) - 10
	g.by = g.groundBallY + 0.001
	g.vx, g.vy = 0, 0

	g.Step()

	p1, p2 := g.Score()
	if p1 != 0 || p2 != 1 {
		t.Fatalf("expected score 0:1, got %d:%d", p1, p2)
	}
	if !g.WaitingServe() {
		t.Fatalf("expected waitingServe=true after point")
	}
	_, _, vx, vy := g.Ball()
	if vx != 0 || vy != 0 {
		t.Fatalf("expected ball velocity reset to 0 after point, got vx=%v vy=%v", vx, vy)
	}
}

func TestGame_Scoring_RightSideAwardsP1AndResetsToWaitingServe(t *testing.T) {
	g := NewGame(80, 24, 200, Controls{
		P1Left:     'A',
		P1Right:    'D',
		P1Jump:     'W',
		P2Left:     'J',
		P2Right:    'L',
		P2Jump:     'I',
		ServeLeft:  'S',
		ServeRight: 'K',
	})

	g.PressKey('S')

	// force ball to "hit ground" on right side
	g.bx = float64(g.netX) + 10
	g.by = g.groundBallY + 0.001
	g.vx, g.vy = 0, 0

	g.Step()

	p1, p2 := g.Score()
	if p1 != 1 || p2 != 0 {
		t.Fatalf("expected score 1:0, got %d:%d", p1, p2)
	}
	if !g.WaitingServe() {
		t.Fatalf("expected waitingServe=true after point")
	}
}
