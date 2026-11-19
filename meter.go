package lifebar

// Meter is the fill state of a life bar, stepped once per tick (Ikemen GO's
// toplife, midlife, midlifeMin, oldlife, mlifetime and gethit, all
// LifeBar fields, fightscreen.go:448-465). Top is the front fill and Mid the
// trailing one, both fractions 0..1.
//
// It is float32 like Ikemen GO: the trailing fill is an iteration whose
// rounding would drift from the engine's in float64.
type Meter struct {
	Top, Mid    float32
	old, midMin float32
	delay       int
	wasHit      bool
}

// LifeTick is what a bar sees of its fighter on one tick.
type LifeTick struct {
	// Life is the fighter's life over its maximum, 0..1.
	Life float32
	// GetHit is Ikemen GO's getHit: the fighter has received hits or is in
	// a hit state.
	GetHit bool
	// HasMid reports whether the bar's mid element has frames; without
	// them the trailing fill follows the front one at once.
	HasMid bool
}

// NewMeter is the state of a bar at the start of a round, as newLifeBar
// leaves it (fightscreen.go:471): the front fill is set from the first
// tick's life.
func NewMeter() Meter { return Meter{Mid: 1, old: 1, midMin: 1} }

// Step advances one tick, as LifeBar.step does (fightscreen.go:554-603) with
// the bar's own mid.freeze, mid.delay, mid.mult and mid.steps.
func (m *Meter) Step(p PlayerLife, t LifeTick) {
	// Parse keeps MidSteps at 1 or more as Ikemen GO does; a PlayerLife
	// built by hand could hold 0 and divide by it.
	steps := max(1, p.MidSteps)
	if m.Top > t.Life {
		m.Top += (t.Life - m.Top) / 2
	} else {
		m.Top = t.Life
	}
	if !p.MidFreeze && t.GetHit && !m.wasHit && t.HasMid {
		m.startDelay(p.MidDelay)
	}
	m.wasHit = t.GetHit
	if p.MidFreeze && t.GetHit && t.HasMid {
		if m.delay < p.MidDelay {
			m.startDelay(p.MidDelay)
		}
	} else {
		if m.delay > 0 {
			m.delay--
		}
		if t.HasMid && m.delay <= 0 && t.Life < m.midMin {
			m.midMin += (t.Life - m.midMin) * (1 / (12 - (t.Life-m.midMin)*144)) * p.MidMult
			m.midMin = max(m.midMin, t.Life)
		} else {
			m.midMin = t.Life
		}
		if (!t.HasMid || m.delay <= 0) && m.Mid > m.midMin {
			m.Mid += (m.midMin - m.Mid) / steps
		}
		m.old = t.Life
	}
	floor := max(m.midMin, t.Life)
	if m.Mid < floor {
		m.Mid += (floor - m.Mid) / 2
	}
}

// startDelay freezes the trailing fill at the life before the hit.
func (m *Meter) startDelay(delay int) {
	m.delay = delay
	m.Mid = m.old
	m.midMin = m.old
}
