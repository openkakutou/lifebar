package lifebar

import (
	"math"

	"github.com/openkakutou/character/air"
)

// referenceSize is the 320x240 game size Ikemen GO measures lifebar
// positions against (fightscreen.go:715-716, common.go:1472).
var referenceSize = [2]float64{320, 240}

// fntScale is the one-unit shift Ikemen GO gives a mirrored element: its
// default fnt_scale (fightscreen.go:4846; 0.5 with MUGEN's doubleres, which
// is not modelled).
const fntScale = 1

// Rect is an area of the view, in view pixels.
type Rect struct{ X, Y, W, H float64 }

// Screen maps lifebar coordinates to a view: it holds the three values
// Ikemen GO's FightScreen.setScale works out, and the game and view sizes
// they were worked out for.
type Screen struct {
	// Scale multiplies lifebar units into game units.
	Scale float64
	// OffsetX is added to every x position, in lifebar units.
	OffsetX float64
	// PortraitScale is the factor Ikemen GO applies to face sprites.
	PortraitScale float64

	game, pixels [2]float64
}

// Screen is FightScreen.setScale (fightscreen.go:5972) for a game of
// game[0] x game[1] units shown on a view of pixels[0] x pixels[1] pixels.
// A localcoord or game size that is not positive is replaced by 320x240, so
// a zero value never divides.
func (lb *Lifebar) Screen(game, pixels [2]int) Screen {
	localcoord := lb.Localcoord
	if localcoord[0] <= 0 || localcoord[1] <= 0 {
		localcoord = defaultLocalcoord
	}
	gameSize := [2]float64{float64(game[0]), float64(game[1])}
	if game[0] <= 0 || game[1] <= 0 {
		gameSize = referenceSize
	}
	localW := float64(localcoord[0])
	viewport43 := viewportWidth(localcoord, [2]float64{4, 3})
	viewport := viewportWidth(localcoord, gameSize)
	calcScale := viewport / localW
	calcOffsetX := (viewport43 - localW*calcScale) / 2
	return Screen{
		Scale:         referenceSize[0] / viewport43 * calcScale,
		OffsetX:       calcOffsetX / calcScale,
		PortraitScale: localW / viewport43 * calcScale,
		game:          gameSize,
		pixels:        [2]float64{float64(pixels[0]), float64(pixels[1])},
	}
}

// viewportWidth is the width of getViewport (system.go:584): the part of a
// src-sized area that has the aspect ratio dst.
func viewportWidth(src [2]int, dst [2]float64) float64 {
	srcW, srcH := float64(src[0]), float64(src[1])
	fromRatio := srcW * dst[1]
	toRatio := srcH * dst[0]
	if fromRatio > toRatio {
		return srcH * dst[0] / dst[1]
	}
	return srcW
}

// axisView is what calcBarFillRect and the placement need of one axis.
type axisView struct {
	offset float64 // fightScreen.offsetX / offsetY, in lifebar units
	scale  float64 // fightScreen.scale
	mid    float64 // (game size - 320x240) / 2, in game units
	screen float64 // view pixels per game unit
}

func (s Screen) axis(i int) axisView {
	v := axisView{
		scale:  s.Scale,
		mid:    (s.game[i] - referenceSize[i]) / 2,
		screen: float64(float32(s.pixels[i]) / float32(s.game[i])), // sys.widthScale is float32
	}
	if i == 0 {
		v.offset = s.OffsetX
	}
	return v
}

// fillSpan is calcBarFillRect (fightscreen.go:418) for one axis: the start
// and size, in view pixels, of a bar whose range is rng (descending when it
// empties toward its first value) filled at fill (0..1). It computes in
// float32 in Ikemen GO's operation order, as the +0.5 truncation is
// sensitive to the rounding.
func fillSpan(pos int, rng [2]int, fill float32, v axisView) (start, size int) {
	r0, r1 := rng[0], rng[1]
	descending := r0 > r1
	if descending {
		r0, r1 = r1, r0
	}
	offset, scale := float32(v.offset), float32(v.scale)
	mid, screen := float32(v.mid), float32(v.screen)
	fillLength := float32(r1 - r0 + 1)
	size = int(fillLength*scale*fill*screen + 0.5)
	base := float32(pos + r0)
	start = int(((base+offset)*scale+mid)*screen + 0.5)
	if descending {
		start = int(((float32(pos+r1+1)+offset)*scale+mid)*screen+0.5) - size
	}
	return start, size
}

