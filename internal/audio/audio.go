package audio

import (
	"encoding/binary"
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/assets"
)

const (
	sampleRate  = 22050
	chunkFrames = 512
	maxVoices   = 24
)

var soundFiles = map[string]string{
	"break_stone": "break_stone.wav",
	"break_dirt":  "break_dirt.wav",
	"break_wood":  "break_wood.wav",
	"break_glass": "break_glass.wav",
	"break_grass": "break_grass.wav",
	"place":       "place.wav",
	"step":        "step.wav",
	"hurt":        "hurt.wav",
	"pickup":      "pickup.wav",
	"craft":       "craft.wav",
	"eat":         "eat.wav",
	"hit":         "hit.wav",
}

type voice struct {
	samples []int16
	pos     float32
	step    float32
	gain    float32
}

// Engine is a small software mixer feeding one SDL audio stream. It is safe to
// use as a nil pointer: all methods become no-ops when audio is unavailable.
type Engine struct {
	stream *sdl.AudioStream
	spec   sdl.AudioSpec
	sounds map[string][]int16
	voices []voice
	master float32
}

func Open() *Engine {
	e := &Engine{sounds: make(map[string][]int16), master: 0.7}
	for name, file := range soundFiles {
		data, err := assets.Sounds.ReadFile("sounds/" + file)
		if err != nil {
			continue
		}
		samples, err := decodeWAV(data)
		if err != nil {
			fmt.Printf("audio %s: %v\n", name, err)
			continue
		}
		e.sounds[name] = samples
	}
	spec := sdl.AudioSpec{Format: sdl.AUDIO_S16, Channels: 1, Freq: sampleRate}
	stream := sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDeviceStream(&spec, 0)
	if stream == nil {
		fmt.Println("audio unavailable (no output device)")
		return e
	}
	fmt.Println("audio:", sdl.GetCurrentAudioDriver())
	if err := stream.ResumeDevice(); err != nil {
		stream.Destroy()
		return e
	}
	e.stream = stream
	e.spec = spec
	return e
}

func (e *Engine) Close() {
	if e == nil || e.stream == nil {
		return
	}
	e.stream.Destroy()
	e.stream = nil
	e.voices = nil
}

func (e *Engine) Play(name string) {
	if e == nil || e.stream == nil {
		return
	}
	samples := e.sounds[name]
	if len(samples) == 0 {
		return
	}
	if len(e.voices) >= maxVoices {
		e.voices = e.voices[len(e.voices)-maxVoices+1:]
	}
	e.voices = append(e.voices, voice{
		samples: samples,
		step:    0.96 + float32(len(e.voices)%5)*0.02,
		gain:    0.85,
	})
}

// SetMasterVolume sets the mixed output gain in the range [0,1].
func (e *Engine) SetMasterVolume(v float32) {
	if e == nil {
		return
	}
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}
	e.master = v
}

// Update tops up the device queue with mixed samples. Called once per frame.
func (e *Engine) Update() {
	if e == nil || e.stream == nil || len(e.voices) == 0 {
		return
	}
	queued, err := e.stream.Queued()
	if err != nil {
		return
	}
	target := int32(e.spec.Freq * 2 / 12)
	for queued < target {
		buf := e.mix(chunkFrames)
		if err := e.stream.PutData(buf); err != nil {
			return
		}
		queued += int32(len(buf))
		if len(e.voices) == 0 {
			break
		}
	}
}

func (e *Engine) mix(frames int) []byte {
	acc := make([]int32, frames)
	active := e.voices[:0]
	for _, v := range e.voices {
		for i := 0; i < frames; i++ {
			idx := int(v.pos)
			if idx >= len(v.samples)-1 {
				break
			}
			frac := v.pos - float32(idx)
			a := float32(v.samples[idx])
			b := float32(v.samples[idx+1])
			acc[i] += int32((a + (b-a)*frac) * v.gain)
			v.pos += v.step
		}
		if int(v.pos) < len(v.samples)-1 {
			active = append(active, v)
		}
	}
	e.voices = active
	out := make([]byte, frames*2)
	for i, s := range acc {
		f := float32(s) / 32767 * e.master
		if f > 1 {
			f = 1
		}
		if f < -1 {
			f = -1
		}
		binary.LittleEndian.PutUint16(out[i*2:], uint16(int16(f*32767)))
	}
	return out
}

func decodeWAV(data []byte) ([]int16, error) {
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a RIFF/WAVE file")
	}
	pos := 12
	channels, rate, bits := 0, 0, 0
	format := 0
	var pcm []byte
	for pos+8 <= len(data) {
		id := string(data[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		body := pos + 8
		if size < 0 || body+size > len(data) {
			break
		}
		switch id {
		case "fmt ":
			if size >= 16 {
				format = int(binary.LittleEndian.Uint16(data[body:]))
				channels = int(binary.LittleEndian.Uint16(data[body+2:]))
				rate = int(binary.LittleEndian.Uint32(data[body+4:]))
				bits = int(binary.LittleEndian.Uint16(data[body+14:]))
			}
		case "data":
			pcm = data[body : body+size]
		}
		pos = body + size
		if pos&1 != 0 {
			pos++
		}
	}
	if format != 1 || channels < 1 || bits != 16 || rate <= 0 || len(pcm) < 2 {
		return nil, fmt.Errorf("unsupported WAV format")
	}
	frames := len(pcm) / (channels * 2)
	out := make([]int16, frames)
	for i := 0; i < frames; i++ {
		var sum int32
		for c := 0; c < channels; c++ {
			off := (i*channels + c) * 2
			sum += int32(int16(binary.LittleEndian.Uint16(pcm[off:])))
		}
		out[i] = int16(sum / int32(channels))
	}
	return out, nil
}
