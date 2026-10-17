package audio

import (
	"testing"

	"sevg/assets"
)

func TestDecodeEmbeddedSound(t *testing.T) {
	data, err := assets.Sounds.ReadFile("sounds/step.wav")
	if err != nil {
		t.Fatal(err)
	}
	samples, err := decodeWAV(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) < 100 {
		t.Fatalf("decoded only %d samples", len(samples))
	}
}

func TestNilAudioEngineIsNoop(t *testing.T) {
	var e *Engine
	e.Play("step")
	e.Update()
	e.Close()
}
