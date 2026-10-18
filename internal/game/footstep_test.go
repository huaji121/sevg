package game

import "testing"

func TestFootstepTimerPausesWhileAirborne(t *testing.T) {
	g := newTestGame(t)
	g.stepTimer = 1.0
	g.player.Body.OnGround = false
	g.updateFootsteps(0.016)
	if g.stepTimer != 1.0 {
		t.Fatalf("step timer reset in air: %f", g.stepTimer)
	}
	g.player.Body.OnGround = true
	g.player.Body.Vel.X = 4
	g.updateFootsteps(0.016)
	if g.stepTimer >= 1.0 {
		t.Fatalf("step timer did not accumulate after landing: %f", g.stepTimer)
	}
}
