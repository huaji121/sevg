package entity

import (
	"math"
	"math/rand"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/physics"
	"sevg/internal/world"
)

type Kind uint8

const (
	Pig Kind = iota
	Sheep
)

func (k Kind) Name() string {
	if k == Sheep {
		return "Sheep"
	}
	return "Pig"
}

type Box struct {
	Center mathx.Vec3
	Size   mathx.Vec3
	Color  [4]float32
	// Block, when non-air, makes this box use the real block atlas texture
	// instead of a flat palette colour.
	Block core.BlockID
}

type Mob struct {
	ID          int
	Kind        Kind
	Body        physics.Body
	Health      float32
	MaxHealth   float32
	Yaw         float32
	Dead        bool
	HitFlash    float32
	WanderTimer float32
	WanderYaw   float32
	JumpTimer   float32
	OnGround    bool
	// WalkPhase advances with horizontal speed and drives the leg/head
	// animation, mirroring the player model's gait cycle.
	WalkPhase float32
}

// mobBody returns the AABB dimensions for a kind, matching Minecraft hitboxes
// (pig 0.9x0.9, sheep 0.9x1.3).
func mobBody(k Kind) (width, height float32) {
	if k == Sheep {
		return 0.9, 1.3
	}
	return 0.9, 0.9
}

func (m *Mob) Forward() mathx.Vec3 {
	return mathx.Vec3{X: float32(math.Sin(float64(m.Yaw))), Z: -float32(math.Cos(float64(m.Yaw)))}
}

func (m *Mob) Right() mathx.Vec3 {
	return mathx.Vec3{X: float32(math.Cos(float64(m.Yaw))), Z: float32(math.Sin(float64(m.Yaw)))}
}

func (m *Mob) Bounds() (mathx.Vec3, mathx.Vec3) { return m.Body.Min(), m.Body.Max() }

type DropVisual uint8

const (
	DropVisualCube DropVisual = iota
	DropVisualBlock
	DropVisualGlass
	DropVisualPlant
	DropVisualSword
	DropVisualPickaxe
)

type ItemDrop struct {
	ID          int
	Item        core.ItemID
	Count       int
	Body        physics.Body
	Age         float32
	PickupDelay float32
	Bob         float32
	Dead        bool
}

func (d *ItemDrop) Boxes() []Box {
	c := d.Body.Pos.Add(mathx.Vec3{Y: 0.18})
	col := [4]float32{0.8, 0.8, 0.8, 1}
	block := core.Air
	if d.Item.Valid() {
		def := d.Item.Def()
		col = [4]float32{def.Color[0], def.Color[1], def.Color[2], 1}
		if def.Kind == core.KindBlock {
			block = def.Block
		}
	}
	return []Box{{Center: c, Size: mathx.Vec3{X: 0.28, Y: 0.28, Z: 0.28}, Color: col, Block: block}}
}

func (d *ItemDrop) Visual() DropVisual {
	return VisualForItem(d.Item)
}

func VisualForItem(item core.ItemID) DropVisual {
	if !item.Valid() {
		return DropVisualCube
	}
	def := item.Def()
	if def.Kind == core.KindBlock {
		switch {
		case def.Block == core.Air:
			return DropVisualCube
		case def.Block.Cross():
			return DropVisualPlant
		case def.Block == core.Glass || def.Block == core.Ice:
			return DropVisualGlass
		default:
			return DropVisualBlock
		}
	}
	switch item {
	case core.ItemWoodSword, core.ItemIronSword:
		return DropVisualSword
	case core.ItemStonePickaxe, core.ItemIronPickaxe:
		return DropVisualPickaxe
	default:
		return DropVisualCube
	}
}

func (d *ItemDrop) Transparent() bool {
	return d.Visual() == DropVisualGlass
}

type Manager struct {
	Mobs        []*Mob
	Items       []*ItemDrop
	Falling     []*FallingBlock
	Projectiles []*Projectile
	// ProjectileImpacts collects the world positions where projectiles hit,
	// so the renderer can spawn impact particles.
	ProjectileImpacts []mathx.Vec3
	PickupEvents      int
	rng               *rand.Rand
	nextID            int
	spawnTimer        float32
}

