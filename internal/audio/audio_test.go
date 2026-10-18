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
func TestMasterVolumeClamp(t *testing.T) {
	e := &Engine{}
	e.SetMasterVolume(2)
	if e.master != 1 {
		t.Fatalf("high volume = %f", e.master)
	}
	e.SetMasterVolume(-1)
	if e.master != 0 {
		t.Fatalf("low volume = %f", e.master)
	}
	e.SetMasterVolume(0.4)
	if e.master != 0.4 {
		t.Fatalf("volume = %f", e.master)
	}
}
