// Package render implements the SDL_GPU voxel renderer.
package render

import (
	"embed"
	"errors"
	"fmt"
	"sort"
	"unsafe"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/ui"
	"sevg/internal/world"
)

//go:embed shaders/*.spv
var shaderFS embed.FS

type Camera struct {
	Pos, Dir, Up mathx.Vec3
	Fov          float32
	Near, Far    float32
	Aspect       float32
}

func (c Camera) ViewProj() mathx.Mat4 {
	proj := mathx.Perspective(c.Fov, c.Aspect, c.Near, c.Far)
	view := mathx.LookAt(c.Pos, c.Pos.Add(c.Dir), c.Up)
	return proj.Mul(view)
}

type WorldUniforms struct {
	FogColor [3]float32
	FogStart float32
	FogEnd   float32
	SunTint  [3]float32
	Clear    [4]float32
}

type Frame struct {
	Camera            Camera
	World             WorldUniforms
	Entity            world.MeshData
	EntityTransparent world.MeshData
	EntityIcon        world.MeshData
	Held              world.MeshData
	HeldTransparent   world.MeshData
	HeldIcon          world.MeshData
	UI                *ui.Canvas
}

type worldUBO struct {
	ViewProj  mathx.Mat4
	FogColor  [4]float32
	FogParams [4]float32
	Cam       [4]float32
	Sun       [4]float32
}

type uiUBO struct {
	Screen [4]float32
}

type gpuMesh struct {
	vb, ib *sdl.GPUBuffer
	count  uint32
}

func (m *gpuMesh) release(d *sdl.GPUDevice) {
	if m.vb != nil {
		d.ReleaseBuffer(m.vb)
	}
	if m.ib != nil {
		d.ReleaseBuffer(m.ib)
	}
	m.vb, m.ib, m.count = nil, nil, 0
}

type gpuChunk struct {
	pos         world.ChunkPos
	opaque      gpuMesh
	transparent gpuMesh
}

type Renderer struct {
	Dev  *sdl.GPUDevice
	Win  *sdl.Window
	W, H int

	swapFormat  sdl.GPUTextureFormat
	depthFormat sdl.GPUTextureFormat
	depth       *sdl.GPUTexture

	atlasTex  *sdl.GPUTexture
	uiTex     *sdl.GPUTexture
	fontAtlas ui.FontAtlas
	sampler   *sdl.GPUSampler

	blockOpaque *sdl.GPUGraphicsPipeline
	blockTrans  *sdl.GPUGraphicsPipeline
	blockView   *sdl.GPUGraphicsPipeline
	uiPipeline  *sdl.GPUGraphicsPipeline

	chunks map[[2]int]*gpuChunk

	uiVB, entVB, entIB         *sdl.GPUBuffer
	entTransVB, entTransIB     *sdl.GPUBuffer
	entIconVB, entIconIB       *sdl.GPUBuffer
	heldVB, heldIB             *sdl.GPUBuffer
	heldTransVB, heldTransIB   *sdl.GPUBuffer
	heldIconVB, heldIconIB     *sdl.GPUBuffer
	uiCap, entCap, entTransCap uint32
	entIconCap, heldCap        uint32
	heldTransCap, heldIconCap  uint32
	entICount, entTransICount  uint32
	entIconICount              uint32
	heldICount                 uint32
	heldTransICount            uint32
	heldIconICount             uint32
	staging                    *sdl.GPUTransferBuffer
	stagingCap                 uint32
}

func New(win *sdl.Window, w, h int, fontAtlas ui.FontAtlas) (*Renderer, error) {
	dev, err := sdl.CreateGPUDevice(sdl.GPU_SHADERFORMAT_SPIRV, true, "")
	if err != nil {
		return nil, fmt.Errorf("create GPU device (Vulkan is required): %w", err)
	}
	if err := dev.ClaimWindow(win); err != nil {
		dev.Destroy()
		return nil, fmt.Errorf("claim window: %w", err)
	}
	if err := dev.SetSwapchainParameters(win, sdl.GPU_SWAPCHAINCOMPOSITION_SDR, sdl.GPU_PRESENTMODE_VSYNC); err != nil {
		fmt.Printf("vsync unavailable: %v\n", err)
	}
	r := &Renderer{
		Dev: dev, Win: win, W: w, H: h,
		fontAtlas:  fontAtlas,
		swapFormat: dev.SwapchainTextureFormat(win),
		chunks:     make(map[[2]int]*gpuChunk),
	}
	if err := r.initDepthFormat(); err != nil {
		r.Close()
		return nil, err
	}
	if err := r.createResources(); err != nil {
		r.Close()
		return nil, err
	}
	fmt.Printf("SDL_GPU driver: %s, swapchain format: %d, depth format: %d\n", dev.Driver(), r.swapFormat, r.depthFormat)
	return r, nil
}

func (r *Renderer) initDepthFormat() error {
	usage := sdl.GPU_TEXTUREUSAGE_DEPTH_STENCIL_TARGET
	for _, f := range []sdl.GPUTextureFormat{sdl.GPU_TEXTUREFORMAT_D32_FLOAT, sdl.GPU_TEXTUREFORMAT_D32_FLOAT_S8_UINT} {
		if r.Dev.TextureSupportsFormat(f, sdl.GPU_TEXTURETYPE_2D, usage) {
			r.depthFormat = f
			return nil
		}
	}
	return errors.New("no supported depth texture format")
}

