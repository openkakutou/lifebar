package lifebar

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/openkakutou/character/air"
)

// defaultLocalcoord is the coordinate space of Ikemen GO's default motif,
// which a fight.def without [Info] localcoord inherits.
var defaultLocalcoord = [2]int{320, 240}

// modelledSections reads the sections Parse models into a Lifebar, by their
// lowercased header name. [Info] is read by Parse itself.
var modelledSections = map[string]func(r reader, lb *Lifebar, s iniSection){
	"files": func(_ reader, lb *Lifebar, s iniSection) {
		lb.Files = readFiles(s)
	},
	"lifebar": func(r reader, lb *Lifebar, s iniSection) {
		for i := range lb.Life {
			lb.Life[i] = r.playerLife(s.forPlayer("p", i+1), playerPrefix(i))
		}
	},
	"powerbar": func(r reader, lb *Lifebar, s iniSection) {
		for i := range lb.Power {
			lb.Power[i] = r.powerBar(s.forPlayer("p", i+1), playerPrefix(i))
		}
	},
	"face": func(r reader, lb *Lifebar, s iniSection) {
		for i := range lb.Face {
			lb.Face[i] = r.face(s.forPlayer("p", i+1), playerPrefix(i))
		}
	},
	"name": func(r reader, lb *Lifebar, s iniSection) {
		for i := range lb.Name {
			lb.Name[i] = r.name(s.forPlayer("p", i+1), playerPrefix(i))
		}
	},
	"winicon": func(r reader, lb *Lifebar, s iniSection) {
		for i := range lb.WinIcon {
			lb.WinIcon[i] = r.winIcon(s.forPlayer("p", i+1), playerPrefix(i))
		}
	},
	"time": func(r reader, lb *Lifebar, s iniSection) {
		lb.Time = r.time(s.keys)
	},
	"combo": func(r reader, lb *Lifebar, s iniSection) {
		for i := range lb.Combo {
			lb.Combo[i] = r.combo(s, i+1)
		}
	},
	"action": func(r reader, lb *Lifebar, s iniSection) {
		for i := range lb.Action {
			lb.Action[i] = r.action(s.forPlayer("team", i+1), i+1)
		}
	},
	"round": func(r reader, lb *Lifebar, s iniSection) {
		lb.Round = r.round(s.keys)
	},
}

// unmodelledSections are the sections Ikemen GO's loader reads that Parse
// does not model yet. They are not reported as ignored.
var unmodelledSections = map[string]bool{
	"fightfx": true, "guardbar": true, "stunbar": true, "timer": true,
	"score": true, "match": true, "ailevel": true, "wincount": true, "mode": true,
	"turns ": true, "simul ": true, "simul_3p ": true, "simul_4p ": true,
	"tag ": true, "tag_3p ": true, "tag_4p ": true,
}

// Parse reads a fight.def as Ikemen GO's loadFightScreen does (src/fightscreen.go
// @ 54d4148a). It never fails on content: malformed lines are read the way
// Ikemen GO reads them, and the sections its loader skips come back as
// warnings. Only the first section of each kind is read; a modelled section
// the file lacks keeps the defaults of an empty one.
func Parse(src string) (Lifebar, []Warning) {
	sections := tokenize(src)
	lb := Lifebar{Localcoord: readLocalcoord(sections)}
	var warnings []Warning
	lb.Animations, warnings = readAnimations(sections)
	// Ikemen GO picks the combo syntax from the [Info] ikemenversion seen so far.
	r := reader{animations: lb.Animations, localcoordWidth: lb.Localcoord[0], legacyCombo: true}

	read := map[string]bool{}
	for _, s := range sections {
		if s.name == "info" {
			r.legacyCombo = isVersionZero(s.keys["ikemenversion"])
			continue
		}
		readSection, modelled := modelledSections[s.name]
		if !modelled {
			if !unmodelledSections[s.name] && !isBeginAction(s) {
				warnings = append(warnings, Warning{Section: s.inner, Line: s.line})
			}
			continue
		}
		if !read[s.name] {
			read[s.name] = true
			readSection(r, &lb, s)
		}
	}
	for name, readSection := range modelledSections {
		if !read[name] {
			readSection(r, &lb, iniSection{keys: iniKeys{}})
		}
	}
	slices.SortFunc(warnings, func(a, b Warning) int { return a.Line - b.Line })
	return lb, warnings
}

