package mathx

import (
	"math"
	"testing"
)

func TestPerspectiveMapsDepth(t *testing.T) {
	m := Perspective(Radians(70), 16.0/9.0, 0.1, 100)
	near := m.TransformPoint(V3(0, 0, -0.1))
	far := m.TransformPoint(V3(0, 0, -100))
	if math.Abs(float64(near.Z)) > 1e-4 {
		t.Fatalf("near z=%f", near.Z)
	}
	if math.Abs(float64(far.Z-1)) > 1e-3 {
		t.Fatalf("far z=%f", far.Z)
	}
}

func TestLookAtOrigin(t *testing.T) {
	m := LookAt(V3(0, 0, 5), V3(0, 0, 0), V3(0, 1, 0))
	p := m.TransformPoint(V3(0, 0, 0))
	if math.Abs(float64(p.Z+5)) > 1e-5 {
		t.Fatalf("origin z=%f", p.Z)
	}
}