func (r *Renderer) createResources() error {
	if r.Dev.ShaderFormats()&sdl.GPU_SHADERFORMAT_SPIRV == 0 {
		return errors.New("GPU backend does not expose SPIR-V shader support")
	}
	if err := r.createDepth(); err != nil {
		return err
	}

	// Atlas and font textures.
	atlas := BuildAtlas()
	var err error
	r.atlasTex, err = r.uploadTexture(atlas, core.AtlasW, core.AtlasH)
	if err != nil {
		return err
	}
	r.uiTex, err = r.uploadTexture(r.fontAtlas.Pixels, r.fontAtlas.W, r.fontAtlas.H)
	if err != nil {
		return err
	}
	r.sampler, err = r.Dev.CreateSampler(&sdl.GPUSamplerCreateInfo{
		MinFilter:    sdl.GPU_FILTER_NEAREST,
		MagFilter:    sdl.GPU_FILTER_NEAREST,
		MipmapMode:   sdl.GPU_SAMPLERMIPMAPMODE_NEAREST,
		AddressModeU: sdl.GPU_SAMPLERADDRESSMODE_CLAMP_TO_EDGE,
		AddressModeV: sdl.GPU_SAMPLERADDRESSMODE_CLAMP_TO_EDGE,
		AddressModeW: sdl.GPU_SAMPLERADDRESSMODE_CLAMP_TO_EDGE,
	})
	if err != nil {
		return err
	}

	blockVS, err := r.loadShader("block.vert", sdl.GPU_SHADERSTAGE_VERTEX, 0, 1)
	if err != nil {
		return err
	}
	blockFS, err := r.loadShader("block.frag", sdl.GPU_SHADERSTAGE_FRAGMENT, 1, 1)
	if err != nil {
		r.Dev.ReleaseShader(blockVS)
		return err
	}
	blockLayout := vertexLayout(sdl.GPU_VERTEXELEMENTFORMAT_FLOAT3, sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2)
	blockLayout.VertexAttributes = []sdl.GPUVertexAttribute{
		{BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT3, Location: 0, Offset: 0},
		{BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2, Location: 1, Offset: 12},
		{BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT, Location: 2, Offset: 20},
	}
	blockLayout.VertexBufferDescriptions[0].Pitch = 24
	r.blockOpaque, err = r.makePipeline(blockVS, blockFS, blockLayout, sdl.GPU_PRIMITIVETYPE_TRIANGLELIST, true, false, true, true)
	if err != nil {
		r.Dev.ReleaseShader(blockVS)
		r.Dev.ReleaseShader(blockFS)
		return err
	}
	r.blockTrans, err = r.makePipeline(blockVS, blockFS, blockLayout, sdl.GPU_PRIMITIVETYPE_TRIANGLELIST, true, true, false, true)
	if err != nil {
		r.Dev.ReleaseShader(blockVS)
		r.Dev.ReleaseShader(blockFS)
		return err
	}
	r.blockView, err = r.makePipeline(blockVS, blockFS, blockLayout, sdl.GPU_PRIMITIVETYPE_TRIANGLELIST, true, true, true, true)
	if err != nil {
		r.Dev.ReleaseShader(blockVS)
		r.Dev.ReleaseShader(blockFS)
		return err
	}
	r.Dev.ReleaseShader(blockVS)
	r.Dev.ReleaseShader(blockFS)

	uiVS, err := r.loadShader("ui.vert", sdl.GPU_SHADERSTAGE_VERTEX, 0, 1)
	if err != nil {
		return err
	}
	uiFS, err := r.loadShader("ui.frag", sdl.GPU_SHADERSTAGE_FRAGMENT, 1, 0)
	if err != nil {
		r.Dev.ReleaseShader(uiVS)
		return err
	}
	uiLayout := sdl.GPUVertexInputState{}
	uiLayout.VertexBufferDescriptions = []sdl.GPUVertexBufferDescription{{Slot: 0, Pitch: 32, InputRate: sdl.GPU_VERTEXINPUTRATE_VERTEX}}
	uiLayout.VertexAttributes = []sdl.GPUVertexAttribute{
		{BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2, Location: 0, Offset: 0},
		{BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT2, Location: 1, Offset: 8},
		{BufferSlot: 0, Format: sdl.GPU_VERTEXELEMENTFORMAT_FLOAT4, Location: 2, Offset: 16},
	}
	r.uiPipeline, err = r.makePipeline(uiVS, uiFS, uiLayout, sdl.GPU_PRIMITIVETYPE_TRIANGLELIST, false, true, false, false)
	r.Dev.ReleaseShader(uiVS)
	r.Dev.ReleaseShader(uiFS)
	if err != nil {
		return err
	}

	r.uiCap = 2 * 1024 * 1024
	r.uiVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.uiCap})
	if err != nil {
		return err
	}
	r.entCap = 2 * 1024 * 1024
	r.entVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.entCap})
	if err != nil {
		return err
	}
	r.entIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.entCap})
	if err != nil {
		return err
	}
	r.entTransCap = r.entCap
	r.entTransVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.entTransCap})
	if err != nil {
		return err
	}
	r.entTransIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.entTransCap})
	if err != nil {
		return err
	}
	r.heldCap = r.entCap
	r.heldVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.heldCap})
	if err != nil {
		return err
	}
	r.heldIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.heldCap})
	if err != nil {
		return err
	}
	r.heldTransCap = r.entCap
	r.heldTransVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.heldTransCap})
	if err != nil {
		return err
	}
	r.heldTransIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.heldTransCap})
	if err != nil {
		return err
	}
	r.entIconCap = r.entCap
	r.entIconVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.entIconCap})
	if err != nil {
		return err
	}
	r.entIconIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.entIconCap})
	if err != nil {
		return err
	}
	r.heldIconCap = r.entCap
	r.heldIconVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.heldIconCap})
	if err != nil {
		return err
	}
	r.heldIconIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.heldIconCap})
	if err != nil {
		return err
	}
	r.stagingCap = 8 * 1024 * 1024
	r.staging, err = r.Dev.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD, Size: r.stagingCap})
	return err
}