func playerPrefix(index int) string {
	return fmt.Sprintf("p%d.", index+1)
}

// readLocalcoord reads [Info] localcoord from the first [Info] section only
// (fightscreen.go:4883-4893).
func readLocalcoord(sections []iniSection) [2]int {
	localcoord := defaultLocalcoord
	for _, s := range sections {
		if s.name == "info" {
			s.keys.ints("localcoord", &localcoord[0], &localcoord[1])
			break
		}
	}
	return localcoord
}

// isVersionZero reports whether an ikemenversion value has a major and a
// minor number of 0, or stops parsing before a non-zero one (Ikemen GO
// ParseIkemenVersion, common.go:931).
func isVersionZero(version string) bool {
	parts := strings.Split(version, ".")
	for i := 0; i < len(parts) && i < 2; i++ {
		number, err := strconv.ParseUint(strings.TrimSpace(parts[i]), 10, 16)
		if err != nil {
			return true
		}
		if number != 0 {
			return false
		}
	}
	return true
}

func isBeginAction(s iniSection) bool {
	return s.name == "begin " && strings.HasPrefix(strings.ToLower(s.subname), "action")
}

// readAnimations parses each [Begin Action N] block alone, so a corrupt block
// costs only itself and comes back as a warning. The first block of a number
// wins, as in Ikemen GO (anim.go:1334). A block without frames takes the
// frames of the next block in file order, chained, as Ikemen GO's
// compileAction does (anim.go:1344); one still empty at the end of the file
// stays empty. A "Copy Action N" block takes action N's frames, through up to
// 8 copies, and is dropped when it cannot (resolveCopyAction,
// anim.go:1386-1424).
func readAnimations(sections []iniSection) (map[int]air.Animation, []Warning) {
	var warnings []Warning
	var blocks []animBlock
	for _, s := range sections {
		if !isBeginAction(s) {
			continue
		}
		if b, ok := copyBlock(s); ok {
			blocks = append(blocks, b)
			continue
		}
		parsed, err := air.Parse(strings.NewReader(strings.Join(s.lines, "\n")))
		if err != nil {
			warnings = append(warnings, Warning{Section: s.inner, Line: s.line, Reason: err.Error()})
			continue
		}
		for _, a := range parsed {
			blocks = append(blocks, animBlock{anim: a, copyOf: -1})
		}
	}
	compiled := map[int]animBlock{}
	var compile func(i int) animBlock
	compile = func(i int) animBlock {
		b := blocks[i]
		if existing, taken := compiled[b.anim.Number]; taken {
			return existing
		}
		if len(b.anim.Frames) == 0 && b.copyOf < 0 && i+1 < len(blocks) {
			next := compile(i + 1)
			b.anim.Frames, b.anim.LoopStart, b.copyOf = next.anim.Frames, next.anim.LoopStart, next.copyOf
		}
		compiled[b.anim.Number] = b
		return b
	}
	for i := range blocks {
		compile(i)
	}
	return resolveCopies(compiled), warnings
}

// animBlock is a [Begin Action N] block: its animation, or the action it
// copies (-1 for none).
type animBlock struct {
	anim   air.Animation
	copyOf int
}

// copyBlock reads a block whose first line is "Copy Action N".
func copyBlock(s iniSection) (animBlock, bool) {
	for _, line := range s.lines {
		line = strings.ToLower(strings.TrimSpace(strings.SplitN(line, ";", 2)[0]))
		if line == "" || strings.HasPrefix(line, "[") {
			continue // blank, or the block's header
		}
		if !strings.HasPrefix(line, "copy action ") {
			return animBlock{}, false
		}
		number, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(strings.ToLower(s.subname), "action")))
		if err != nil {
			return animBlock{}, false
		}
		target, err := strconv.Atoi(strings.TrimSpace(line[len("copy action "):]))
		if err != nil {
			target = -1 // an invalid number copies nothing
		}
		return animBlock{anim: air.Animation{Number: number}, copyOf: target}, true
	}
	return animBlock{}, false
}