// Clips are the clip areas of a bar's front fill (at top) and trailing fill
// (at mid), both fractions 0..1, as LifeBar.draw works them out
// (fightscreen.go:717-770, without scalefill). The trailing area only keeps
// the part the front does not cover. An axis without a range spans the view.
func (b Bar) Clips(s Screen, top, mid float32) (front, trail Rect) {
	front = b.clip(s, top)
	trail = b.clip(s, mid)
	if b.RangeY != [2]int{} {
		if b.RangeY[0] < b.RangeY[1] {
			trail.Y += front.H
		}
		trail.H -= min(trail.H, front.H)
		return front, trail
	}
	if b.RangeX[0] < b.RangeX[1] {
		trail.X += front.W
	}
	trail.W -= min(trail.W, front.W)
	return front, trail
}

func (b Bar) clip(s Screen, fill float32) Rect {
	r := Rect{W: s.pixels[0], H: s.pixels[1]}
	if b.RangeX != [2]int{} {
		x, w := fillSpan(b.Pos[0], b.RangeX, fill, s.axis(0))
		r.X, r.W = float64(x), float64(w)
	}
	if b.RangeY != [2]int{} {
		y, h := fillSpan(b.Pos[1], b.RangeY, fill, s.axis(1))
		r.Y, r.H = float64(y), float64(h)
	}
	return r
}

// SpriteBox is a sprite's own axis and size, in sprite pixels.
type SpriteBox struct{ Axis, Size [2]int }

// Placement is where a frame lands in the view: its top-left corner and
// size, and whether it is drawn mirrored.
type Placement struct {
	Rect
	FlipX, FlipY bool
}

// Place is where frame fr of element e, owned by something at pos, lands in
// the view, as Layout.DrawAnim and Animation.Draw place it (common.go:1445,
// anim.go:1056). The x position includes the
// (game width - 320) / 2 centering Clips applies (fightscreen.go:715); the
// y position includes the game height - 240 shift of common.go:1472.
func (s Screen) Place(e Element, fr air.Frame, spr SpriteBox, pos [2]int) Placement {
	lay := e.Layout
	xv, yv := s.axis(0), s.axis(1)
	x, w, flipX := placeAxis(xv, axisInput{
		pos: float64(pos[0]), offset: float64(lay.Offset[0]), facing: lay.Facing,
		scale: float64(lay.Scale[0]), frameShift: float64(fr.X),
		flip: fr.Flip == air.FlipH || fr.Flip == air.FlipHV,
		axis: spr.Axis[0], size: spr.Size[0],
	})
	y, h, flipY := placeAxis(yv, axisInput{
		pos: float64(pos[1]) + s.game[1] - referenceSize[1], offset: float64(lay.Offset[1]), facing: lay.VFacing,
		scale: float64(lay.Scale[1]), frameShift: float64(fr.Y),
		flip: fr.Flip == air.FlipV || fr.Flip == air.FlipHV,
		axis: spr.Axis[1], size: spr.Size[1],
	})
	x += xv.mid * xv.screen
	return Placement{Rect: Rect{X: x, Y: y, W: w, H: h}, FlipX: flipX, FlipY: flipY}
}

// axisInput is what placing a frame needs of one axis. facing is the
// element's facing (x) or vfacing (y), flip whether the frame is mirrored
// along that axis, axis and size the sprite's own.
type axisInput struct {
	pos, offset float64
	facing      int
	scale       float64
	frameShift  float64
	flip        bool
	axis, size  int
}

// placeAxis places one axis. A mirrored axis shifts by fntScale
// (common.go:1462-1467). The signed factor is the layout scale times the
// facing times the frame flip times the lifebar scale (anim.go:1072-1073);
// the frame offset is multiplied by it (anim.go:1091-1092). When it is
// negative the sprite is mirrored around its anchor, so its axis sits at the
// far side.
func placeAxis(v axisView, in axisInput) (start, length float64, mirrored bool) {
	sign := 1.0
	shift := 0.0
	if in.facing < 0 {
		sign = -1
		shift = fntScale
	}
	if in.flip {
		sign = -sign
	}
	signed := in.scale * v.scale * sign
	anchor := (in.pos+v.offset+shift+in.offset)*v.scale + signed*in.frameShift
	extent := math.Abs(signed)
	near := float64(in.axis)
	if signed < 0 {
		near = float64(in.size - in.axis)
	}
	start = (anchor - near*extent) * v.screen
	return start, float64(in.size) * extent * v.screen, signed < 0
}
