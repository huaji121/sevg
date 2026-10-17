package render

import (
	"os"
	"strings"
	"testing"
)

func TestBlockShadersUseSkyAndBlockLight(t *testing.T) {
	vs, err := os.ReadFile("shaders/block.vert")
	if err != nil {
		t.Fatal(err)
	}
	fs, err := os.ReadFile("shaders/block.frag")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(vs), "vSky = aSky") || !strings.Contains(string(vs), "vBlock = aBlock") {
		t.Fatal("block vertex shader does not pass sky/block light varyings")
	}
	if !strings.Contains(string(fs), "uSun.rgb * vSky") || !strings.Contains(string(fs), "max(vBlock, dyn)") {
		t.Fatal("block fragment shader does not combine sky/block and dynamic light")
	}
}
