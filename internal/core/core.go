// Package core defines the block/item registries and inventory shared by the game.
package core

import "fmt"

type BlockID uint8

const (
	Air BlockID = iota
	Grass
	Dirt
	Stone
	Cobble
	Sand
	Water
	Log
	Leaves
	Planks
	CoalOre
	IronOre
	Snow
	SnowyGrass
	Cactus
	Glass
	Bedrock
	Gravel
	Flower
	TallGrass
	Brick
	StoneBrick
	Ice
	Obsidian
	Wool
	BlockCount
)

// Atlas tile indices. The atlas is generated procedurally at startup.
const (
	TileGrassTop = iota
	TileGrassSide
	TileDirt
	TileStone
	TileCobble
	TileSand
	TileWater
	TileLogSide
	TileLogTop
	TileLeaves
	TilePlanks
	TileCoalOre
	TileIronOre
	TileSnow
	TileSnowSide
	TileCactusSide
	TileCactusTop
	TileGlass
	TileBedrock
	TileGravel
	TileFlower
	TileTallGrass
	TileBrick
	TileStoneBrick
	TileIce
	TileObsidian
	TileWool
	TileWhite
	TilePig
	TileDark
	TileRed
	TileGold
	TileGray
	TileBlack
	TileBrown
	TilePink
	TileGreen
	TileTileCount
)

// Procedural block atlas geometry. Cell includes a 4px gutter so nearest
// sampling never bleeds between tiles.
const (
	AtlasCols   = 8
	AtlasRows   = 8
	AtlasTilePx = 16
	AtlasPad    = 4
	AtlasCell   = AtlasTilePx + AtlasPad*2
	AtlasW      = AtlasCols * AtlasCell
	AtlasH      = AtlasRows * AtlasCell
	AtlasMips   = 4
)

type ItemID uint16

const (
	ItemNone ItemID = iota
	ItemGrass
	ItemDirt
	ItemStone
	ItemCobble
	ItemSand
	ItemLog
	ItemLeaves
	ItemPlanks
	ItemCoalOre
	ItemIronOre
	ItemSnow
	ItemCactus
	ItemGlass
	ItemGravel
	ItemFlower
	ItemTallGrass
	ItemBrick
	ItemStoneBrick
	ItemIce
	ItemObsidian
	ItemStick
	ItemApple
	ItemCoal
	ItemIronIngot
	ItemWoodSword
	ItemStonePickaxe
	ItemIronSword
	ItemIronPickaxe
	ItemWool
	ItemCount
)

type ItemKind uint8

const (
	KindBlock ItemKind = iota
	KindMaterial
	KindTool
	KindFood
)

type BlockDef struct {
	Name     string
	Solid    bool
	Opaque   bool
	Liquid   bool
	Cross    bool
	Hardness float32
	Tiles    [6]uint8 // +X, -X, +Y, -Y, +Z, -Z
	Drop     ItemID
}

