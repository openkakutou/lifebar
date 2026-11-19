package lifebar

import (
	"math"
	"testing"
)

func lifeFrom(t *testing.T, section string) PlayerLife {
	t.Helper()
	lb, _ := Parse(section)
	return lb.Life[0]
}

func expectMeter(t *testing.T, tick int, m Meter, top, mid float64) {
	t.Helper()
	if math.Abs(float64(m.Top)-top) > 1e-6 || math.Abs(float64(m.Mid)-mid) > 1e-6 {
		t.Errorf("tick %d: top %v mid %v, want top %v mid %v", tick, m.Top, m.Mid, top, mid)
	}
}

// run steps the meter once per entry: life 1 at tick 1, 0.6 afterwards, with
// getHit on ticks 2 to 4.
func run(t *testing.T, p PlayerLife, ticks int, check map[int][2]float64) {
	t.Helper()
	m := NewMeter()
	for tick := 1; tick <= ticks; tick++ {
		life := float32(0.6)
		if tick == 1 {
			life = 1
		}
		m.Step(p, LifeTick{Life: life, GetHit: tick >= 2 && tick <= 4, HasMid: true})
		if want, ok := check[tick]; ok {
			expectMeter(t, tick, m, want[0], want[1])
		}
	}
}

func TestMeterStep_TrailsAfterTheDelayWithDefaultSettings(t *testing.T) {
	p := lifeFrom(t, "")
	// Defaults: mid.freeze 1, mid.delay 30, mid.mult 1, mid.steps 8.
	// Top halves its gap to 0.6 each tick: 1 -> 0.8 -> 0.7 -> 0.65 -> 0.625.
	// Hit ticks 2-4 (freeze): the first sets delay 30, mid = oldlife = 1.
	// Tick 5 (no hit) delay 29, then one less per tick: delay 0 at tick 34.
	// Tick 34: mid += (0.6 - 1) / 8 -> 0.95; tick 35: 0.95 + (0.6-0.95)/8 =
	// 0.90625; tick 36: 0.90625 + (0.6-0.90625)/8 = 0.86796875.
	run(t, p, 36, map[int][2]float64{
		1: {1, 1}, 2: {0.8, 1}, 3: {0.7, 1}, 4: {0.65, 1}, 5: {0.625, 1},
		24: {0.6 + 0.05/(1<<20), 1}, 33: {0.6, 1},
		34: {0.6, 0.95}, 35: {0.6, 0.90625}, 36: {0.6, 0.86796875},
	})
}

func TestMeterStep_TrailsSoonerWithoutFreezeAndFewerSteps(t *testing.T) {
	p := lifeFrom(t, "[Lifebar]\nmid.freeze = 0\nmid.steps = 4\n")
	// Rising edge at tick 2 sets delay 30 and mid = oldlife = 1, then the
	// delay counts down in that same tick (29) and on the held-hit ticks
	// 3 and 4 (no new edge): delay 29 - (tick-2), 0 at tick 31.
	// Tick 31: mid += (0.6 - 1) / 4 = 0.9; tick 32: 0.9 + (0.6-0.9)/4 =
	// 0.825; tick 33: 0.825 + (0.6-0.825)/4 = 0.76875.
	run(t, p, 33, map[int][2]float64{
		1: {1, 1}, 2: {0.8, 1}, 4: {0.65, 1}, 30: {0.6, 1},
		31: {0.6, 0.9}, 32: {0.6, 0.825}, 33: {0.6, 0.76875},
	})
}

func TestMeterStep_FollowsAtOnceWithoutAMidElement(t *testing.T) {
	m := NewMeter()
	p := lifeFrom(t, "")
	m.Step(p, LifeTick{Life: 1})
	// mid > life and no mid element: mid += (0.6 - 1) / 8 on the first tick.
	m.Step(p, LifeTick{Life: 0.6, GetHit: true})
	expectMeter(t, 2, m, 0.8, 0.95)
	m.Step(p, LifeTick{Life: 0.6})
	expectMeter(t, 3, m, 0.7, 0.90625)
}

func TestMeterStep_RefillFollowsAtOnce(t *testing.T) {
	m := NewMeter()
	m.Top, m.Mid = 0.2, 0.2
	m.Step(lifeFrom(t, ""), LifeTick{Life: 0.6})
	// top jumps to life; mid catches half the gap: 0.2 + (0.6-0.2)/2.
	expectMeter(t, 1, m, 0.6, 0.4)
}

func TestMeterStep_StartsFromTheFirstTicksLife(t *testing.T) {
	m := NewMeter()
	m.Step(lifeFrom(t, ""), LifeTick{Life: 0.5, HasMid: true})
	if m.Top != 0.5 {
		t.Errorf("Top = %v, want 0.5 (toplife starts at 0 in Ikemen GO)", m.Top)
	}
}

func TestMeterStep_DoesNotDivideByAZeroMidSteps(t *testing.T) {
	m := NewMeter()
	m.Step(PlayerLife{}, LifeTick{Life: 0.5})
	// steps read as 1: mid += (0.5 - 1) / 1.
	expectMeter(t, 1, m, 0.5, 0.5)
}

func TestMeterStep_StaysFullAtFullLife(t *testing.T) {
	m := NewMeter()
	p := lifeFrom(t, "")
	for range 50 {
		m.Step(p, LifeTick{Life: 1, HasMid: true})
	}
	expectMeter(t, 50, m, 1, 1)
}

func TestMeterStep_DecaysTheTrailAfterAHitlessDropAndRestartsOnALaterHit(t *testing.T) {
	p := lifeFrom(t, "[Lifebar]\nmid.mult = 2\nmid.delay = 5\n")
	// Defaults otherwise: freeze 1, steps 8. Life per tick: 1, then 0.5
	// (poison, no hit) on ticks 2-4, then 0.4 with a hit on tick 5 only.
	lives := []float32{1, 0.5, 0.5, 0.5, 0.4, 0.4, 0.4, 0.4, 0.4, 0.4}
	// Tick 2: delay 0, life < midMin 1: midMin += (0.5-1) * 1/(12+0.5*144) * 2
	// = 1 - 1/84 = 0.9880952; mid += (0.9880952-1)/8 = 0.9985119; top 0.75.
	// Tick 3: midMin += -0.4880952 * 1/(12+70.2857) * 2 = 0.9762318;
	// mid += (0.9762318-0.9985119)/8 = 0.9957269; top 0.625.
	// Tick 4: midMin += -0.4762318 * 1/(12+68.5774) * 2 = 0.9644113;
	// mid += (0.9644113-0.9957269)/8 = 0.9918125; top 0.5625.
	// Tick 5 (hit, freeze): delay 5, mid = midMin = oldlife 0.5; top 0.48125.
	// Ticks 6-9: delay 4..1, midMin follows life 0.4, mid stays 0.5.
	// Tick 10: delay 0, midMin = 0.4 = life: mid += (0.4-0.5)/8 = 0.4875.
	want := map[int][2]float64{
		2: {0.75, 0.9985119}, 3: {0.625, 0.9957269}, 4: {0.5625, 0.9918125},
		5: {0.48125, 0.5}, 6: {0.440625, 0.5}, 9: {0.4050781, 0.5},
		10: {0.4025391, 0.4875},
	}
	m := NewMeter()
	for i, life := range lives {
		tick := i + 1
		m.Step(p, LifeTick{Life: life, GetHit: tick == 5, HasMid: true})
		if w, ok := want[tick]; ok {
			expectMeter(t, tick, m, w[0], w[1])
		}
	}
}