func vertexLayout(formatA, formatB sdl.GPUVertexElementFormat) sdl.GPUVertexInputState {
	return sdl.GPUVertexInputState{
		VertexBufferDescriptions: []sdl.GPUVertexBufferDescription{{Slot: 0, Pitch: 32, InputRate: sdl.GPU_VERTEXINPUTRATE_VERTEX}},
		VertexAttributes: []sdl.GPUVertexAttribute{
			{BufferSlot: 0, Format: formatA, Location: 0, Offset: 0},
			{BufferSlot: 0, Format: formatB, Location: 1, Offset: 12},
		},
	}
}

func (r *Renderer) makePipeline(vs, fs *sdl.GPUShader, layout sdl.GPUVertexInputState, prim sdl.GPUPrimitiveType, depthTarget, blend, depthWrite, depthTest bool) (*sdl.GPUGraphicsPipeline, error) {
	blendState := sdl.GPUColorTargetBlendState{}
	if blend {
		blendState = sdl.GPUColorTargetBlendState{
			SrcColorBlendfactor: sdl.GPU_BLENDFACTOR_SRC_ALPHA,
			DstColorBlendfactor: sdl.GPU_BLENDFACTOR_ONE_MINUS_SRC_ALPHA,
			ColorBlendOp:        sdl.GPU_BLENDOP_ADD,
			SrcAlphaBlendfactor: sdl.GPU_BLENDFACTOR_ONE,
			DstAlphaBlendfactor: sdl.GPU_BLENDFACTOR_ONE_MINUS_SRC_ALPHA,
			AlphaBlendOp:        sdl.GPU_BLENDOP_ADD,
			EnableBlend:         true,
		}
	}
	info := sdl.GPUGraphicsPipelineCreateInfo{
		VertexShader:     vs,
		FragmentShader:   fs,
		VertexInputState: layout,
		PrimitiveType:    prim,
		TargetInfo: sdl.GPUGraphicsPipelineTargetInfo{
			ColorTargetDescriptions: []sdl.GPUColorTargetDescription{{Format: r.swapFormat, BlendState: blendState}},
			DepthStencilFormat:      r.depthFormat,
			HasDepthStencilTarget:   depthTarget,
		},
		RasterizerState: sdl.GPURasterizerState{
			FillMode:        sdl.GPU_FILLMODE_FILL,
			CullMode:        sdl.GPU_CULLMODE_NONE,
			FrontFace:       sdl.GPU_FRONTFACE_COUNTER_CLOCKWISE,
			EnableDepthClip: true,
		},
		MultisampleState: sdl.GPUMultisampleState{SampleCount: sdl.GPU_SAMPLECOUNT_1},
		DepthStencilState: sdl.GPUDepthStencilState{
			CompareOp:        sdl.GPU_COMPAREOP_LESS,
			EnableDepthTest:  depthTest,
			EnableDepthWrite: depthWrite,
		},
	}
	return r.Dev.CreateGraphicsPipeline(&info)
}

func (r *Renderer) loadShader(name string, stage sdl.GPUShaderStage, samplers, uniforms uint32) (*sdl.GPUShader, error) {
	code, err := shaderFS.ReadFile("shaders/" + name + ".spv")
	if err != nil {
		return nil, err
	}
	return r.Dev.CreateGPUShader(&sdl.GPUShaderCreateInfo{
		Code:              code,
		Entrypoint:        "main",
		Format:            sdl.GPU_SHADERFORMAT_SPIRV,
		Stage:             stage,
		NumSamplers:       samplers,
		NumUniformBuffers: uniforms,
	})
}

func (r *Renderer) uploadTexture(pixels []byte, w, h int) (*sdl.GPUTexture, error) {
	tex, err := r.Dev.CreateTexture(&sdl.GPUTextureCreateInfo{
		Type:              sdl.GPU_TEXTURETYPE_2D,
		Format:            sdl.GPU_TEXTUREFORMAT_R8G8B8A8_UNORM,
		Usage:             sdl.GPU_TEXTUREUSAGE_SAMPLER,
		Width:             uint32(w),
		Height:            uint32(h),
		LayerCountOrDepth: 1,
		NumLevels:         1,
		SampleCount:       sdl.GPU_SAMPLECOUNT_1,
	})
	if err != nil {
		return nil, err
	}
	tb, err := r.Dev.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{
		Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD,
		Size:  uint32(len(pixels)),
	})
	if err != nil {
		r.Dev.ReleaseTexture(tex)
		return nil, err
	}
	ptr, err := r.Dev.MapTransferBuffer(tb, false)
	if err != nil {
		r.Dev.ReleaseTransferBuffer(tb)
		r.Dev.ReleaseTexture(tex)
		return nil, err
	}
	copy(mappedBytes(ptr, len(pixels)), pixels)
	r.Dev.UnmapTransferBuffer(tb)
	cb, err := r.Dev.AcquireCommandBuffer()
	if err != nil {
		r.Dev.ReleaseTransferBuffer(tb)
		r.Dev.ReleaseTexture(tex)
		return nil, err
	}
	cp := cb.BeginCopyPass()
	cp.UploadToGPUTexture(
		&sdl.GPUTextureTransferInfo{TransferBuffer: tb, Offset: 0, PixelsPerRow: uint32(w), RowsPerLayer: uint32(h)},
		&sdl.GPUTextureRegion{Texture: tex, W: uint32(w), H: uint32(h), D: 1},
		false,
	)
	cp.End()
	if err := cb.Submit(); err != nil {
		r.Dev.ReleaseTransferBuffer(tb)
		r.Dev.ReleaseTexture(tex)
		return nil, err
	}
	r.Dev.ReleaseTransferBuffer(tb)
	return tex, nil
}