// resolveCopies gives each copy block its target's frames, dropping those
// whose target is missing, itself or more than 8 copies away.
func resolveCopies(blocks map[int]animBlock) map[int]air.Animation {
	animations := make(map[int]air.Animation, len(blocks))
	for number, b := range blocks {
		target := b.copyOf
		for range 8 {
			if target < 0 {
				break
			}
			t, ok := blocks[target]
			if !ok || target == number {
				break
			}
			if t.copyOf >= 0 {
				target = t.copyOf
				continue
			}
			b.anim.Frames, b.anim.LoopStart = t.anim.Frames, t.anim.LoopStart
			target = -1
		}
		if b.copyOf >= 0 && target >= 0 {
			continue // unresolved: Ikemen GO drops it
		}
		animations[number] = b.anim
	}
	return animations
}

// readFiles reads the file references of [Files] (fightscreen.go:4931).
// Ikemen GO iterates a map for the fontN keys, so with both font1 and font01
// the result there is random; the last one in the file wins here.
func readFiles(s iniSection) Files {
	files := Files{
		Sprites:           s.keys["sff"],
		Sounds:            s.keys["snd"],
		FightFxSprites:    s.keys["fightfx.sff"],
		FightFxAnimations: s.keys["fightfx.air"],
		FightFxSounds:     s.keys["common.snd"],
		Fonts:             map[int]Font{},
	}
	for i := range files.FightFx {
		files.FightFx[i] = s.keys[fmt.Sprintf("fx%d", i+1)]
	}
	declared := map[int]Font{}
	for _, key := range s.order {
		rest, found := strings.CutPrefix(key, "font")
		digits := len(rest) - len(strings.TrimLeft(rest, "0123456789"))
		if !found || digits == 0 {
			continue
		}
		number := atoi(rest[:digits])
		font := declared[number]
		switch rest[digits:] {
		case "":
			font.File = s.keys[key]
		case ".height":
			font.Height = atoi(s.keys[key])
		}
		declared[number] = font
	}
	for number, font := range declared {
		if font.File != "" {
			files.Fonts[number] = font
		}
	}
	return files
}

// reader builds the model elements from the keys of a section.
type reader struct {
	animations      map[int]air.Animation
	localcoordWidth int
	legacyCombo     bool
}

func newLayout(layerNo int) Layout {
	return Layout{Facing: 1, VFacing: 1, LayerNo: layerNo, Scale: [2]float32{1, 1}}
}

// readLayout is Ikemen GO's Layout.Read (common.go:1348) for the keys modelled.
func readLayout(k iniKeys, pre string, layerNo int) Layout {
	l := newLayout(layerNo)
	k.floats(pre+"offset", &l.Offset[0], &l.Offset[1])
	if k[pre+"facing"] != "" {
		l.Facing = signOf(atoi(k[pre+"facing"]))
	}
	if k[pre+"vfacing"] != "" {
		l.VFacing = signOf(atoi(k[pre+"vfacing"]))
	}
	k.ints(pre+"layerno", &l.LayerNo)
	l.LayerNo = min(2, max(l.LayerNo, math.MinInt16))
	k.floats(pre+"scale", &l.Scale[0], &l.Scale[1])
	return l
}

func signOf(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}

// element is Ikemen GO's ReadAnimLayout (common.go:1521): an anim key counts
// only when the action exists.
func (r reader) element(k iniKeys, pre string, layerNo int) Element {
	e := Element{}
	e.HasSprite = k.ints(pre+"spr", &e.Sprite.Group, &e.Sprite.Image)
	var anim int
	if k.ints(pre+"anim", &anim) {
		if _, exists := r.animations[anim]; exists {
			e.HasAnim, e.Anim = true, anim
		}
	}
	e.Layout = readLayout(k, pre, layerNo)
	return e
}

// text is Ikemen GO's readFSText (fightscreen.go:245). fallback is the text
// when the text key is missing; align is the alignment when the font key does
// not give one.
func text(k iniKeys, pre, fallback string, layerNo, align int) Text {
	t := Text{
		Font:   TextFont{No: -1, Align: align, Color: [3]int{-1, -1, -1}, Alpha: 255},
		Text:   fallback,
		Layout: readLayout(k, pre, layerNo),
	}
	f := &t.Font
	k.ints(pre+"font", &f.No, &f.Bank, &f.Align, &f.Color[0], &f.Color[1], &f.Color[2], &f.Alpha)
	if value, ok := k.quoted(pre + "text"); ok {
		t.Text = value
	}
	return t
}

