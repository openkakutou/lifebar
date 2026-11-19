package lifebar

import (
	"math"
	"testing"

	"github.com/openkakutou/character/air"
)

func approx(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("%s = %v, want %v", what, got, want)
	}
}

// view320 shows a lifebar of the given localcoord on a 320x240 game drawn at
// 320x240 pixels.
func view320(localcoord [2]int) Screen {
	lb := Lifebar{Localcoord: localcoord}
	return lb.Screen([2]int{320, 240}, [2]int{320, 240})
}

func TestScreen_ScalesFromLocalcoordOnA4By3Game(t *testing.T) {
	s := view320([2]int{1280, 720})
	// viewport43 = 960 wide, viewport = 960 wide: calcScale 0.75,
	// scale = 320/960*0.75, offsetX = 0, portrait = 1280/960*0.75.
	approx(t, "Scale", s.Scale, 0.25)
	approx(t, "OffsetX", s.OffsetX, 0)
	approx(t, "PortraitScale", s.PortraitScale, 1)
}

func TestScreen_ScalesAndOffsetsALifebarOnA16By9Game(t *testing.T) {
	lb := Lifebar{Localcoord: [2]int{1280, 720}}
	s := lb.Screen([2]int{427, 240}, [2]int{854, 480})
	// viewport43 = 960 wide; 1280*240 < 720*427 so viewport = 1280 wide:
	// calcScale 1, calcOffsetX = (960-1280)/2 = -160.
	approx(t, "Scale", s.Scale, 1.0/3)
	approx(t, "OffsetX", s.OffsetX, -160)
	approx(t, "PortraitScale", s.PortraitScale, 4.0/3)
}

func TestScreen_FallsBackTo320By240WhenSizesAreZero(t *testing.T) {
	lb := Lifebar{}
	s := lb.Screen([2]int{}, [2]int{})
	approx(t, "Scale", s.Scale, 1)
	approx(t, "OffsetX", s.OffsetX, 0)
	approx(t, "PortraitScale", s.PortraitScale, 1)
}

func TestFillSpan_MatchesCalcBarFillRect(t *testing.T) {
	unit := axisView{scale: 1, screen: 1}
	// Ascending: starts at pos+r0, grows right.
	if s, n := fillSpan(0, [2]int{0, 99}, 1, unit); s != 0 || n != 100 {
		t.Errorf("full ascending = %d,%d", s, n)
	}
	if s, n := fillSpan(10, [2]int{0, 99}, 0.5, unit); s != 10 || n != 50 {
		t.Errorf("half ascending = %d,%d", s, n)
	}
	// Descending: ends at pos+r1+1, 429*0.25*0.5+0.5 = 54.1, 564*0.25+0.5 = 141.5.
	if s, n := fillSpan(0, [2]int{563, 135}, 0.5, axisView{scale: 0.25, screen: 1}); s != 87 || n != 54 {
		t.Errorf("half descending = %d,%d", s, n)
	}
}

func TestFillSpan_IsEmptyForAnEmptyBar(t *testing.T) {
	if _, n := fillSpan(0, [2]int{0, 99}, 0, axisView{scale: 1, screen: 1}); n != 0 {
		t.Errorf("empty size = %d", n)
	}
}

func TestFillSpan_FillsOnePixelForAZeroRange(t *testing.T) {
	// A range of one value spans r1-r0+1 = 1 unit.
	if s, n := fillSpan(5, [2]int{3, 3}, 1, axisView{scale: 1, screen: 1}); s != 8 || n != 1 {
		t.Errorf("single-value range = %d,%d", s, n)
	}
}

func TestClips_AscendingTrailStartsAfterFront(t *testing.T) {
	b := Bar{RangeX: [2]int{0, 99}}
	front, trail := b.Clips(view320([2]int{320, 240}), 0.5, 0.8)
	if front != (Rect{0, 0, 50, 240}) {
		t.Errorf("front = %+v", front)
	}
	if trail != (Rect{50, 0, 30, 240}) {
		t.Errorf("trail = %+v", trail)
	}
}