func (r *Renderer) createDepth() error {
	if r.W <= 0 || r.H <= 0 {
		return fmt.Errorf("invalid render size %dx%d", r.W, r.H)
	}
	tex, err := r.Dev.CreateTexture(&sdl.GPUTextureCreateInfo{
		Type:              sdl.GPU_TEXTURETYPE_2D,
		Format:            r.depthFormat,
		Usage:             sdl.GPU_TEXTUREUSAGE_DEPTH_STENCIL_TARGET,
		Width:             uint32(r.W),
		Height:            uint32(r.H),
		LayerCountOrDepth: 1,
		NumLevels:         1,
		SampleCount:       sdl.GPU_SAMPLECOUNT_1,
	})
	if err != nil {
		return err
	}
	if r.depth != nil {
		r.Dev.ReleaseTexture(r.depth)
	}
	r.depth = tex
	return nil
}

func (r *Renderer) Resize(w, h int) {
	if w == r.W && h == r.H {
		return
	}
	r.W, r.H = w, h
	r.Dev.WaitForIdle()
	if err := r.createDepth(); err != nil {
		fmt.Println("depth resize failed:", err)
	}
}

func (r *Renderer) UploadChunk(pos world.ChunkPos, mesh world.ChunkMesh) error {
	key := pos.Key()
	gc := r.chunks[key]
	if gc == nil {
		gc = &gpuChunk{pos: pos}
		r.chunks[key] = gc
	}
	var err error
	if gc.opaque, err = r.uploadMesh(gc.opaque, mesh.Opaque); err != nil {
		return err
	}
	if gc.transparent, err = r.uploadMesh(gc.transparent, mesh.Transparent); err != nil {
		return err
	}
	return nil
}

func (r *Renderer) uploadMesh(old gpuMesh, md world.MeshData) (gpuMesh, error) {
	old.release(r.Dev)
	if len(md.Vertices) == 0 || len(md.Indices) == 0 {
		return gpuMesh{}, nil
	}
	vb, err := r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: uint32(len(md.Vertices) * 24)})
	if err != nil {
		return gpuMesh{}, err
	}
	ib, err := r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: uint32(len(md.Indices) * 4)})
	if err != nil {
		r.Dev.ReleaseBuffer(vb)
		return gpuMesh{}, err
	}
	size := len(md.Vertices)*24 + len(md.Indices)*4
	tb, err := r.Dev.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD, Size: uint32(size)})
	if err != nil {
		r.Dev.ReleaseBuffer(vb)
		r.Dev.ReleaseBuffer(ib)
		return gpuMesh{}, err
	}
	ptr, err := r.Dev.MapTransferBuffer(tb, false)
	if err != nil {
		r.Dev.ReleaseTransferBuffer(tb)
		r.Dev.ReleaseBuffer(vb)
		r.Dev.ReleaseBuffer(ib)
		return gpuMesh{}, err
	}
	mem := mappedBytes(ptr, size)
	vbytes := unsafe.Slice((*byte)(unsafe.Pointer(&md.Vertices[0])), len(md.Vertices)*24)
	copy(mem, vbytes)
	ibytes := unsafe.Slice((*byte)(unsafe.Pointer(&md.Indices[0])), len(md.Indices)*4)
	copy(mem[len(vbytes):], ibytes)
	r.Dev.UnmapTransferBuffer(tb)
	cb, err := r.Dev.AcquireCommandBuffer()
	if err != nil {
		r.Dev.ReleaseTransferBuffer(tb)
		r.Dev.ReleaseBuffer(vb)
		r.Dev.ReleaseBuffer(ib)
		return gpuMesh{}, err
	}
	cp := cb.BeginCopyPass()
	cp.UploadToGPUBuffer(
		&sdl.GPUTransferBufferLocation{TransferBuffer: tb, Offset: 0},
		&sdl.GPUBufferRegion{Buffer: vb, Offset: 0, Size: uint32(len(vbytes))},
		false,
	)
	cp.UploadToGPUBuffer(
		&sdl.GPUTransferBufferLocation{TransferBuffer: tb, Offset: uint32(len(vbytes))},
		&sdl.GPUBufferRegion{Buffer: ib, Offset: 0, Size: uint32(len(ibytes))},
		false,
	)
	cp.End()
	if err := cb.Submit(); err != nil {
		r.Dev.ReleaseTransferBuffer(tb)
		r.Dev.ReleaseBuffer(vb)
		r.Dev.ReleaseBuffer(ib)
		return gpuMesh{}, err
	}
	r.Dev.ReleaseTransferBuffer(tb)
	return gpuMesh{vb: vb, ib: ib, count: uint32(len(md.Indices))}, nil
}

func (r *Renderer) ChunkPositions() []world.ChunkPos {
	out := make([]world.ChunkPos, 0, len(r.chunks))
	for _, gc := range r.chunks {
		out = append(out, gc.pos)
	}
	return out
}

func (r *Renderer) ReleaseAllChunks() {
	for _, gc := range r.chunks {
		gc.opaque.release(r.Dev)
		gc.transparent.release(r.Dev)
	}
	r.chunks = make(map[[2]int]*gpuChunk)
}

func (r *Renderer) ReleaseChunk(pos world.ChunkPos) {
	key := pos.Key()
	if gc := r.chunks[key]; gc != nil {
		gc.opaque.release(r.Dev)
		gc.transparent.release(r.Dev)
		delete(r.chunks, key)
	}
}