// bar reads what [Lifebar] and [Powerbar] share (fightscreen.go:490, :904).
func (r reader) bar(k iniKeys, pre string) Bar {
	var b Bar
	k.ints(pre+"pos", &b.Pos[0], &b.Pos[1])
	k.ints(pre+"range.x", &b.RangeX[0], &b.RangeX[1])
	k.ints(pre+"range.y", &b.RangeY[0], &b.RangeY[1])
	b.Bg0 = r.element(k, pre+"bg0.", 0)
	b.Bg1 = r.element(k, pre+"bg1.", 0)
	b.Bg2 = r.element(k, pre+"bg2.", 0)
	b.Mid = r.element(k, pre+"mid.", 0)
	b.Front = r.element(k, pre+"front.", 0)
	b.Top = r.element(k, pre+"top.", 0)
	k.flag("leaderontop", &b.LeaderOnTop)
	return b
}

// variants reads the per-value elements pre+name+N.* (and, withMax,
// pre+name+max.*) of a bar (readMultipleValuesF, fightscreen.go:378-396;
// readPBKeys, 911-935), in rising order.
func (r reader) variants(k iniKeys, pre, name string, withMax bool) []Variant {
	seen := map[string]bool{}
	var out []Variant
	for key := range k {
		rest, ok := strings.CutPrefix(key, pre+name)
		if !ok {
			continue
		}
		dot := strings.IndexByte(rest, '.')
		if dot <= 0 || seen[rest[:dot]] {
			continue
		}
		suffix := rest[:dot]
		v := Variant{Element: r.element(k, pre+name+suffix+".", 0)}
		n, err := strconv.Atoi(suffix)
		numeric := err == nil && n >= 0
		if !numeric && !(withMax && suffix == "max") {
			continue
		}
		v.From = n
		v.Max = !numeric
		seen[suffix] = true
		out = append(out, v)
	}
	slices.SortFunc(out, func(a, b Variant) int { return a.From - b.From })
	return out
}

// withVariants is b with its per-value elements.
func (r reader) withVariants(b Bar, k iniKeys, pre string, withMax bool) Bar {
	b.Bg0Variants = r.variants(k, pre, "bg0", withMax)
	b.Bg1Variants = r.variants(k, pre, "bg1", withMax)
	b.Bg2Variants = r.variants(k, pre, "bg2", withMax)
	b.FrontVariants = r.variants(k, pre, "front", withMax)
	return b
}

func (r reader) playerLife(k iniKeys, pre string) PlayerLife {
	l := PlayerLife{Bar: r.withVariants(r.bar(k, pre), k, pre, false), MidFreeze: true, MidDelay: 30, MidMult: 1, MidSteps: 8}
	k.flag("mid.shift", &l.MidShift)
	k.flag("mid.freeze", &l.MidFreeze)
	k.ints("mid.delay", &l.MidDelay)
	k.floats("mid.mult", &l.MidMult)
	k.floats("mid.steps", &l.MidSteps)
	l.MidSteps = max(1, l.MidSteps)
	return l
}

func (r reader) powerBar(k iniKeys, pre string) PowerBar {
	p := PowerBar{Bar: r.withVariants(r.bar(k, pre), k, pre, true), Counter: text(k, pre+"counter.", "%i", 0, 0)}
	k.flag(pre+"levelbars", &p.LevelBars)
	return p
}