func NewManager(seed int64) *Manager {
	return &Manager{rng: rand.New(rand.NewSource(seed + 909))}
}

func (m *Manager) SpawnInitial(w *world.World, center mathx.Vec3) {
	for i := 0; i < 7; i++ {
		m.trySpawn(w, center, 8, 30)
	}
}

// SpawnAt creates a mob at an exact position (used by tests and the world
// population code).
func (m *Manager) SpawnAt(kind Kind, pos mathx.Vec3) *Mob {
	mob := &Mob{
		ID:        m.nextID,
		Kind:      kind,
		Health:    mobHealth(kind),
		MaxHealth: mobHealth(kind),
		Yaw:       m.rng.Float32() * mathx.Radians(360),
		WanderYaw: m.rng.Float32() * mathx.Radians(360),
	}
	m.nextID++
	w, h := mobBody(kind)
	mob.Body = physics.Body{Pos: pos, Width: w, Height: h}
	m.Mobs = append(m.Mobs, mob)
	return mob
}

func (m *Manager) trySpawn(w *world.World, center mathx.Vec3, minR, maxR float32) bool {
	for attempt := 0; attempt < 14; attempt++ {
		ang := m.rng.Float64() * math.Pi * 2
		dist := minR + m.rng.Float32()*(maxR-minR)
		x := int(math.Floor(float64(center.X + float32(math.Cos(ang))*dist)))
		z := int(math.Floor(float64(center.Z + float32(math.Sin(ang))*dist)))
		if !w.HasChunk(world.ChunkOf(x, z)) {
			continue
		}
		h := w.SurfaceHeight(x, z)
		above := w.GetBlock(x, h+1, z)
		above2 := w.GetBlock(x, h+2, z)
		if above != core.Air || above2 != core.Air {
			continue
		}
		if !w.GetBlock(x, h, z).Solid() {
			continue
		}
		kind := Pig
		if m.rng.Intn(2) == 0 {
			kind = Sheep
		}
		mob := &Mob{
			ID:        m.nextID,
			Kind:      kind,
			Health:    mobHealth(kind),
			MaxHealth: mobHealth(kind),
			Yaw:       m.rng.Float32() * mathx.Radians(360),
			WanderYaw: m.rng.Float32() * mathx.Radians(360),
		}
		m.nextID++
		bw, bh := mobBody(kind)
		mob.Body = physics.Body{
			Pos:    mathx.Vec3{X: float32(x) + 0.5, Y: float32(h) + 1.02, Z: float32(z) + 0.5},
			Width:  bw,
			Height: bh,
		}
		m.Mobs = append(m.Mobs, mob)
		_ = minR
		return true
	}
	return false
}

func mobHealth(k Kind) float32 {
	if k == Sheep {
		return 8
	}
	return 10
}