func (r *Renderer) Render(f *Frame) error {
	if w, h, err := r.Win.SizeInPixels(); err == nil && w > 0 && h > 0 && (int(w) != r.W || int(h) != r.H) {
		r.Resize(int(w), int(h))
		if f.UI != nil {
			f.UI.Resize(r.W, r.H)
		}
	}
	if f.UI != nil {
		f.UI.Resize(r.W, r.H)
	}
	if err := r.updateDynamic(f); err != nil {
		return err
	}
	cb, err := r.Dev.AcquireCommandBuffer()
	if err != nil {
		return err
	}
	sw, err := cb.WaitAndAcquireGPUSwapchainTexture(r.Win)
	if err != nil {
		return err
	}
	if sw == nil {
		return cb.Submit()
	}
	if int(sw.Width) != r.W || int(sw.Height) != r.H {
		fmt.Printf("swapchain size %dx%d differs from render size %dx%d\n", sw.Width, sw.Height, r.W, r.H)
	}
	w := float32(sw.Width)
	h := float32(sw.Height)
	aspect := w / h
	f.Camera.Aspect = aspect
	viewProj := f.Camera.ViewProj()

	wu := worldUBO{
		ViewProj:  viewProj,
		FogColor:  [4]float32{f.World.FogColor[0], f.World.FogColor[1], f.World.FogColor[2], 1},
		FogParams: [4]float32{f.World.FogStart, f.World.FogEnd, 0, 0},
		Cam:       [4]float32{f.Camera.Pos.X, f.Camera.Pos.Y, f.Camera.Pos.Z, 0},
		Sun:       [4]float32{f.World.SunTint[0], f.World.SunTint[1], f.World.SunTint[2], 1},
	}
	color := sdl.GPUColorTargetInfo{
		Texture:    sw.Texture,
		ClearColor: sdl.FColor{R: f.World.Clear[0], G: f.World.Clear[1], B: f.World.Clear[2], A: f.World.Clear[3]},
		LoadOp:     sdl.GPU_LOADOP_CLEAR,
		StoreOp:    sdl.GPU_STOREOP_STORE,
	}
	depth := sdl.GPUDepthStencilTargetInfo{
		Texture:        r.depth,
		Cycle:          true,
		ClearDepth:     1,
		ClearStencil:   0,
		LoadOp:         sdl.GPU_LOADOP_CLEAR,
		StoreOp:        sdl.GPU_STOREOP_DONT_CARE,
		StencilLoadOp:  sdl.GPU_LOADOP_DONT_CARE,
		StencilStoreOp: sdl.GPU_STOREOP_DONT_CARE,
	}
	rp := cb.BeginRenderPass([]sdl.GPUColorTargetInfo{color}, &depth)
	rp.BindGraphicsPipeline(r.blockOpaque)
	cb.PushVertexUniformData(0, bytesOf(&wu))
	cb.PushFragmentUniformData(0, bytesOf(&wu))
	rp.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{{Texture: r.atlasTex, Sampler: r.sampler}})
	for _, gc := range r.chunks {
		if !r.chunkVisible(gc, f.Camera, f.World.FogEnd) {
			continue
		}
		if gc.opaque.count == 0 {
			continue
		}
		r.drawMesh(rp, &gc.opaque)
	}

	if r.entICount > 0 && r.entVB != nil && r.entIB != nil {
		rp.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.entVB, Offset: 0}})
		rp.BindIndexBuffer(&sdl.GPUBufferBinding{Buffer: r.entIB, Offset: 0}, sdl.GPU_INDEXELEMENTSIZE_32BIT)
		rp.DrawIndexedPrimitives(r.entICount, 1, 0, 0, 0)
	}
	if r.entIconICount > 0 && r.entIconVB != nil && r.entIconIB != nil {
		rp.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{{Texture: r.uiTex, Sampler: r.sampler}})
		rp.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.entIconVB, Offset: 0}})
		rp.BindIndexBuffer(&sdl.GPUBufferBinding{Buffer: r.entIconIB, Offset: 0}, sdl.GPU_INDEXELEMENTSIZE_32BIT)
		rp.DrawIndexedPrimitives(r.entIconICount, 1, 0, 0, 0)
	}

	// Transparent chunk pass, back to front.
	rp.BindGraphicsPipeline(r.blockTrans)
	cb.PushVertexUniformData(0, bytesOf(&wu))
	cb.PushFragmentUniformData(0, bytesOf(&wu))
	rp.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{{Texture: r.atlasTex, Sampler: r.sampler}})
	trans := make([]*gpuChunk, 0, len(r.chunks))
	for _, gc := range r.chunks {
		if gc.transparent.count > 0 && r.chunkVisible(gc, f.Camera, f.World.FogEnd) {
			trans = append(trans, gc)
		}
	}
	sort.Slice(trans, func(i, j int) bool {
		return chunkDistSq(trans[i], f.Camera.Pos) > chunkDistSq(trans[j], f.Camera.Pos)
	})
	for _, gc := range trans {
		r.drawMesh(rp, &gc.transparent)
	}
	if r.entTransICount > 0 && r.entTransVB != nil && r.entTransIB != nil {
		rp.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.entTransVB, Offset: 0}})
		rp.BindIndexBuffer(&sdl.GPUBufferBinding{Buffer: r.entTransIB, Offset: 0}, sdl.GPU_INDEXELEMENTSIZE_32BIT)
		rp.DrawIndexedPrimitives(r.entTransICount, 1, 0, 0, 0)
	}
	rp.End()

	// First-person held item: a dedicated pass with a fresh depth buffer so
	// its own front faces occlude its back faces without clipping into the
	// nearby world geometry.
	if r.heldICount > 0 || r.heldTransICount > 0 || r.heldIconICount > 0 {
		heldColor := sdl.GPUColorTargetInfo{
			Texture: sw.Texture,
			LoadOp:  sdl.GPU_LOADOP_LOAD,
			StoreOp: sdl.GPU_STOREOP_STORE,
		}
		heldDepth := sdl.GPUDepthStencilTargetInfo{
			Texture:        r.depth,
			Cycle:          true,
			ClearDepth:     1,
			ClearStencil:   0,
			LoadOp:         sdl.GPU_LOADOP_CLEAR,
			StoreOp:        sdl.GPU_STOREOP_DONT_CARE,
			StencilLoadOp:  sdl.GPU_LOADOP_DONT_CARE,
			StencilStoreOp: sdl.GPU_STOREOP_DONT_CARE,
		}
		hvp := cb.BeginRenderPass([]sdl.GPUColorTargetInfo{heldColor}, &heldDepth)
		hvp.BindGraphicsPipeline(r.blockView)
		cb.PushVertexUniformData(0, bytesOf(&wu))
		cb.PushFragmentUniformData(0, bytesOf(&wu))
		hvp.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{{Texture: r.atlasTex, Sampler: r.sampler}})
		if r.heldICount > 0 && r.heldVB != nil && r.heldIB != nil {
			hvp.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.heldVB, Offset: 0}})
			hvp.BindIndexBuffer(&sdl.GPUBufferBinding{Buffer: r.heldIB, Offset: 0}, sdl.GPU_INDEXELEMENTSIZE_32BIT)
			hvp.DrawIndexedPrimitives(r.heldICount, 1, 0, 0, 0)
		}
		if r.heldTransICount > 0 && r.heldTransVB != nil && r.heldTransIB != nil {
			// Glass/ice: transparent blend, no depth writes. Otherwise one
			// cube face writes depth first and erases a neighbouring edge.
			hvp.BindGraphicsPipeline(r.blockTrans)
			hvp.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{{Texture: r.atlasTex, Sampler: r.sampler}})
			hvp.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.heldTransVB, Offset: 0}})
			hvp.BindIndexBuffer(&sdl.GPUBufferBinding{Buffer: r.heldTransIB, Offset: 0}, sdl.GPU_INDEXELEMENTSIZE_32BIT)
			hvp.DrawIndexedPrimitives(r.heldTransICount, 1, 0, 0, 0)
		}
		if r.heldIconICount > 0 && r.heldIconVB != nil && r.heldIconIB != nil {
			// Match dropped items: alpha test only, no blending. The icon
			// texture has anti-aliased edge pixels; blending them against
			// the world produced a translucent pixel fringe.
			hvp.BindGraphicsPipeline(r.blockOpaque)
			hvp.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{{Texture: r.uiTex, Sampler: r.sampler}})
			hvp.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.heldIconVB, Offset: 0}})
			hvp.BindIndexBuffer(&sdl.GPUBufferBinding{Buffer: r.heldIconIB, Offset: 0}, sdl.GPU_INDEXELEMENTSIZE_32BIT)
			hvp.DrawIndexedPrimitives(r.heldIconICount, 1, 0, 0, 0)
		}
		hvp.End()
	}

	if f.UI != nil && len(f.UI.Verts) > 0 && r.uiVB != nil {
		uiColor := sdl.GPUColorTargetInfo{
			Texture: sw.Texture,
			LoadOp:  sdl.GPU_LOADOP_LOAD,
			StoreOp: sdl.GPU_STOREOP_STORE,
		}
		rp2 := cb.BeginRenderPass([]sdl.GPUColorTargetInfo{uiColor}, nil)
		rp2.BindGraphicsPipeline(r.uiPipeline)
		cb.PushVertexUniformData(0, bytesOf(&uiUBO{Screen: [4]float32{w, h, 0, 0}}))
		rp2.BindFragmentSamplers([]sdl.GPUTextureSamplerBinding{{Texture: r.uiTex, Sampler: r.sampler}})
		rp2.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: r.uiVB, Offset: 0}})
		rp2.DrawPrimitives(uint32(len(f.UI.Verts)), 1, 0, 0)
		rp2.End()
	}
	return cb.Submit()
}