// face is readFightScreenFace (fightscreen.go:1870).
func (r reader) face(k iniKeys, pre string) Face {
	f := Face{
		FaceSprite:   SpriteRef{Group: -1},
		FacePalShare: true,
		Teammate:     FaceTeammate{FaceSprite: SpriteRef{Group: -1}, FacePalShare: true},
	}
	k.ints(pre+"pos", &f.Pos[0], &f.Pos[1])
	f.Bg = r.element(k, pre+"bg.", 0)
	f.Bg0 = r.element(k, pre+"bg0.", 0)
	f.Bg1 = r.element(k, pre+"bg1.", 0)
	f.Bg2 = r.element(k, pre+"bg2.", 0)
	f.Top = r.element(k, pre+"top.", 0)
	f.Ko = r.element(k, pre+"ko.", 0)
	k.ints(pre+"face.spr", &f.FaceSprite.Group, &f.FaceSprite.Image)
	f.FaceLayout = readLayout(k, pre+"face.", 0)
	k.flag(pre+"face.palshare", &f.FacePalShare)
	k.flag(pre+"face.palfxshare", &f.FacePalFxShare)
	k.flag(pre+"face.darkenshare", &f.FaceDarkenShare)
	k.flag("leaderontop", &f.LeaderOnTop)

	t := &f.Teammate
	k.ints(pre+"teammate.pos", &t.Pos[0], &t.Pos[1])
	k.ints(pre+"teammate.spacing", &t.Spacing[0], &t.Spacing[1])
	t.Bg = r.element(k, pre+"teammate.bg.", 0)
	t.Bg0 = r.element(k, pre+"teammate.bg0.", 0)
	t.Bg1 = r.element(k, pre+"teammate.bg1.", 0)
	t.Bg2 = r.element(k, pre+"teammate.bg2.", 0)
	t.Top = r.element(k, pre+"teammate.top.", 0)
	t.Ko = r.element(k, pre+"teammate.ko.", 0)
	k.ints(pre+"teammate.face.spr", &t.FaceSprite.Group, &t.FaceSprite.Image)
	t.FaceLayout = readLayout(k, pre+"teammate.face.", 0)
	k.flag(pre+"teammate.ko.hide", &t.KoHide)
	k.flag(pre+"teammate.face.palshare", &t.FacePalShare)
	return f
}

// name is readFightScreenName (fightscreen.go:2125).
func (r reader) name(k iniKeys, pre string) Name {
	n := Name{
		Name: text(k, pre+"name.", "", 0, 0),
		Bg:   r.element(k, pre+"bg.", 0),
		Top:  r.element(k, pre+"top.", 0),
		Teammate: NameTeammate{
			Name: text(k, pre+"teammate.name.", "", 0, 0),
			Bg:   r.element(k, pre+"teammate.bg.", 0),
		},
	}
	k.ints(pre+"pos", &n.Pos[0], &n.Pos[1])
	k.flag("leaderontop", &n.LeaderOnTop)
	k.ints(pre+"teammate.pos", &n.Teammate.Pos[0], &n.Teammate.Pos[1])
	k.ints(pre+"teammate.spacing", &n.Teammate.Spacing[0], &n.Teammate.Spacing[1])
	k.flag(pre+"teammate.ko.hide", &n.Teammate.KoHide)
	return n
}

// winIcon is readFightScreenWinIcon (fightscreen.go:2238).
func (r reader) winIcon(k iniKeys, pre string) WinIcon {
	w := WinIcon{
		UseIconUpTo: 4,
		Counter:     text(k, pre+"counter.", "%i", 0, 0),
		Bg0:         r.element(k, pre+"bg0.", 0),
		Top:         r.element(k, pre+"top.", 0),
		Icons: WinIcons{
			Normal:   r.element(k, pre+"n.", 0),
			Special:  r.element(k, pre+"s.", 0),
			Hyper:    r.element(k, pre+"h.", 0),
			Cheese:   r.element(k, pre+"c.", 0),
			Time:     r.element(k, pre+"t.", 0),
			Throw:    r.element(k, pre+"throw.", 0),
			Suicide:  r.element(k, pre+"suicide.", 0),
			Teammate: r.element(k, pre+"teammate.", 0),
			Perfect:  r.element(k, pre+"perfect.", 0),
			Clutch:   r.element(k, pre+"clutch.", 0),
		},
	}
	k.ints(pre+"pos", &w.Pos[0], &w.Pos[1])
	k.ints(pre+"iconoffset", &w.IconOffset[0], &w.IconOffset[1])
	k.ints("useiconupto", &w.UseIconUpTo)
	return w
}

// time is readFightScreenTime (fightscreen.go:2405).
func (r reader) time(k iniKeys) Time {
	t := Time{
		Counter:        text(k, "counter.", "", 0, 0),
		Bg:             r.element(k, "bg.", 0),
		Top:            r.element(k, "top.", 0),
		FramesPerCount: 60,
	}
	k.ints("pos", &t.Pos[0], &t.Pos[1])
	k.ints("framespercount", &t.FramesPerCount)
	return t
}