func TestClips_DescendingTrailKeepsItsLeftPart(t *testing.T) {
	b := Bar{RangeX: [2]int{99, 0}}
	front, trail := b.Clips(view320([2]int{320, 240}), 0.5, 0.8)
	if front != (Rect{50, 0, 50, 240}) || trail != (Rect{20, 0, 30, 240}) {
		t.Errorf("front %+v trail %+v", front, trail)
	}
}

func TestClips_VerticalAndUnranged(t *testing.T) {
	b := Bar{RangeY: [2]int{0, 9}}
	front, trail := b.Clips(view320([2]int{160, 120}), 0.5, 1)
	if front != (Rect{0, 0, 320, 10}) || trail != (Rect{0, 10, 320, 10}) {
		t.Errorf("vertical front %+v trail %+v", front, trail)
	}
	front, trail = Bar{}.Clips(view320([2]int{320, 240}), 0.3, 0.6)
	if front != (Rect{0, 0, 320, 240}) || trail.W != 0 {
		t.Errorf("no range: front %+v trail %+v", front, trail)
	}
}

func TestClips_CentersABarOnAWiderGame(t *testing.T) {
	lb := Lifebar{Localcoord: [2]int{320, 240}}
	s := lb.Screen([2]int{427, 240}, [2]int{854, 480})
	b := Bar{Pos: [2]int{10, 0}, RangeX: [2]int{0, 99}}
	front, _ := b.Clips(s, 0.5, 0.5)
	// mid = (427-320)/2 = 53.5 game units; start = (10*1+53.5)*2+0.5 = 127.5
	// -> 127; size = 100*1*0.5*2+0.5 = 100.5 -> 100; the rows span the view.
	if front != (Rect{127, 0, 100, 480}) {
		t.Errorf("front = %+v", front)
	}
}

var spriteBox = SpriteBox{Axis: [2]int{3, 5}, Size: [2]int{20, 8}}