func (r *Renderer) updateDynamic(f *Frame) error {
	uiBytes := uiVertBytes(f.UI)
	entV := meshVertBytes(f.Entity)
	entI := meshIndexBytes(f.Entity)
	entTV := meshVertBytes(f.EntityTransparent)
	entTI := meshIndexBytes(f.EntityTransparent)
	entIV := meshVertBytes(f.EntityIcon)
	entII := meshIndexBytes(f.EntityIcon)
	heldV := meshVertBytes(f.Held)
	heldI := meshIndexBytes(f.Held)
	heldTV := meshVertBytes(f.HeldTransparent)
	heldTI := meshIndexBytes(f.HeldTransparent)
	heldIV := meshVertBytes(f.HeldIcon)
	heldII := meshIndexBytes(f.HeldIcon)
	r.entICount = uint32(len(f.Entity.Indices))
	r.entTransICount = uint32(len(f.EntityTransparent.Indices))
	r.entIconICount = uint32(len(f.EntityIcon.Indices))
	r.heldICount = uint32(len(f.Held.Indices))
	r.heldTransICount = uint32(len(f.HeldTransparent.Indices))
	r.heldIconICount = uint32(len(f.HeldIcon.Indices))
	total := len(uiBytes) + len(entV) + len(entI) + len(entTV) + len(entTI) + len(entIV) + len(entII) + len(heldV) + len(heldI) + len(heldTV) + len(heldTI) + len(heldIV) + len(heldII)
	if total == 0 {
		return nil
	}
	if uint32(total) > r.stagingCap {
		r.Dev.ReleaseTransferBuffer(r.staging)
		r.stagingCap = uint32(total * 2)
		var err error
		r.staging, err = r.Dev.CreateTransferBuffer(&sdl.GPUTransferBufferCreateInfo{Usage: sdl.GPU_TRANSFERBUFFERUSAGE_UPLOAD, Size: r.stagingCap})
		if err != nil {
			return err
		}
	}
	if uint32(len(uiBytes)) > r.uiCap {
		r.Dev.ReleaseBuffer(r.uiVB)
		r.uiCap = uint32(len(uiBytes) * 2)
		var err error
		r.uiVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.uiCap})
		if err != nil {
			return err
		}
	}
	needEnt := maxInt(len(entV), len(entI))
	if uint32(needEnt) > r.entCap {
		r.Dev.ReleaseBuffer(r.entVB)
		r.Dev.ReleaseBuffer(r.entIB)
		r.entCap = uint32(needEnt * 2)
		var err error
		r.entVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.entCap})
		if err != nil {
			return err
		}
		r.entIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.entCap})
		if err != nil {
			return err
		}
	}
	needTrans := maxInt(len(entTV), len(entTI))
	if uint32(needTrans) > r.entTransCap {
		r.Dev.ReleaseBuffer(r.entTransVB)
		r.Dev.ReleaseBuffer(r.entTransIB)
		r.entTransCap = uint32(needTrans * 2)
		var err error
		r.entTransVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.entTransCap})
		if err != nil {
			return err
		}
		r.entTransIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.entTransCap})
		if err != nil {
			return err
		}
	}
	needEntIcon := maxInt(len(entIV), len(entII))
	if uint32(needEntIcon) > r.entIconCap {
		r.Dev.ReleaseBuffer(r.entIconVB)
		r.Dev.ReleaseBuffer(r.entIconIB)
		r.entIconCap = uint32(needEntIcon * 2)
		var err error
		r.entIconVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.entIconCap})
		if err != nil {
			return err
		}
		r.entIconIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.entIconCap})
		if err != nil {
			return err
		}
	}
	needHeld := maxInt(len(heldV), len(heldI))
	if uint32(needHeld) > r.heldCap {
		r.Dev.ReleaseBuffer(r.heldVB)
		r.Dev.ReleaseBuffer(r.heldIB)
		r.heldCap = uint32(needHeld * 2)
		var err error
		r.heldVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.heldCap})
		if err != nil {
			return err
		}
		r.heldIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.heldCap})
		if err != nil {
			return err
		}
	}
	needHeldTrans := maxInt(len(heldTV), len(heldTI))
	if uint32(needHeldTrans) > r.heldTransCap {
		r.Dev.ReleaseBuffer(r.heldTransVB)
		r.Dev.ReleaseBuffer(r.heldTransIB)
		r.heldTransCap = uint32(needHeldTrans * 2)
		var err error
		r.heldTransVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.heldTransCap})
		if err != nil {
			return err
		}
		r.heldTransIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.heldTransCap})
		if err != nil {
			return err
		}
	}
	needHeldIcon := maxInt(len(heldIV), len(heldII))
	if uint32(needHeldIcon) > r.heldIconCap {
		r.Dev.ReleaseBuffer(r.heldIconVB)
		r.Dev.ReleaseBuffer(r.heldIconIB)
		r.heldIconCap = uint32(needHeldIcon * 2)
		var err error
		r.heldIconVB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_VERTEX, Size: r.heldIconCap})
		if err != nil {
			return err
		}
		r.heldIconIB, err = r.Dev.CreateBuffer(&sdl.GPUBufferCreateInfo{Usage: sdl.GPU_BUFFERUSAGE_INDEX, Size: r.heldIconCap})
		if err != nil {
			return err
		}
	}
	ptr, err := r.Dev.MapTransferBuffer(r.staging, true)
	if err != nil {
		return err
	}
	mem := mappedBytes(ptr, total)
	offUI := 0
	offEntV := offUI + len(uiBytes)
	offEntI := offEntV + len(entV)
	offEntTV := offEntI + len(entI)
	offEntTI := offEntTV + len(entTV)
	offEntIV := offEntTI + len(entTI)
	offEntII := offEntIV + len(entIV)
	offHeldV := offEntII + len(entII)
	offHeldI := offHeldV + len(heldV)
	offHeldTV := offHeldI + len(heldI)
	offHeldTI := offHeldTV + len(heldTV)
	offHeldIV := offHeldTI + len(heldTI)
	offHeldII := offHeldIV + len(heldIV)
	copy(mem[offUI:], uiBytes)
	copy(mem[offEntV:], entV)
	copy(mem[offEntI:], entI)
	copy(mem[offEntTV:], entTV)
	copy(mem[offEntTI:], entTI)
	copy(mem[offEntIV:], entIV)
	copy(mem[offEntII:], entII)
	copy(mem[offHeldV:], heldV)
	copy(mem[offHeldI:], heldI)
	copy(mem[offHeldTV:], heldTV)
	copy(mem[offHeldTI:], heldTI)
	copy(mem[offHeldIV:], heldIV)
	copy(mem[offHeldII:], heldII)
	r.Dev.UnmapTransferBuffer(r.staging)
	cb, err := r.Dev.AcquireCommandBuffer()
	if err != nil {
		return err
	}
	cp := cb.BeginCopyPass()
	if len(uiBytes) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offUI)},
			&sdl.GPUBufferRegion{Buffer: r.uiVB, Offset: 0, Size: uint32(len(uiBytes))},
			true,
		)
	}
	if len(entV) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offEntV)},
			&sdl.GPUBufferRegion{Buffer: r.entVB, Offset: 0, Size: uint32(len(entV))},
			true,
		)
	}
	if len(entI) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offEntI)},
			&sdl.GPUBufferRegion{Buffer: r.entIB, Offset: 0, Size: uint32(len(entI))},
			true,
		)
	}
	if len(entTV) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offEntTV)},
			&sdl.GPUBufferRegion{Buffer: r.entTransVB, Offset: 0, Size: uint32(len(entTV))},
			true,
		)
	}
	if len(entTI) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offEntTI)},
			&sdl.GPUBufferRegion{Buffer: r.entTransIB, Offset: 0, Size: uint32(len(entTI))},
			true,
		)
	}
	if len(entIV) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offEntIV)},
			&sdl.GPUBufferRegion{Buffer: r.entIconVB, Offset: 0, Size: uint32(len(entIV))},
			true,
		)
	}
	if len(entII) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offEntII)},
			&sdl.GPUBufferRegion{Buffer: r.entIconIB, Offset: 0, Size: uint32(len(entII))},
			true,
		)
	}
	if len(heldV) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offHeldV)},
			&sdl.GPUBufferRegion{Buffer: r.heldVB, Offset: 0, Size: uint32(len(heldV))},
			true,
		)
	}
	if len(heldI) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offHeldI)},
			&sdl.GPUBufferRegion{Buffer: r.heldIB, Offset: 0, Size: uint32(len(heldI))},
			true,
		)
	}
	if len(heldTV) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offHeldTV)},
			&sdl.GPUBufferRegion{Buffer: r.heldTransVB, Offset: 0, Size: uint32(len(heldTV))},
			true,
		)
	}
	if len(heldTI) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offHeldTI)},
			&sdl.GPUBufferRegion{Buffer: r.heldTransIB, Offset: 0, Size: uint32(len(heldTI))},
			true,
		)
	}
	if len(heldIV) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offHeldIV)},
			&sdl.GPUBufferRegion{Buffer: r.heldIconVB, Offset: 0, Size: uint32(len(heldIV))},
			true,
		)
	}
	if len(heldII) > 0 {
		cp.UploadToGPUBuffer(
			&sdl.GPUTransferBufferLocation{TransferBuffer: r.staging, Offset: uint32(offHeldII)},
			&sdl.GPUBufferRegion{Buffer: r.heldIconIB, Offset: 0, Size: uint32(len(heldII))},
			true,
		)
	}
	cp.End()
	return cb.Submit()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func meshVertBytes(m world.MeshData) []byte {
	if len(m.Vertices) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&m.Vertices[0])), len(m.Vertices)*24)
}