// combo is readFightScreenCombo (fightscreen.go:2514) for team 1 or 2. Without
// an Ikemen GO version, MUGEN syntax applies: team N reads the teamN. keys when
// teamN.pos is written and the unprefixed keys otherwise; Ikemen GO reads
// teamN. keys with selectors (fightscreen.go:5332-5353).
func (r reader) combo(s iniSection, team int) Combo {
	prefix := fmt.Sprintf("team%d.", team)
	k := s.keys
	if !r.legacyCombo {
		k = s.forPlayer("team", team)
	} else if _, ok := k[prefix+"pos"]; !ok {
		prefix = ""
	}
	align := 0
	if prefix == "" {
		align = 1
		if team == 2 {
			align = -1
		}
	}
	c := Combo{
		Counter:     text(k, prefix+"counter.", "%i", 2, align),
		Text:        text(k, prefix+"text.", "", 2, align),
		Bg:          r.element(k, prefix+"bg0.", 2),
		Top:         r.element(k, prefix+"top.", 2),
		DisplayTime: 90,
		ShowSpeed:   8,
		HideSpeed:   4,
		AutoAlign:   true,
	}
	k.ints(prefix+"pos", &c.Pos[0], &c.Pos[1])
	k.floats(prefix+"start.x", &c.StartX)
	k.ints(prefix+"displaytime", &c.DisplayTime)
	k.floats(prefix+"showspeed", &c.ShowSpeed)
	c.ShowSpeed = max(1, c.ShowSpeed)
	k.floats(prefix+"hidespeed", &c.HideSpeed)
	k.flag(prefix+"autoalign", &c.AutoAlign)
	if team == 2 && prefix == "" {
		// MUGEN 1.0 shares the keys between teams; team 2 is the mirror.
		c.Pos[0] = r.localcoordWidth - c.Pos[0]
	}
	if team == 2 && prefix != "" {
		c.StartX = r.mirrorStartX(c.Pos[0], c.StartX)
	}
	return c
}

// action is readFightScreenAction (fightscreen.go:3088) for team 1 or 2.
func (r reader) action(k iniKeys, team int) Action {
	pre := fmt.Sprintf("team%d.", team)
	a := Action{
		Text:        text(k, pre+"text.", "", 2, 0),
		DisplayTime: 90,
		ShowSpeed:   8,
		HideSpeed:   4,
		Max:         8,
	}
	k.ints(pre+"pos", &a.Pos[0], &a.Pos[1])
	k.ints(pre+"spacing", &a.Spacing[0], &a.Spacing[1])
	k.floats(pre+"start.x", &a.StartX)
	k.ints(pre+"displaytime", &a.DisplayTime)
	k.floats(pre+"showspeed", &a.ShowSpeed)
	a.ShowSpeed = max(1, a.ShowSpeed)
	k.floats(pre+"hidespeed", &a.HideSpeed)
	k.ints(pre+"max", &a.Max)
	if team == 2 {
		a.StartX = r.mirrorStartX(a.Pos[0], a.StartX)
	}
	return a
}

// mirrorStartX is how Ikemen GO turns the start.x of the second team into the
// side it slides in from. A mirror at or right of the left edge keeps the
// movement enabled by sliding from the distance to the team's position.
func (r reader) mirrorStartX(posX int, startX float32) float32 {
	mirrored := float32(r.localcoordWidth) - startX
	if mirrored >= 0 {
		return min(float32(posX)-startX, 0)
	}
	return mirrored
}

// announcementPrefix is the prefix a side's announcement is read from, as Ikemen
// GO's readATS(pPre, pre) chooses it (fightscreen.go:3457-3468): the
// player's own keys when its text, spr or anim is written, otherwise the
// shared ones. The two sets are never merged.
func announcementPrefix(k iniKeys, p, pre string) string {
	for _, key := range []string{"text", "spr", "anim"} {
		if _, ok := k[p+pre+key]; ok {
			return p + pre
		}
	}
	return pre
}

// announcement is Ikemen GO's ReadAnimTextSnd (common.go:1648).
func (r reader) announcement(k iniKeys, pre string, layerNo int) Announcement {
	a := Announcement{
		Sound:       SoundRef{Group: -1},
		Text:        text(k, pre, "", layerNo, 0),
		Element:     r.element(k, pre, layerNo),
		DisplayTime: -2,
	}
	k.ints(pre+"snd", &a.Sound.Group, &a.Sound.Number)
	k.ints(pre+"displaytime", &a.DisplayTime)
	return a
}

