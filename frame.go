package lifebar

import "github.com/openkakutou/character/air"

// Frame returns the frame an element shows ticks ticks after its owner
// started, and false when it shows nothing: no sprite, an animation without
// frames, or a blank frame (Animation.isBlank, anim.go:604). A sprite is a
// single frame; an animation replaces it when both are set.
func (lb *Lifebar) Frame(e Element, ticks int) (air.Frame, bool) {
	var fr air.Frame
	switch {
	case e.HasAnim:
		var ok bool
		fr, ok = animFrame(lb.Animations[e.Anim], ticks)
		if !ok {
			return air.Frame{}, false
		}
	case e.HasSprite:
		fr = air.Frame{Group: e.Sprite.Group, Image: e.Sprite.Image}
	default:
		return air.Frame{}, false
	}
	if fr.IsBlank() {
		return air.Frame{}, false
	}
	return fr, true
}

// animFrame is the frame shown ticks ticks into a playback, as
// Animation.Action and animSeek step it (anim.go:923, 753): frames of time 0
// or less are skipped, except a last frame of time -1, which holds forever
// (anim.go:445); otherwise the playback returns to LoopStart once every
// frame has played. A LoopStart outside the frames is 0 (anim.go:440).
func animFrame(a air.Animation, ticks int) (air.Frame, bool) {
	last := len(a.Frames) - 1
	if last < 0 {
		return air.Frame{}, false
	}
	holds := a.Frames[last].Time == -1
	total := playTime(a.Frames)
	t := max(ticks, 0)
	if t >= total && !holds {
		loopStart := a.LoopStart
		if loopStart < 0 || loopStart > last {
			loopStart = 0
		}
		loop := playTime(a.Frames[loopStart:])
		if loop == 0 {
			return a.Frames[last], true
		}
		start := total - loop
		t = start + (t-start)%loop
	}
	elapsed := 0
	for _, fr := range a.Frames {
		elapsed += max(fr.Time, 0)
		if fr.Time > 0 && t < elapsed {
			return fr, true
		}
	}
	return a.Frames[last], true
}

// playTime is the ticks one pass over frames lasts.
func playTime(frames []air.Frame) int {
	total := 0
	for _, fr := range frames {
		total += max(fr.Time, 0)
	}
	return total
}