func meshIndexBytes(m world.MeshData) []byte {
	if len(m.Indices) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&m.Indices[0])), len(m.Indices)*4)
}

func (r *Renderer) drawMesh(rp *sdl.GPURenderPass, m *gpuMesh) {
	rp.BindVertexBuffers([]sdl.GPUBufferBinding{{Buffer: m.vb, Offset: 0}})
	rp.BindIndexBuffer(&sdl.GPUBufferBinding{Buffer: m.ib, Offset: 0}, sdl.GPU_INDEXELEMENTSIZE_32BIT)
	rp.DrawIndexedPrimitives(m.count, 1, 0, 0, 0)
}

func (r *Renderer) chunkVisible(gc *gpuChunk, cam Camera, fogEnd float32) bool {
	dx := float32(gc.pos.X*ChunkSize) + ChunkSize/2 - cam.Pos.X
	dz := float32(gc.pos.Z*ChunkSize) + ChunkSize/2 - cam.Pos.Z
	dy := 64 - cam.Pos.Y
	return dx*dx+dy*dy+dz*dz < (fogEnd+40)*(fogEnd+40)
}

func chunkDistSq(gc *gpuChunk, p mathx.Vec3) float32 {
	dx := float32(gc.pos.X*ChunkSize+ChunkSize/2) - p.X
	dz := float32(gc.pos.Z*ChunkSize+ChunkSize/2) - p.Z
	return dx*dx + dz*dz
}