// round is readFightScreenRound (fightscreen.go:3342).
func (r reader) round(k iniKeys) Round {
	ro := Round{
		StartWaitTime: 30, CtrlTime: 30, SlowTime: 60, SlowSpeed: 0.25,
		OverHitTime: 10, OverWaitTime: 45, OverWinTime: 45, OverForceWinTime: 900,
		OverTime: 210, ShutterTime: 15, ClutchThreshold: 10, CallFightTime: 60,
	}
	k.ints("pos", &ro.Pos[0], &ro.Pos[1])
	k.ints("start.waittime", &ro.StartWaitTime)
	ro.StartWaitTime = max(1, ro.StartWaitTime)
	k.ints("round.time", &ro.RoundTime)
	ro.RoundSndTime = ro.RoundTime
	k.ints("round.sndtime", &ro.RoundSndTime)
	for i := range ro.Round {
		ro.Round[i] = r.announcement(k, fmt.Sprintf("round%d.", i+1), 2)
	}
	ro.RoundDefault = r.announcement(k, "round.default.", 2)
	ro.RoundSingle = r.announcement(k, "round.single.", 2)
	ro.RoundFinal = r.announcement(k, "round.final.", 2)

	k.ints("fight.time", &ro.FightTime)
	ro.FightSndTime = ro.FightTime
	k.ints("fight.sndtime", &ro.FightSndTime)
	ro.Fight = r.announcement(k, "fight.", 2)
	k.ints("ctrl.time", &ro.CtrlTime)
	ro.CtrlTime = max(1, ro.CtrlTime)

	k.ints("ko.time", &ro.KoTime)
	ro.KoSndTime = ro.KoTime
	k.ints("ko.sndtime", &ro.KoSndTime)
	ro.Ko = r.announcement(k, "ko.", 1)
	ro.DkoTime, ro.DkoSndTime = ro.KoTime, ro.KoSndTime
	k.ints("dko.time", &ro.DkoTime)
	k.ints("dko.sndtime", &ro.DkoSndTime)
	k.flag("dko.showdraw", &ro.DkoShowDraw)
	ro.Dko = r.announcement(k, "dko.", 1)
	ro.ToTime, ro.ToSndTime = ro.KoTime, ro.KoSndTime
	k.ints("to.time", &ro.ToTime)
	k.ints("to.sndtime", &ro.ToSndTime)
	ro.To = r.announcement(k, "to.", 1)
	ro.Draw = r.announcement(k, "draw.", 1)
	for i := range ro.Win {
		pre := announcementPrefix(k, playerPrefix(i), "win.")
		ro.Win[i] = r.announcement(k, pre, 1)
	}

	k.ints("slow.time", &ro.SlowTime)
	ro.SlowFadeTime = int(float32(ro.SlowTime) * 0.75)
	k.ints("slow.fadetime", &ro.SlowFadeTime)
	ro.SlowFadeTime = min(ro.SlowFadeTime, ro.SlowTime)
	k.floats("slow.speed", &ro.SlowSpeed)
	ro.SlowSpeed = min(max(ro.SlowSpeed, 0.01), 1)
	k.ints("over.hittime", &ro.OverHitTime)
	ro.OverHitTime = max(1, ro.OverHitTime)
	k.ints("over.waittime", &ro.OverWaitTime)
	ro.OverWaitTime = max(1, ro.OverWaitTime)
	k.ints("over.wintime", &ro.OverWinTime)
	ro.OverWinTime = max(1, ro.OverWinTime)
	k.ints("over.forcewintime", &ro.OverForceWinTime)
	ro.OverForceWinTime = max(1, ro.OverForceWinTime)
	k.ints("over.time", &ro.OverTime)
	ro.OverTime = max(1, ro.OverTime)
	k.ints("win.time", &ro.WinTime)
	ro.WinSndTime = ro.WinTime
	k.ints("win.sndtime", &ro.WinSndTime)
	k.ints("shutter.time", &ro.ShutterTime)
	k.ints("clutch.threshold", &ro.ClutchThreshold)
	k.ints("callfight.time", &ro.CallFightTime)
	return ro
}
