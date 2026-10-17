package main

import (
	"image"
	"image/color"
)

type rawImage struct {
	data []byte
	w, h int
}

func (r *rawImage) ColorModel() color.Model { return color.NRGBAModel }
func (r *rawImage) Bounds() image.Rectangle { return image.Rect(0, 0, r.w, r.h) }
func (r *rawImage) At(x, y int) color.Color {
	i := (y*r.w + x) * 4
	return color.NRGBA{R: r.data[i], G: r.data[i+1], B: r.data[i+2], A: r.data[i+3]}
}