var BlockDefs = [BlockCount]BlockDef{
	Air:        {Name: "Air", Tiles: [6]uint8{}},
	Grass:      {Name: "Grass Block", Solid: true, Opaque: true, Hardness: 0.65, Tiles: [6]uint8{TileGrassSide, TileGrassSide, TileGrassTop, TileDirt, TileGrassSide, TileGrassSide}, Drop: ItemDirt},
	Dirt:       {Name: "Dirt", Solid: true, Opaque: true, Hardness: 0.55, Tiles: [6]uint8{TileDirt, TileDirt, TileDirt, TileDirt, TileDirt, TileDirt}},
	Stone:      {Name: "Stone", Solid: true, Opaque: true, Hardness: 1.6, Tiles: [6]uint8{TileStone, TileStone, TileStone, TileStone, TileStone, TileStone}, Drop: ItemCobble},
	Cobble:     {Name: "Cobblestone", Solid: true, Opaque: true, Hardness: 1.8, Tiles: [6]uint8{TileCobble, TileCobble, TileCobble, TileCobble, TileCobble, TileCobble}},
	Sand:       {Name: "Sand", Solid: true, Opaque: true, Hardness: 0.5, Tiles: [6]uint8{TileSand, TileSand, TileSand, TileSand, TileSand, TileSand}},
	Water:      {Name: "Water", Liquid: true, Opaque: false, Hardness: -1, Tiles: [6]uint8{TileWater, TileWater, TileWater, TileWater, TileWater, TileWater}},
	Log:        {Name: "Oak Log", Solid: true, Opaque: true, Hardness: 1.2, Tiles: [6]uint8{TileLogSide, TileLogSide, TileLogTop, TileLogTop, TileLogSide, TileLogSide}},
	Leaves:     {Name: "Leaves", Solid: true, Opaque: true, Hardness: 0.25, Tiles: [6]uint8{TileLeaves, TileLeaves, TileLeaves, TileLeaves, TileLeaves, TileLeaves}, Drop: ItemNone},
	Planks:     {Name: "Oak Planks", Solid: true, Opaque: true, Hardness: 1.0, Tiles: [6]uint8{TilePlanks, TilePlanks, TilePlanks, TilePlanks, TilePlanks, TilePlanks}},
	CoalOre:    {Name: "Coal Ore", Solid: true, Opaque: true, Hardness: 2.2, Tiles: [6]uint8{TileCoalOre, TileCoalOre, TileCoalOre, TileCoalOre, TileCoalOre, TileCoalOre}, Drop: ItemCoal},
	IronOre:    {Name: "Iron Ore", Solid: true, Opaque: true, Hardness: 2.8, Tiles: [6]uint8{TileIronOre, TileIronOre, TileIronOre, TileIronOre, TileIronOre, TileIronOre}, Drop: ItemIronIngot},
	Snow:       {Name: "Snow", Solid: true, Opaque: true, Hardness: 0.3, Tiles: [6]uint8{TileSnow, TileSnow, TileSnow, TileSnow, TileSnow, TileSnow}},
	SnowyGrass: {Name: "Snowy Grass", Solid: true, Opaque: true, Hardness: 0.65, Tiles: [6]uint8{TileSnowSide, TileSnowSide, TileSnow, TileDirt, TileSnowSide, TileSnowSide}, Drop: ItemDirt},
	Cactus:     {Name: "Cactus", Solid: true, Opaque: true, Hardness: 0.55, Tiles: [6]uint8{TileCactusSide, TileCactusSide, TileCactusTop, TileCactusTop, TileCactusSide, TileCactusSide}},
	Glass:      {Name: "Glass", Solid: true, Opaque: false, Hardness: 0.4, Tiles: [6]uint8{TileGlass, TileGlass, TileGlass, TileGlass, TileGlass, TileGlass}, Drop: ItemNone},
	Bedrock:    {Name: "Bedrock", Solid: true, Opaque: true, Hardness: -1, Tiles: [6]uint8{TileBedrock, TileBedrock, TileBedrock, TileBedrock, TileBedrock, TileBedrock}},
	Gravel:     {Name: "Gravel", Solid: true, Opaque: true, Hardness: 0.7, Tiles: [6]uint8{TileGravel, TileGravel, TileGravel, TileGravel, TileGravel, TileGravel}},
	Flower:     {Name: "Flower", Cross: true, Hardness: 0.05, Tiles: [6]uint8{TileFlower, TileFlower, TileFlower, TileFlower, TileFlower, TileFlower}},
	TallGrass:  {Name: "Tall Grass", Cross: true, Hardness: 0.05, Tiles: [6]uint8{TileTallGrass, TileTallGrass, TileTallGrass, TileTallGrass, TileTallGrass, TileTallGrass}},
	Brick:      {Name: "Bricks", Solid: true, Opaque: true, Hardness: 2.0, Tiles: [6]uint8{TileBrick, TileBrick, TileBrick, TileBrick, TileBrick, TileBrick}},
	StoneBrick: {Name: "Stone Bricks", Solid: true, Opaque: true, Hardness: 2.0, Tiles: [6]uint8{TileStoneBrick, TileStoneBrick, TileStoneBrick, TileStoneBrick, TileStoneBrick, TileStoneBrick}},
	Ice:        {Name: "Ice", Solid: true, Opaque: false, Hardness: 0.5, Tiles: [6]uint8{TileIce, TileIce, TileIce, TileIce, TileIce, TileIce}, Drop: ItemNone},
	Obsidian:   {Name: "Obsidian", Solid: true, Opaque: true, Hardness: 8.0, Tiles: [6]uint8{TileObsidian, TileObsidian, TileObsidian, TileObsidian, TileObsidian, TileObsidian}},
	Wool:       {Name: "Wool", Solid: true, Opaque: true, Hardness: 0.6, Tiles: [6]uint8{TileWool, TileWool, TileWool, TileWool, TileWool, TileWool}},
}

