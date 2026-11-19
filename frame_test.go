package lifebar

import (
	"testing"

	"github.com/openkakutou/character/air"
)

func frames(times ...int) []air.Frame {
	out := make([]air.Frame, len(times))
	for i, tm := range times {
		out[i] = air.Frame{Group: 1, Image: i, Time: tm}
	}
	return out
}

// lifebarWith holds animation 1 and an element playing it.
func lifebarWith(a air.Animation) (*Lifebar, Element) {
	return &Lifebar{Animations: map[int]air.Animation{1: a}}, Element{HasAnim: true, Anim: 1}
}

func imageAt(t *testing.T, lb *Lifebar, e Element, tick int) int {
	t.Helper()
	fr, ok := lb.Frame(e, tick)
	if !ok {
		t.Fatalf("tick %d: no frame", tick)
	}
	return fr.Image
}

func TestFrame_FollowsTheAnimationAndLoops(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: []air.Frame{
		{Group: 11, Image: 3, Time: 10}, {Group: 11, Image: 4, Time: 5},
	}})
	cases := map[int][2]int{0: {11, 3}, 9: {11, 3}, 10: {11, 4}, 15: {11, 3}, 27: {11, 4}}
	for tick, want := range cases {
		fr, ok := lb.Frame(e, tick)
		if !ok || [2]int{fr.Group, fr.Image} != want {
			t.Errorf("tick %d: %v %v, want %v", tick, fr, ok, want)
		}
	}
}

func TestFrame_ShowsASpriteElementForever(t *testing.T) {
	lb := &Lifebar{}
	e := Element{HasSprite: true, Sprite: SpriteRef{Group: 11, Image: 4}}
	if fr, ok := lb.Frame(e, 500); !ok || fr.Group != 11 || fr.Image != 4 {
		t.Errorf("spr element = %v %v", fr, ok)
	}
}

func TestFrame_AnimationReplacesTheSprite(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: frames(5)})
	e.HasSprite, e.Sprite = true, SpriteRef{Group: 9, Image: 9}
	if got := imageAt(t, lb, e, 0); got != 0 {
		t.Errorf("image = %d, want the animation's frame 0", got)
	}
}

func TestFrame_ShowsNothingForAnUnsetElement(t *testing.T) {
	if _, ok := (&Lifebar{}).Frame(Element{}, 0); ok {
		t.Error("an unset element has no frame")
	}
}

func TestFrame_ShowsNothingForAnEmptyOrMissingAnimation(t *testing.T) {
	lb, e := lifebarWith(air.Animation{})
	if _, ok := lb.Frame(e, 3); ok {
		t.Error("an animation without frames shows nothing")
	}
	e.Anim = 7
	if _, ok := lb.Frame(e, 3); ok {
		t.Error("an action the lifebar does not have shows nothing")
	}
}

func TestFrame_ShowsNothingOnABlankFrame(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: []air.Frame{{Group: -1, Image: 0, Time: 5}}})
	if _, ok := lb.Frame(e, 0); ok {
		t.Error("a -1 group frame shows nothing")
	}
	sprite := Element{HasSprite: true, Sprite: SpriteRef{Group: 0, Image: -1}}
	if _, ok := lb.Frame(sprite, 0); ok {
		t.Error("a -1 image sprite shows nothing")
	}
}

func TestFrame_HoldsALastFrameOfTimeMinusOne(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: frames(3, -1)})
	want := map[int]int{0: 0, 2: 0, 3: 1, 1000: 1}
	for tick, image := range want {
		if got := imageAt(t, lb, e, tick); got != image {
			t.Errorf("tick %d: image %d, want %d", tick, got, image)
		}
	}
}

func TestFrame_LoopsFromLoopStart(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: frames(2, 3, 4), LoopStart: 1})
	// Total 9, loop 7: tick 9 wraps to 2 (frame 1), tick 15 to 8 (frame 2).
	want := map[int]int{0: 0, 2: 1, 5: 2, 8: 2, 9: 1, 15: 2, 16: 1}
	for tick, image := range want {
		if got := imageAt(t, lb, e, tick); got != image {
			t.Errorf("tick %d: image %d, want %d", tick, got, image)
		}
	}
}

func TestFrame_LoopsFromTheFirstFrameWhenLoopStartIsBeyondTheFrames(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: frames(2, 3), LoopStart: 5})
	if got := imageAt(t, lb, e, 5); got != 0 {
		t.Errorf("image = %d, want 0 (anim.go:440 resets a loopstart past the frames)", got)
	}
}

func TestFrame_SkipsFramesWithoutDuration(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: frames(0, 2, 0)})
	for _, tick := range []int{0, 1, 2, 3} {
		if got := imageAt(t, lb, e, tick); got != 1 {
			t.Errorf("tick %d: image %d, want 1", tick, got)
		}
	}
}

func TestFrame_ShowsTheLastFrameWhenNothingHasADuration(t *testing.T) {
	lb, e := lifebarWith(air.Animation{Frames: frames(0, 0)})
	if got := imageAt(t, lb, e, 10); got != 1 {
		t.Errorf("image = %d, want 1", got)
	}
}