func TestPlace_UsesAxisOffsetAndFrameShift(t *testing.T) {
	e := Element{Layout: Layout{Offset: [2]float32{10, 4}, Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	s := view320([2]int{640, 480}) // scale 0.5
	got := s.Place(e, air.Frame{X: 2, Y: 1}, spriteBox, [2]int{100, 0})
	// x = (100+10)*0.5 + 0.5*2 - 3*0.5 = 54.5; y = 4*0.5 + 0.5*1 - 5*0.5 = 0.
	want := Placement{Rect: Rect{54.5, 0, 10, 4}}
	if got != want {
		t.Errorf("right-facing = %+v, want %+v", got, want)
	}
}

func TestPlace_MirroredElementMirrorsTheFrameOffsetAndShiftsByOneUnit(t *testing.T) {
	e := Element{Layout: Layout{Offset: [2]float32{10, 4}, Facing: -1, VFacing: 1, Scale: [2]float32{1, 1}}}
	s := view320([2]int{640, 480})
	got := s.Place(e, air.Frame{X: 2, Y: 1}, spriteBox, [2]int{100, 0})
	// anchor = (100+1+10)*0.5 - 0.5*2 = 54.5 (common.go:1462 adds the unit);
	// the axis sits at the far side: 54.5 - (20-3)*0.5 = 46.
	want := Placement{Rect: Rect{46, 0, 10, 4}, FlipX: true}
	if got != want {
		t.Errorf("mirrored = %+v, want %+v", got, want)
	}
}

func TestPlace_FlippedFrameOnAMirroredElementIsUnmirrored(t *testing.T) {
	e := Element{Layout: Layout{Offset: [2]float32{10, 4}, Facing: -1, VFacing: 1, Scale: [2]float32{1, 1}}}
	got := view320([2]int{320, 240}).Place(e, air.Frame{X: 2, Flip: air.FlipH}, spriteBox, [2]int{})
	// sign = -1 * -1 = +1: anchor = (0+1+10) + 2 = 13, left = 13 - 3 = 10.
	if got.FlipX || got.X != 10 {
		t.Errorf("got %+v, want unmirrored at x 10", got)
	}
}

func TestPlace_FlippedFrameMirrorsItsOffset(t *testing.T) {
	e := Element{Layout: Layout{Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	got := view320([2]int{320, 240}).Place(e, air.Frame{X: 2, Flip: air.FlipH}, spriteBox, [2]int{100, 0})
	// signed factor -1: anchor = 100 - 2 = 98, left = 98 - (20-3) = 81
	// (anim.go:1091-1092 multiplies the offset by the flipped scale).
	if !got.FlipX || got.X != 81 {
		t.Errorf("got %+v, want mirrored at x 81", got)
	}
}

func TestPlace_VerticalFacingMirrorsAroundTheAnchor(t *testing.T) {
	e := Element{Layout: Layout{Facing: 1, VFacing: -1, Scale: [2]float32{1, 1}}}
	got := view320([2]int{320, 240}).Place(e, air.Frame{Y: 1}, spriteBox, [2]int{})
	// anchor = (0+1) - 1 = 0; the axis sits at the far side: 0 - (8-5) = -3.
	if got.Y != -3 || got.H != 8 || !got.FlipY || got.FlipX {
		t.Errorf("got %+v", got)
	}
}

func TestPlace_CentersOnAWiderGame(t *testing.T) {
	lb := Lifebar{Localcoord: [2]int{320, 240}}
	s := lb.Screen([2]int{427, 240}, [2]int{427, 240})
	e := Element{Layout: Layout{Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	got := s.Place(e, air.Frame{}, SpriteBox{Size: [2]int{5, 5}}, [2]int{10, 0})
	if got.X != 63.5 { // 10 + (427-320)/2
		t.Errorf("X = %v, want 63.5", got.X)
	}
}

func TestPlace_ZeroScaleElementHasNoArea(t *testing.T) {
	got := view320([2]int{320, 240}).Place(Element{}, air.Frame{X: 4}, spriteBox, [2]int{7, 7})
	if got.W != 0 || got.H != 0 || math.IsNaN(got.X) || math.IsNaN(got.Y) {
		t.Errorf("got %+v, want a zero-size rect", got)
	}
}

func TestFillSpan_RoundsInFloat32LikeIkemen(t *testing.T) {
	// float32(0.005) is 0.0049999999; times 100 it rounds up to exactly 0.5 in
	// float32, so size = int(0.5+0.5) = 1. In float64 the product stays below
	// 0.5 and the size would be 0.
	if _, n := fillSpan(0, [2]int{0, 99}, float32(0.005), axisView{scale: 1, screen: 1}); n != 1 {
		t.Errorf("size = %d, want 1", n)
	}
	// scale float32(1/3) = 0.33333334, offset -160, mid 53.5, pos 1:
	// (1-160)*0.33333334 = -53.0 in float32 (-53.0000015 rounds to -53), so
	// start = int(0.5+0.5) = 1; in float64 it is 0.49999998+0.5, giving 0.
	v := axisView{offset: -160, scale: float64(float32(1) / 3), mid: 53.5, screen: 1}
	if s, _ := fillSpan(1, [2]int{0, 99}, 1, v); s != 1 {
		t.Errorf("start = %d, want 1", s)
	}
}

func approxRect(t *testing.T, what string, got, want Rect) {
	t.Helper()
	approx(t, what+".X", got.X, want.X)
	approx(t, what+".Y", got.Y, want.Y)
	approx(t, what+".W", got.W, want.W)
	approx(t, what+".H", got.H, want.H)
}

// wide shows a 1280x720 lifebar on a 427x240 game drawn at 854x480 pixels:
// scale 1/3, offsetX -160, 53.5 units of centering, 2 pixels per unit.
func wide() Screen {
	lb := Lifebar{Localcoord: [2]int{1280, 720}}
	return lb.Screen([2]int{427, 240}, [2]int{854, 480})
}

func TestClips_AppliesTheLifebarOffsetOnAWideLocalcoord(t *testing.T) {
	b := Bar{Pos: [2]int{30, 0}, RangeX: [2]int{0, 99}}
	front, _ := b.Clips(wide(), 0.5, 0.5)
	// start = ((30-160)/3 + 53.5)*2 + 0.5 = 20.83 -> 20;
	// size = 100/3*0.5*2 + 0.5 = 33.83 -> 33; the rows span the view.
	if front != (Rect{20, 0, 33, 480}) {
		t.Errorf("front = %+v", front)
	}
}

func TestPlace_AppliesTheLifebarOffsetOnAWideLocalcoord(t *testing.T) {
	e := Element{Layout: Layout{Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	got := wide().Place(e, air.Frame{}, spriteBox, [2]int{30, 0})
	// x = ((30-160)/3 - 3/3)*2 + 53.5*2 = 18.333; y = -5/3*2 = -3.333;
	// w = 20/3*2 = 13.333; h = 8/3*2 = 5.333.
	approxRect(t, "rect", got.Rect, Rect{55.0 / 3, -10.0 / 3, 40.0 / 3, 16.0 / 3})
	if got.FlipX || got.FlipY {
		t.Errorf("got %+v, want unmirrored", got)
	}
}

func TestPlace_ScalesByThePixelsPerGameUnit(t *testing.T) {
	e := Element{Layout: Layout{Offset: [2]float32{10, 4}, Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	lb := Lifebar{Localcoord: [2]int{320, 240}}
	s := lb.Screen([2]int{320, 240}, [2]int{640, 480})
	got := s.Place(e, air.Frame{X: 2, Y: 1}, spriteBox, [2]int{100, 0})
	// x = (100+10+2-3)*2 = 218; y = (4+1-5)*2 = 0; w = 20*2; h = 8*2.
	if got != (Placement{Rect: Rect{218, 0, 40, 16}}) {
		t.Errorf("got %+v", got)
	}
}

// tall shows a 320x300 lifebar on a 320x300 game at 1 pixel per unit: scale 1,
// no x offset, and 60 units of height beyond 320x240.
func tall() Screen {
	lb := Lifebar{Localcoord: [2]int{320, 300}}
	return lb.Screen([2]int{320, 300}, [2]int{320, 300})
}

func TestPlace_ShiftsYByTheGameHeightBeyond240(t *testing.T) {
	e := Element{Layout: Layout{Offset: [2]float32{10, 4}, Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	got := tall().Place(e, air.Frame{X: 2, Y: 1}, spriteBox, [2]int{100, 0})
	// x = 100+10+2-3 = 109; y = (0 + 60) + 4 + 1 - 5 = 60 (common.go:1472).
	if got != (Placement{Rect: Rect{109, 60, 20, 8}}) {
		t.Errorf("got %+v", got)
	}
}

func TestClips_CentersAYRangedBarOnATallerGame(t *testing.T) {
	b := Bar{Pos: [2]int{0, 10}, RangeY: [2]int{0, 9}}
	front, trail := b.Clips(tall(), 0.5, 1)
	// midY = (300-240)/2 = 30: start = 10+30 = 40; front h = int(10*0.5+0.5) = 5;
	// the trail starts after the front (45) and keeps 10-5 = 5. Columns span the view.
	if front != (Rect{0, 40, 320, 5}) || trail != (Rect{0, 45, 320, 5}) {
		t.Errorf("front %+v trail %+v", front, trail)
	}
}

func TestPlace_VerticallyFlippedFrameMirrorsOnlyY(t *testing.T) {
	e := Element{Layout: Layout{Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	got := view320([2]int{320, 240}).Place(e, air.Frame{X: 2, Y: 1, Flip: air.FlipV}, spriteBox, [2]int{100, 10})
	// x = 100+2-3 = 99. y: signed -1, anchor = 10 - 1 = 9, the axis sits at
	// the far side: 9 - (8-5) = 6 (anim.go:1091-1092).
	if got != (Placement{Rect: Rect{99, 6, 20, 8}, FlipY: true}) {
		t.Errorf("got %+v", got)
	}
}

func TestPlace_FlippedBothWaysMirrorsXAndY(t *testing.T) {
	e := Element{Layout: Layout{Facing: 1, VFacing: 1, Scale: [2]float32{1, 1}}}
	got := view320([2]int{320, 240}).Place(e, air.Frame{X: 2, Y: 1, Flip: air.FlipHV}, spriteBox, [2]int{100, 10})
	// x: anchor = 100 - 2 = 98, left = 98 - (20-3) = 81; y as above: 6.
	if got != (Placement{Rect: Rect{81, 6, 20, 8}, FlipX: true, FlipY: true}) {
		t.Errorf("got %+v", got)
	}
}