func (b BlockID) Valid() bool { return b < BlockCount }
func (b BlockID) Def() *BlockDef {
	if !b.Valid() {
		return &BlockDefs[Air]
	}
	return &BlockDefs[b]
}
func (b BlockID) Name() string        { return b.Def().Name }
func (b BlockID) IsAir() bool         { return b == Air }
func (b BlockID) Solid() bool         { return b.Def().Solid }
func (b BlockID) Opaque() bool        { return b.Def().Opaque }
func (b BlockID) Liquid() bool        { return b.Def().Liquid }
func (b BlockID) Cross() bool         { return b.Def().Cross }
func (b BlockID) Hardness() float32   { return b.Def().Hardness }
func (b BlockID) Tile(face int) uint8 { return b.Def().Tiles[face&5] }

type ItemDef struct {
	Name     string
	Kind     ItemKind
	Block    BlockID
	MaxStack int
	Color    [3]float32
	Mining   float32 // mining speed multiplier while held
	Damage   float32
	Food     float32
}

var blockItems [BlockCount]ItemID

var ItemDefs = [ItemCount]ItemDef{
	ItemNone:         {},
	ItemGrass:        {Name: "Grass Block", Kind: KindBlock, Block: Grass, MaxStack: 99, Color: [3]float32{0.42, 0.68, 0.30}},
	ItemDirt:         {Name: "Dirt", Kind: KindBlock, Block: Dirt, MaxStack: 99, Color: [3]float32{0.48, 0.33, 0.22}},
	ItemStone:        {Name: "Stone", Kind: KindBlock, Block: Stone, MaxStack: 99, Color: [3]float32{0.52, 0.52, 0.54}},
	ItemCobble:       {Name: "Cobblestone", Kind: KindBlock, Block: Cobble, MaxStack: 99, Color: [3]float32{0.46, 0.46, 0.48}},
	ItemSand:         {Name: "Sand", Kind: KindBlock, Block: Sand, MaxStack: 99, Color: [3]float32{0.83, 0.76, 0.52}},
	ItemLog:          {Name: "Oak Log", Kind: KindBlock, Block: Log, MaxStack: 99, Color: [3]float32{0.45, 0.31, 0.17}},
	ItemLeaves:       {Name: "Leaves", Kind: KindBlock, Block: Leaves, MaxStack: 99, Color: [3]float32{0.24, 0.52, 0.20}},
	ItemPlanks:       {Name: "Oak Planks", Kind: KindBlock, Block: Planks, MaxStack: 99, Color: [3]float32{0.66, 0.48, 0.27}},
	ItemCoalOre:      {Name: "Coal Ore", Kind: KindBlock, Block: CoalOre, MaxStack: 99, Color: [3]float32{0.30, 0.30, 0.32}},
	ItemIronOre:      {Name: "Iron Ore", Kind: KindBlock, Block: IronOre, MaxStack: 99, Color: [3]float32{0.62, 0.55, 0.49}},
	ItemSnow:         {Name: "Snow", Kind: KindBlock, Block: Snow, MaxStack: 99, Color: [3]float32{0.92, 0.95, 0.98}},
	ItemCactus:       {Name: "Cactus", Kind: KindBlock, Block: Cactus, MaxStack: 99, Color: [3]float32{0.24, 0.55, 0.25}},
	ItemGlass:        {Name: "Glass", Kind: KindBlock, Block: Glass, MaxStack: 99, Color: [3]float32{0.72, 0.85, 0.88}},
	ItemGravel:       {Name: "Gravel", Kind: KindBlock, Block: Gravel, MaxStack: 99, Color: [3]float32{0.48, 0.45, 0.43}},
	ItemFlower:       {Name: "Flower", Kind: KindBlock, Block: Flower, MaxStack: 99, Color: [3]float32{0.92, 0.35, 0.55}},
	ItemTallGrass:    {Name: "Tall Grass", Kind: KindBlock, Block: TallGrass, MaxStack: 99, Color: [3]float32{0.38, 0.62, 0.28}},
	ItemBrick:        {Name: "Bricks", Kind: KindBlock, Block: Brick, MaxStack: 99, Color: [3]float32{0.66, 0.30, 0.24}},
	ItemStoneBrick:   {Name: "Stone Bricks", Kind: KindBlock, Block: StoneBrick, MaxStack: 99, Color: [3]float32{0.50, 0.50, 0.48}},
	ItemIce:          {Name: "Ice", Kind: KindBlock, Block: Ice, MaxStack: 99, Color: [3]float32{0.62, 0.82, 0.95}},
	ItemObsidian:     {Name: "Obsidian", Kind: KindBlock, Block: Obsidian, MaxStack: 99, Color: [3]float32{0.16, 0.10, 0.23}},
	ItemStick:        {Name: "Stick", Kind: KindMaterial, MaxStack: 99, Color: [3]float32{0.55, 0.40, 0.22}},
	ItemApple:        {Name: "Apple", Kind: KindFood, MaxStack: 16, Color: [3]float32{0.82, 0.12, 0.13}, Food: 4},
	ItemCoal:         {Name: "Coal", Kind: KindMaterial, MaxStack: 99, Color: [3]float32{0.12, 0.12, 0.13}},
	ItemIronIngot:    {Name: "Iron Ingot", Kind: KindMaterial, MaxStack: 99, Color: [3]float32{0.78, 0.78, 0.82}},
	ItemWoodSword:    {Name: "Wooden Sword", Kind: KindTool, MaxStack: 1, Color: [3]float32{0.62, 0.45, 0.24}, Damage: 4, Mining: 1.2},
	ItemStonePickaxe: {Name: "Stone Pickaxe", Kind: KindTool, MaxStack: 1, Color: [3]float32{0.53, 0.53, 0.56}, Damage: 2, Mining: 2.4},
	ItemIronSword:    {Name: "Iron Sword", Kind: KindTool, MaxStack: 1, Color: [3]float32{0.82, 0.82, 0.86}, Damage: 7, Mining: 1.2},
	ItemIronPickaxe:  {Name: "Iron Pickaxe", Kind: KindTool, MaxStack: 1, Color: [3]float32{0.84, 0.84, 0.88}, Damage: 3, Mining: 4.0},
	ItemWool:         {Name: "Wool", Kind: KindBlock, Block: Wool, MaxStack: 99, Color: [3]float32{0.92, 0.92, 0.90}},
}