func (m *Manager) Update(dt float32, w *world.World, playerPos mathx.Vec3, inv *core.Inventory) {
	if dt > 0.05 {
		dt = 0.05
	}
	for _, mob := range m.Mobs {
		if mob.Dead {
			continue
		}
		if mob.HitFlash > 0 {
			mob.HitFlash -= dt
		}
		mob.WanderTimer -= dt
		if mob.WanderTimer <= 0 {
			mob.WanderTimer = 2 + m.rng.Float32()*4
			if m.rng.Float32() < 0.3 {
				mob.WanderYaw = mob.Yaw + mathx.Radians(180)
			} else {
				mob.WanderYaw = m.rng.Float32() * mathx.Radians(360)
			}
		}
		mob.Yaw += angleDelta(mob.Yaw, mob.WanderYaw) * mathx.Clamp(dt*5, 0, 1)
		fwd := mob.Forward()
		walk := float32(1.6)
		if mob.Body.Pos.Sub(playerPos).Length() > 48 {
			walk = 0
		}
		mob.Body.Vel.X += (fwd.X*walk - mob.Body.Vel.X) * mathx.Clamp(dt*8, 0, 1)
		mob.Body.Vel.Z += (fwd.Z*walk - mob.Body.Vel.Z) * mathx.Clamp(dt*8, 0, 1)

		inWater := w.GetBlock(int(math.Floor(float64(mob.Body.Pos.X))), int(math.Floor(float64(mob.Body.Pos.Y))), int(math.Floor(float64(mob.Body.Pos.Z)))) == core.Water
		if inWater {
			mob.Body.Vel.Y -= 5 * dt
			mob.Body.Vel.Y *= 0.92
			mob.Body.Vel.Y += 2.8 * dt
		} else {
			mob.Body.Vel.Y -= 25 * dt
			if mob.Body.Vel.Y < -40 {
				mob.Body.Vel.Y = -40
			}
		}
		mob.JumpTimer -= dt
		hx, _, hz := mob.Body.Move(w, mob.Body.Vel.Mul(dt))
		if (hx || hz) && mob.Body.Grounded(w) && mob.JumpTimer <= 0 {
			mob.Body.Vel.Y = 7.6
			mob.JumpTimer = 0.8
		}
		if mob.Body.Pos.Y < -5 {
			mob.Dead = true
		}
		hs := float32(math.Hypot(float64(mob.Body.Vel.X), float64(mob.Body.Vel.Z)))
		if mob.Body.OnGround && hs > 0.15 {
			mob.WalkPhase += dt * hs * 2.4
		}
	}

	m.UpdateFalling(w, dt)
	m.UpdateProjectiles(w, dt)

	// Collect pickups and expire old drops.
	for _, d := range m.Items {
		if d.Dead {
			continue
		}
		d.Age += dt
		d.Body.Vel.Y -= 18 * dt
		if d.Body.Vel.Y < -30 {
			d.Body.Vel.Y = -30
		}
		d.Body.Move(w, d.Body.Vel.Mul(dt))
		if d.Body.OnGround {
			d.Body.Vel.X *= 0.7
			d.Body.Vel.Z *= 0.7
		}
		if d.Body.Pos.Y > 300 {
			d.Dead = true
		}
		if d.Age > d.PickupDelay {
			delta := d.Body.Pos.Add(mathx.Vec3{Y: 0.2}).Sub(playerPos)
			if delta.Length() < 2.2 {
				before := d.Count
				left := inv.Add(d.Item, d.Count)
				if left < before {
					m.PickupEvents++
				}
				if left <= 0 {
					d.Dead = true
				} else {
					d.Count = left
				}
			}
		}
		if d.Age > 300 {
			d.Dead = true
		}
	}
	m.mergeItems()

	// Compact slices and spawn new animals while the herd is small.
	live := m.Mobs[:0]
	for _, mob := range m.Mobs {
		if !mob.Dead {
			live = append(live, mob)
		}
	}
	m.Mobs = live
	li := m.Items[:0]
	for _, d := range m.Items {
		if !d.Dead {
			li = append(li, d)
		}
	}
	m.Items = li
	m.spawnTimer -= dt
	if m.spawnTimer <= 0 {
		m.spawnTimer = 6
		if len(m.Mobs) < 9 {
			m.trySpawn(w, playerPos, 22, 44)
		}
	}
}

func angleDelta(a, b float32) float32 {
	d := b - a
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}

func (m *Manager) Drop(item core.ItemID, count int, pos mathx.Vec3, vel mathx.Vec3) {
	m.drop(item, count, pos, vel, 0.45)
}

// DropPlayer creates a drop thrown by the player. It cannot be picked up for a
// short time so Q-dropped items do not immediately return to the inventory.
func (m *Manager) DropPlayer(item core.ItemID, count int, pos mathx.Vec3, vel mathx.Vec3) {
	m.drop(item, count, pos, vel, 1.25)
}

// dropMergeRadius is the horizontal distance within which dropped items of the
// same type combine into a single stack, matching Minecraft's item merging.
const dropMergeRadius = float32(0.75)

// dropStackLimit is the maximum count a single dropped stack may hold.
func dropStackLimit(item core.ItemID) int {
	if !item.Valid() {
		return core.StackLimit
	}
	limit := item.Def().MaxStack
	if limit <= 0 {
		limit = core.StackLimit
	}
	return limit
}

