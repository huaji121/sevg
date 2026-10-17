// Package mathx contains the small amount of vector and matrix math needed by SEVG.
package mathx

import "math"

type Vec2 struct{ X, Y float32 }
type Vec3 struct{ X, Y, Z float32 }

func V3(x, y, z float32) Vec3 { return Vec3{x, y, z} }

func (a Vec3) Add(b Vec3) Vec3    { return Vec3{a.X + b.X, a.Y + b.Y, a.Z + b.Z} }
func (a Vec3) Sub(b Vec3) Vec3    { return Vec3{a.X - b.X, a.Y - b.Y, a.Z - b.Z} }
func (a Vec3) Mul(s float32) Vec3 { return Vec3{a.X * s, a.Y * s, a.Z * s} }
func (a Vec3) Div(s float32) Vec3 { return Vec3{a.X / s, a.Y / s, a.Z / s} }
func (a Vec3) Dot(b Vec3) float32 { return a.X*b.X + a.Y*b.Y + a.Z*b.Z }
func (a Vec3) Length() float32    { return float32(math.Sqrt(float64(a.Dot(a)))) }
func (a Vec3) LengthSq() float32  { return a.Dot(a) }
func (a Vec3) Lerp(b Vec3, t float32) Vec3 {
	return Vec3{a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t, a.Z + (b.Z-a.Z)*t}
}
func (a Vec3) Cross(b Vec3) Vec3 {
	return Vec3{a.Y*b.Z - a.Z*b.Y, a.Z*b.X - a.X*b.Z, a.X*b.Y - a.Y*b.X}
}
func (a Vec3) Normalize() Vec3 {
	l := a.Length()
	if l < 1e-6 {
		return Vec3{}
	}
	return a.Div(l)
}
func (a Vec3) Floor() (int, int, int) {
	return int(math.Floor(float64(a.X))), int(math.Floor(float64(a.Y))), int(math.Floor(float64(a.Z)))
}
func (a Vec3) XY() Vec2 { return Vec2{a.X, a.Y} }

func Clamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func ClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func Lerp(a, b, t float32) float32 { return a + (b-a)*t }
func Radians(d float32) float32    { return d * math.Pi / 180 }

// Mat4 is a column-major 4x4 matrix, matching GLSL layout.
type Mat4 [16]float32

func Identity() Mat4 {
	return Mat4{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
}

func (a Mat4) Mul(b Mat4) Mat4 {
	var o Mat4
	for c := 0; c < 4; c++ {
		for r := 0; r < 4; r++ {
			var s float32
			for k := 0; k < 4; k++ {
				s += a[k*4+r] * b[c*4+k]
			}
			o[c*4+r] = s
		}
	}
	return o
}

func Translate(x, y, z float32) Mat4 {
	m := Identity()
	m[12], m[13], m[14] = x, y, z
	return m
}

func Scale(x, y, z float32) Mat4 {
	m := Identity()
	m[0], m[5], m[10] = x, y, z
	return m
}

func RotationY(a float32) Mat4 {
	c, s := float32(math.Cos(float64(a))), float32(math.Sin(float64(a)))
	m := Identity()
	m[0], m[2], m[8], m[10] = c, -s, s, c
	return m
}

func LookAt(eye, center, up Vec3) Mat4 {
	f := center.Sub(eye).Normalize()
	s := f.Cross(up).Normalize()
	u := s.Cross(f)
	m := Identity()
	m[0], m[4], m[8] = s.X, s.Y, s.Z
	m[1], m[5], m[9] = u.X, u.Y, u.Z
	m[2], m[6], m[10] = -f.X, -f.Y, -f.Z
	m[12] = -s.Dot(eye)
	m[13] = -u.Dot(eye)
	m[14] = f.Dot(eye)
	return m
}

// Perspective uses the SDL_GPU/Vulkan convention: z maps to [0, 1].
func Perspective(fovyRad, aspect, near, far float32) Mat4 {
	f := 1 / float32(math.Tan(float64(fovyRad)/2))
	m := Mat4{}
	m[0] = f / aspect
	m[5] = f
	m[10] = far / (near - far)
	m[11] = -1
	m[14] = (far * near) / (near - far)
	return m
}

func (m Mat4) TransformPoint(v Vec3) Vec3 {
	x := m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]
	y := m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]
	z := m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]
	w := m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]
	if w != 0 && w != 1 {
		x, y, z = x/w, y/w, z/w
	}
	return Vec3{x, y, z}
}