// ChunkSize mirrors world.ChunkSize for local culling code.
const ChunkSize = world.ChunkSize

func (r *Renderer) Close() {
	r.Dev.WaitForIdle()
	for _, gc := range r.chunks {
		gc.opaque.release(r.Dev)
		gc.transparent.release(r.Dev)
	}
	r.chunks = map[[2]int]*gpuChunk{}
	for _, b := range []*sdl.GPUBuffer{r.uiVB, r.entVB, r.entIB, r.entTransVB, r.entTransIB, r.entIconVB, r.entIconIB, r.heldVB, r.heldIB, r.heldTransVB, r.heldTransIB, r.heldIconVB, r.heldIconIB} {
		if b != nil {
			r.Dev.ReleaseBuffer(b)
		}
	}
	if r.staging != nil {
		r.Dev.ReleaseTransferBuffer(r.staging)
	}
	for _, p := range []*sdl.GPUGraphicsPipeline{r.blockOpaque, r.blockTrans, r.blockView, r.uiPipeline} {
		if p != nil {
			r.Dev.ReleaseGraphicsPipeline(p)
		}
	}
	if r.atlasTex != nil {
		r.Dev.ReleaseTexture(r.atlasTex)
	}
	if r.uiTex != nil {
		r.Dev.ReleaseTexture(r.uiTex)
	}
	if r.depth != nil {
		r.Dev.ReleaseTexture(r.depth)
	}
	if r.sampler != nil {
		r.Dev.ReleaseSampler(r.sampler)
	}
	r.Dev.ReleaseWindow(r.Win)
	r.Dev.Destroy()
}

//go:nocheckptr
func mappedBytes(ptr uintptr, n int) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(ptr)), n)
}

func bytesOf[T any](v *T) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
}

func uiVertBytes(c *ui.Canvas) []byte {
	if c == nil || len(c.Verts) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&c.Verts[0])), len(c.Verts)*32)
}