func (m *Manager) drop(item core.ItemID, count int, pos mathx.Vec3, vel mathx.Vec3, pickupDelay float32) {
	if item == core.ItemNone || count <= 0 {
		return
	}
	// Fold into a nearby stack of the same item so mined drops pile up instead
	// of spawning one entity per item. The merged stack keeps the longest
	// pickup delay so thrown items still cannot be grabbed immediately.
	if limit := dropStackLimit(item); limit > 1 {
		mergeSq := dropMergeRadius * dropMergeRadius
		for _, d := range m.Items {
			if d.Dead || d.Item != item || d.Count >= limit {
				continue
			}
			if d.Body.Pos.Sub(pos).LengthSq() > mergeSq {
				continue
			}
			take := count
			if room := limit - d.Count; take > room {
				take = room
			}
			d.Count += take
			if d.PickupDelay < pickupDelay {
				d.PickupDelay = pickupDelay
			}
			count -= take
			if count <= 0 {
				return
			}
		}
	}
	d := &ItemDrop{ID: m.nextID, Item: item, Count: count, PickupDelay: pickupDelay, Body: physics.Body{Pos: pos, Vel: vel, Width: 0.25, Height: 0.25}}
	m.nextID++
	m.Items = append(m.Items, d)
}

// mergeItems combines nearby drops of the same type into stacks up to the
// item's stack limit. Items that drifted apart after spawning still merge once
// they settle next to each other.
func (m *Manager) mergeItems() {
	mergeSq := dropMergeRadius * dropMergeRadius
	for i := 0; i < len(m.Items); i++ {
		a := m.Items[i]
		if a.Dead {
			continue
		}
		limit := dropStackLimit(a.Item)
		if a.Count >= limit {
			continue
		}
		for j := i + 1; j < len(m.Items); j++ {
			b := m.Items[j]
			if b.Dead || b.Item != a.Item {
				continue
			}
			if a.Body.Pos.Sub(b.Body.Pos).LengthSq() > mergeSq {
				continue
			}
			room := limit - a.Count
			if room <= 0 {
				break
			}
			take := b.Count
			if take > room {
				take = room
			}
			a.Count += take
			if a.PickupDelay < b.PickupDelay {
				a.PickupDelay = b.PickupDelay
			}
			b.Count -= take
			if b.Count <= 0 {
				b.Dead = true
			}
		}
	}
}

// PickMob finds the nearest mob along a ray without damaging it.
func (m *Manager) PickMob(origin, dir mathx.Vec3, maxDist float32) (*Mob, float32, bool) {
	dir = dir.Normalize()
	bestDist := maxDist
	var best *Mob
	for _, mob := range m.Mobs {
		if mob.Dead {
			continue
		}
		min, max := mob.Bounds()
		if d, ok := world.RayAABB(origin, dir, min, max); ok && d < bestDist {
			bestDist = d
			best = mob
		}
	}
	if best == nil {
		return nil, 0, false
	}
	return best, bestDist, true
}

// DamageMob applies damage, knockback and death drops.
func (m *Manager) DamageMob(mob *Mob, damage float32, playerPos mathx.Vec3) {
	if mob == nil || mob.Dead {
		return
	}
	mob.Health -= damage
	mob.HitFlash = 0.25
	kb := mob.Body.Pos.Sub(playerPos).Normalize().Mul(6)
	mob.Body.Vel.X = kb.X
	mob.Body.Vel.Z = kb.Z
	mob.Body.Vel.Y = 4
	if mob.Health <= 0 {
		mob.Dead = true
		drop := core.ItemWool
		n := 2 + m.rng.Intn(2)
		if mob.Kind == Pig {
			drop = core.ItemApple
			n = 1 + m.rng.Intn(2)
		}
		m.Drop(drop, n, mob.Body.Pos.Add(mathx.Vec3{Y: 0.5}), mathx.Vec3{X: m.rng.Float32()*2 - 1, Y: 4, Z: m.rng.Float32()*2 - 1})
	}
}

func (m *Manager) MobCount() int  { return len(m.Mobs) }
func (m *Manager) ItemCount() int { return len(m.Items) }
