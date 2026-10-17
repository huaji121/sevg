package world

import "math"

// Noise is a seeded classic Perlin noise source. It is deterministic across
// runs and safe for concurrent reads.
type Noise struct {
	perm [512]uint8
	seed uint64
}

func NewNoise(seed int64) *Noise {
	n := &Noise{seed: uint64(seed)}
	var p [256]uint8
	for i := range p {
		p[i] = uint8(i)
	}
	s := uint64(seed) ^ 0x9e3779b97f4a7c15
	for i := 255; i > 0; i-- {
		s = splitmix64(s)
		j := int(s % uint64(i+1))
		p[i], p[j] = p[j], p[i]
	}
	for i := 0; i < 512; i++ {
		n.perm[i] = p[i&255]
	}
	return n
}

func splitmix64(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

func (n *Noise) hash2(x, y int) uint64 {
	h := n.seed ^ uint64(uint32(x))*0x9e3779b185ebca87 ^ uint64(uint32(y))*0xc2b2ae3d27d4eb4f
	return splitmix64(h)
}

func (n *Noise) hash3(x, y, z int) uint64 {
	h := n.seed ^ uint64(uint32(x))*0x9e3779b185ebca87 ^ uint64(uint32(y))*0xc2b2ae3d27d4eb4f ^ uint64(uint32(z))*0x165667b19e3779f9
	return splitmix64(h)
}

func fade(t float64) float64         { return t * t * t * (t*(t*6-15) + 10) }
func lerp64(a, b, t float64) float64 { return a + (b-a)*t }

func grad2(hash uint8, x, y float64) float64 {
	switch hash & 7 {
	case 0:
		return x + y
	case 1:
		return -x + y
	case 2:
		return x - y
	case 3:
		return -x - y
	case 4:
		return x
	case 5:
		return -x
	case 6:
		return y
	default:
		return -y
	}
}

func (n *Noise) Noise2(x, y float64) float64 {
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	xf, yf := x-float64(xi), y-float64(yi)
	u, v := fade(xf), fade(yf)
	xi &= 255
	yi &= 255
	aa := n.perm[int(n.perm[xi])+yi]
	ab := n.perm[int(n.perm[xi])+yi+1]
	ba := n.perm[int(n.perm[xi+1])+yi]
	bb := n.perm[int(n.perm[xi+1])+yi+1]
	x1 := lerp64(grad2(aa, xf, yf), grad2(ba, xf-1, yf), u)
	x2 := lerp64(grad2(ab, xf, yf-1), grad2(bb, xf-1, yf-1), u)
	return lerp64(x1, x2, v)
}

func (n *Noise) Noise3(x, y, z float64) float64 {
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	zi := int(math.Floor(z))
	xf, yf, zf := x-float64(xi), y-float64(yi), z-float64(zi)
	u, v, w := fade(xf), fade(yf), fade(zf)
	xi &= 255
	yi &= 255
	zi &= 255
	a := int(n.perm[xi]) + yi
	aa := int(n.perm[a]) + zi
	ab := int(n.perm[a+1]) + zi
	b := int(n.perm[xi+1]) + yi
	ba := int(n.perm[b]) + zi
	bb := int(n.perm[b+1]) + zi
	grad := func(h uint8, x, y, z float64) float64 {
		switch h & 15 {
		case 0:
			return x + y
		case 1:
			return -x + y
		case 2:
			return x - y
		case 3:
			return -x - y
		case 4:
			return x + z
		case 5:
			return -x + z
		case 6:
			return x - z
		case 7:
			return -x - z
		case 8:
			return y + z
		case 9:
			return -y + z
		case 10:
			return y - z
		case 11:
			return -y - z
		case 12:
			return x + y
		case 13:
			return -x + y
		case 14:
			return -y + z
		default:
			return -y - z
		}
	}
	x1 := lerp64(grad(n.perm[aa], xf, yf, zf), grad(n.perm[ba], xf-1, yf, zf), u)
	x2 := lerp64(grad(n.perm[ab], xf, yf-1, zf), grad(n.perm[bb], xf-1, yf-1, zf), u)
	y1 := lerp64(x1, x2, v)
	x3 := lerp64(grad(n.perm[aa+1], xf, yf, zf-1), grad(n.perm[ba+1], xf-1, yf, zf-1), u)
	x4 := lerp64(grad(n.perm[ab+1], xf, yf-1, zf-1), grad(n.perm[bb+1], xf-1, yf-1, zf-1), u)
	y2 := lerp64(x3, x4, v)
	return lerp64(y1, y2, w)
}

func (n *Noise) Fbm2(x, y float64, octaves int, lacunarity, gain float64) float64 {
	amp, freq, sum, norm := 1.0, 1.0, 0.0, 0.0
	for i := 0; i < octaves; i++ {
		sum += n.Noise2(x*freq, y*freq) * amp
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0
	}
	return sum / norm
}

func (n *Noise) Fbm3(x, y, z float64, octaves int, lacunarity, gain float64) float64 {
	amp, freq, sum, norm := 1.0, 1.0, 0.0, 0.0
	for i := 0; i < octaves; i++ {
		sum += n.Noise3(x*freq, y*freq, z*freq) * amp
		norm += amp
		amp *= gain
		freq *= lacunarity
	}
	if norm == 0 {
		return 0
	}
	return sum / norm
}

func (n *Noise) Ridged2(x, y float64, octaves int) float64 {
	amp, freq, sum, norm := 1.0, 1.0, 0.0, 0.0
	for i := 0; i < octaves; i++ {
		v := 1 - math.Abs(n.Noise2(x*freq, y*freq))
		sum += v * v * amp
		norm += amp
		amp *= 0.5
		freq *= 2.05
	}
	return sum / norm
}

func (n *Noise) Chance(x, y, z int, numerator, denominator uint64) bool {
	return n.Hash3(x, y, z)%denominator < numerator
}

func (n *Noise) Hash2(x, y int) uint64    { return n.hash2(x, y) }
func (n *Noise) Hash3(x, y, z int) uint64 { return n.hash3(x, y, z) }

func (n *Noise) Unit2(x, y int) float64 {
	return float64(n.hash2(x, y)>>11) / float64(1<<53)
}
func (n *Noise) Unit3(x, y, z int) float64 {
	return float64(n.hash3(x, y, z)>>11) / float64(1<<53)
}
