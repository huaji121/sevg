package main

import (
	"fmt"
	"image/png"
	"os"

	"sevg/internal/core"
	"sevg/internal/render"
)

func main() {
	data := render.BuildAtlas()
	i := (84*core.AtlasW + 12) * 4
	fmt.Printf("raw center=%d,%d,%d,%d\n", data[i], data[i+1], data[i+2], data[i+3])
	f, _ := os.Create("atlas.png")
	png.Encode(f, &rawImage{data: data, w: core.AtlasW, h: core.AtlasH})
	f.Close()
	f2, _ := os.Open("atlas.png")
	im, _ := png.Decode(f2)
	f2.Close()
	r, g, b, a := im.At(12, 84).RGBA()
	fmt.Printf("decoded center=%d,%d,%d,%d\n", r>>8, g>>8, b>>8, a>>8)
}