func init() {
	blockItems[Air] = ItemNone
	for id := BlockID(1); id < BlockCount; id++ {
		switch id {
		case Water, Bedrock, Leaves, Flower, TallGrass:
			continue
		case Grass:
			blockItems[id] = ItemGrass
		case Dirt:
			blockItems[id] = ItemDirt
		case Stone:
			blockItems[id] = ItemStone
		case Cobble:
			blockItems[id] = ItemCobble
		case Sand:
			blockItems[id] = ItemSand
		case Log:
			blockItems[id] = ItemLog
		case Planks:
			blockItems[id] = ItemPlanks
		case CoalOre:
			blockItems[id] = ItemCoalOre
		case IronOre:
			blockItems[id] = ItemIronOre
		case Snow:
			blockItems[id] = ItemSnow
		case SnowyGrass:
			blockItems[id] = ItemSnow
		case Cactus:
			blockItems[id] = ItemCactus
		case Glass:
			blockItems[id] = ItemGlass
		case Gravel:
			blockItems[id] = ItemGravel
		case Brick:
			blockItems[id] = ItemBrick
		case StoneBrick:
			blockItems[id] = ItemStoneBrick
		case Ice:
			blockItems[id] = ItemIce
		case Obsidian:
			blockItems[id] = ItemObsidian
		case Wool:
			blockItems[id] = ItemWool
		}
	}
	// Flowers and tall grass are picked up and placeable.
	blockItems[Flower] = ItemFlower
	blockItems[TallGrass] = ItemTallGrass
}

func ItemForBlock(b BlockID) ItemID {
	if b >= BlockCount {
		return ItemNone
	}
	return blockItems[b]
}

func (i ItemID) Valid() bool { return i > ItemNone && i < ItemCount }
func (i ItemID) Def() *ItemDef {
	if !i.Valid() {
		return &ItemDef{Name: "Empty", MaxStack: 1}
	}
	return &ItemDefs[i]
}
func (i ItemID) Name() string { return i.Def().Name }
func (i ItemID) MaxStack() int {
	d := i.Def()
	if d.MaxStack < 1 {
		return 1
	}
	return d.MaxStack
}
func (i ItemID) IsBlock() bool { return i.Valid() && i.Def().Kind == KindBlock && i.Def().Block != Air }

func (i ItemID) String() string { return fmt.Sprintf("%s(%d)", i.Name(), i) }

type Stack struct {
	Item  ItemID
	Count int
}

func (s Stack) Empty() bool { return s.Item == ItemNone || s.Count <= 0 }
func (s Stack) CanMerge(item ItemID) bool {
	return !s.Empty() && s.Item == item && s.Count < item.MaxStack()
}

const (
	InventorySlots = 36
	HotbarSlots    = 9
)

type Inventory struct {
	Slots    [InventorySlots]Stack
	Selected int
}

func (inv *Inventory) SelectedStack() *Stack { return &inv.Slots[inv.Selected] }
func (inv *Inventory) Select(i int) {
	if i >= 0 && i < HotbarSlots {
		inv.Selected = i
	}
}
func (inv *Inventory) Scroll(delta int) {
	inv.Selected = (inv.Selected + delta) % HotbarSlots
	if inv.Selected < 0 {
		inv.Selected += HotbarSlots
	}
}

// Add inserts items and returns the number that did not fit.
func (inv *Inventory) Add(item ItemID, count int) int {
	if item == ItemNone || count <= 0 {
		return count
	}
	max := item.MaxStack()
	for pass := 0; pass < 2 && count > 0; pass++ {
		for i := 0; i < InventorySlots && count > 0; i++ {
			s := &inv.Slots[i]
			if pass == 0 {
				if !s.Empty() && s.Item == item && s.Count < max {
					n := max - s.Count
					if n > count {
						n = count
					}
					s.Count += n
					count -= n
				}
			} else if s.Empty() {
				n := max
				if n > count {
					n = count
				}
				s.Item = item
				s.Count = n
				count -= n
			}
		}
	}
	inv.Normalize()
	return count
}

func (inv *Inventory) Remove(item ItemID, count int) int {
	removed := 0
	for i := 0; i < InventorySlots && count > 0; i++ {
		s := &inv.Slots[i]
		if s.Item == item && s.Count > 0 {
			n := s.Count
			if n > count {
				n = count
			}
			s.Count -= n
			count -= n
			removed += n
			if s.Count == 0 {
				*s = Stack{}
			}
		}
	}
	return removed
}

func (inv *Inventory) TakeSelected(n int) ItemID {
	s := &inv.Slots[inv.Selected]
	if s.Empty() {
		return ItemNone
	}
	item := s.Item
	if n > s.Count {
		n = s.Count
	}
	s.Count -= n
	if s.Count <= 0 {
		*s = Stack{}
	}
	return item
}

func (inv *Inventory) Swap(a, b int) {
	if a < 0 || b < 0 || a >= InventorySlots || b >= InventorySlots || a == b {
		return
	}
	inv.Slots[a], inv.Slots[b] = inv.Slots[b], inv.Slots[a]
}

func (inv *Inventory) Normalize() {
	for i := range inv.Slots {
		if inv.Slots[i].Count <= 0 || inv.Slots[i].Item == ItemNone {
			inv.Slots[i] = Stack{}
		}
	}
	if inv.Selected < 0 || inv.Selected >= HotbarSlots {
		inv.Selected = 0
	}
}

func (inv *Inventory) Count(item ItemID) int {
	n := 0
	for _, s := range inv.Slots {
		if s.Item == item {
			n += s.Count
		}
	}
	return n
}

func (inv *Inventory) FillStarter() {
	inv.Add(ItemPlanks, 32)
	inv.Add(ItemCobble, 24)
	inv.Add(ItemGlass, 12)
	inv.Add(ItemLog, 8)
	inv.Add(ItemStonePickaxe, 1)
	inv.Add(ItemWoodSword, 1)
	inv.Add(ItemApple, 3)
}
