package lifebar

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func fixture(t testing.TB, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestParse_ReadsTheIkemenGODefaultFightDef(t *testing.T) {
	lb, warnings := Parse(fixture(t, "ikemen-default-fight.def"))

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if lb.Localcoord != [2]int{1280, 720} {
		t.Errorf("Localcoord = %v, want 1280x720", lb.Localcoord)
	}
	if lb.Files.Sprites != "fight.sff" || lb.Files.Sounds != "fight.snd" {
		t.Errorf("Files = %+v, want fight.sff and fight.snd", lb.Files)
	}
	if lb.Files.FightFxSprites != "fightfx.sff" || lb.Files.FightFxSounds != "common.snd" {
		t.Errorf("FightFx files = %q %q, want fightfx.sff and common.snd (comments cut)",
			lb.Files.FightFxSprites, lb.Files.FightFxSounds)
	}
	if lb.Files.FightFx[0] != "" {
		t.Errorf("FightFx[0] = %q, want empty (fx1 is commented out)", lb.Files.FightFx[0])
	}
	if got := lb.Files.Fonts[1]; got != (Font{File: "ikemen1/fonts/Menu2.def"}) {
		t.Errorf("Fonts[1] = %+v", got)
	}
	if got := lb.Files.Fonts[9]; got.File != "ikemen1/fonts/Round.def" {
		t.Errorf("Fonts[9] = %+v", got)
	}

	p1, p2 := lb.Life[0], lb.Life[1]
	if p1.Pos != [2]int{595, 40} || p2.Pos != [2]int{684, 40} {
		t.Errorf("life pos = %v %v, want 595,40 and 684,40", p1.Pos, p2.Pos)
	}
	if p1.RangeX != [2]int{15, -460} || p2.RangeX != [2]int{-15, 460} {
		t.Errorf("life range.x = %v %v", p1.RangeX, p2.RangeX)
	}
	if p1.Front.Sprite != (SpriteRef{13, 3}) || !p1.Front.HasSprite || p1.Front.Layout.Facing != 1 {
		t.Errorf("p1 front = %+v, want sprite 13,3 facing 1", p1.Front)
	}
	if p2.Front.Sprite != (SpriteRef{13, 3}) || p2.Front.Layout.Facing != -1 {
		t.Errorf("p2 front = %+v, want sprite 13,3 facing -1", p2.Front)
	}
	if p1.Mid.Sprite != (SpriteRef{12, 0}) || p1.Bg0.Sprite != (SpriteRef{10, 0}) || p1.Top.Sprite != (SpriteRef{11, 0}) {
		t.Errorf("p1 mid/bg0/top = %v %v %v", p1.Mid.Sprite, p1.Bg0.Sprite, p1.Top.Sprite)
	}

	w1 := lb.Power[0]
	if w1.Pos != [2]int{568, 83} || w1.RangeX != [2]int{-4, -206} {
		t.Errorf("power p1 pos/range = %v %v", w1.Pos, w1.RangeX)
	}
	if w1.Front.Sprite != (SpriteRef{43, 0}) || w1.Front.Layout.Offset != [2]float32{-4, 4} {
		t.Errorf("power p1 front = %+v", w1.Front)
	}
	if w1.Bg0.Layout.Offset != [2]float32{0, 3} {
		t.Errorf("power p1 bg0 offset = %v, want 0,3", w1.Bg0.Layout.Offset)
	}

	if f := lb.Face[0]; f.Pos != [2]int{20, 20} || f.FaceSprite != (SpriteRef{9000, 0}) || f.Bg1.Sprite != (SpriteRef{51, 0}) {
		t.Errorf("face p1 = %+v", f)
	}
	if f := lb.Face[1]; f.FaceLayout.Facing != -1 || f.Bg0.Layout.Facing != -1 {
		t.Errorf("face p2 = %+v, want facing -1", f)
	}

	if n := lb.Name[0]; n.Pos != [2]int{13, 145} || n.Name.Font != (TextFont{No: 3, Bank: 0, Align: 1, Color: [3]int{-1, -1, -1}, Alpha: 255}) {
		t.Errorf("name p1 = %+v", n)
	}
	if n := lb.Name[1]; n.Pos != [2]int{1267, 145} || n.Name.Font.Align != -1 {
		t.Errorf("name p2 = %+v", n)
	}

	ti := lb.Time
	if ti.Pos != [2]int{640, 105} || ti.Counter.Font.No != 2 || ti.FramesPerCount != 60 {
		t.Errorf("time = %+v", ti)
	}
	if ti.Counter.Layout.Scale != [2]float32{2, 2} || ti.Counter.Layout.Offset != [2]float32{0, 30} {
		t.Errorf("time counter layout = %+v", ti.Counter.Layout)
	}

	wi := lb.WinIcon[0]
	if wi.Pos != [2]int{544, 120} || wi.IconOffset != [2]int{-32, 0} || wi.Bg0.Sprite != (SpriteRef{120, 0}) {
		t.Errorf("winicon p1 = %+v", wi)
	}
	if wi.Counter.Font != (TextFont{No: 4, Color: [3]int{230, 255, 242}, Alpha: 255}) {
		t.Errorf("winicon counter font = %+v", wi.Counter.Font)
	}

	c0, c1 := lb.Combo[0], lb.Combo[1]
	if c0.Pos != [2]int{130, 260} || c0.StartX != -240 || c0.Counter.Font.No != 6 || c0.Counter.Layout.LayerNo != 2 {
		t.Errorf("combo team1 = %+v", c0)
	}
	if c1.Pos != [2]int{1150, 260} || c1.StartX != -240 {
		t.Errorf("combo team2 pos/start.x = %v %v, want 1150,260 and 1280-1520", c1.Pos, c1.StartX)
	}

	a0, a1 := lb.Action[0], lb.Action[1]
	if a0.Pos != [2]int{30, 377} || a0.Spacing != [2]int{0, 32} || a0.Max != 8 || a0.StartX != -240 {
		t.Errorf("action team1 = %+v", a0)
	}
	if a0.DisplayTime != 60 || a0.ShowSpeed != 2 || a0.HideSpeed != 8 {
		t.Errorf("action team1 timing = %+v", a0)
	}
	if a1.Pos != [2]int{1250, 377} || a1.StartX != -240 {
		t.Errorf("action team2 pos/start.x = %v %v, want 1250,377 and 1280-1520", a1.Pos, a1.StartX)
	}

	r := lb.Round
	if r.StartWaitTime != 30 || r.RoundTime != 0 || r.RoundSndTime != 0 {
		t.Errorf("round times = %d %d %d", r.StartWaitTime, r.RoundTime, r.RoundSndTime)
	}
	if r.RoundDefault.Text.Text != "R   %i" || r.RoundDefault.Text.Font.No != 9 ||
		r.RoundDefault.DisplayTime != 55 || r.RoundDefault.Sound != (SoundRef{0, 11}) {
		t.Errorf("round.default = %+v", r.RoundDefault)
	}
	if r.Round[0].Sound != (SoundRef{0, 1}) || r.Round[8].Sound != (SoundRef{0, 9}) ||
		r.Round[0].Text.Layout.Offset != [2]float32{640, 380} {
		t.Errorf("round1/round9 = %+v %+v", r.Round[0], r.Round[8])
	}
	if r.RoundFinal.Text.Text != "F   R" {
		t.Errorf("round.final.text = %q, want the first of its two keys", r.RoundFinal.Text.Text)
	}
	if !r.Fight.Element.HasAnim || r.Fight.Element.Anim != 520 || r.Fight.Sound != (SoundRef{1, 0}) {
		t.Errorf("fight = %+v", r.Fight)
	}

	a, ok := lb.Animations[400]
	if !ok || len(a.Frames) != 1 || a.Frames[0].Group != 400 || a.Frames[0].Time != -1 {
		t.Errorf("Animations[400] = %+v", a)
	}
	a = lb.Animations[500]
	if len(a.Frames) != 4 || a.Frames[2].X != -150 || a.Frames[2].Time != 10 {
		t.Errorf("Animations[500] = %+v, want 4 frames with frame 2 at x=-150 for 10 ticks", a)
	}
}

func TestParse_ReadsTheVHDBars(t *testing.T) {
	lb, warnings := Parse(fixture(t, "vhd-ikemen-go-fixture.def"))

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if lb.Localcoord != [2]int{1280, 720} {
		t.Errorf("Localcoord = %v", lb.Localcoord)
	}
	if lb.Files.Sprites != "fightHD.sff" || lb.Files.Fonts[5].File != "font/VPHD_Time.def" ||
		lb.Files.FightFxAnimations != "fightfx.air" {
		t.Errorf("Files = %+v", lb.Files)
	}
	bg0 := lb.Life[0].Bg0
	if !bg0.HasAnim || bg0.Anim != 170 || bg0.Layout.Offset != [2]float32{640, 78} || bg0.Layout.LayerNo != 0 {
		t.Errorf("life p1 bg0 = %+v, want anim 170 at 640,78 (the [Simul Lifebar] one is 75)", bg0)
	}
	if lb.Power[0].Bg1.Layout.Scale != [2]float32{0.75, 0.75} {
		t.Errorf("power p1 bg1 scale = %v, want .75,.75", lb.Power[0].Bg1.Layout.Scale)
	}
	if lb.Face[0].FaceLayout.Scale != [2]float32{0.75, 0.75} || lb.Face[0].FaceLayout.Offset != [2]float32{40, 17} {
		t.Errorf("face p1 layout = %+v", lb.Face[0].FaceLayout)
	}
	if lb.Name[0].Pos != [2]int{139, 64} || lb.Name[1].Name.Font.No != 6 || lb.Name[1].Name.Font.Align != -1 {
		t.Errorf("names = %+v %+v", lb.Name[0], lb.Name[1])
	}
	if lb.Time.Pos != [2]int{640, 107} || lb.Time.Counter.Font.No != 5 {
		t.Errorf("time = %+v (the ;640,132 tail is a comment)", lb.Time)
	}
	if lb.WinIcon[0].Pos != [2]int{530, 124} || lb.WinIcon[0].IconOffset != [2]int{-33, 0} ||
		lb.WinIcon[1].Counter.Font.Bank != -1 {
		t.Errorf("winicon = %+v %+v", lb.WinIcon[0], lb.WinIcon[1])
	}
	r := lb.Round
	if r.RoundTime != 20 || !r.RoundDefault.Element.HasAnim || r.RoundDefault.Element.Anim != 300 ||
		r.RoundDefault.DisplayTime != 170 || r.RoundDefault.Text.Layout.Offset != [2]float32{640, 410} {
		t.Errorf("round = %+v", r.RoundDefault)
	}

	if len(lb.Animations) != 3 {
		t.Errorf("Animations = %d, want 170, 220 and 300", len(lb.Animations))
	}
	a := lb.Animations[170]
	if len(a.Frames) != 2 || a.Frames[0].Group != 11 || a.Frames[0].Image != 3 ||
		a.Frames[0].Time != 1080 || a.Frames[1].Time != 1 {
		t.Errorf("Animations[170] = %+v", a)
	}
}

func TestParse_ReadsTheVHDCombosThroughMugenSyntaxWhenNoIkemenVersionIsDeclared(t *testing.T) {
	lb, _ := Parse(fixture(t, "vhd-ikemen-go-fixture.def"))

	c0 := lb.Combo[0]
	if c0.Pos != [2]int{0, 250} || c0.StartX != -150 || c0.DisplayTime != 100 || c0.Text.Text != "G%i" {
		t.Errorf("combo team1 = %+v", c0)
	}
	// No team2.pos: team 2 reads the unprefixed keys and mirrors its x across localcoord.
	c1 := lb.Combo[1]
	if c1.Pos != [2]int{1280, 0} || c1.Counter.Font.No != 2 || c1.Counter.Font.Align != 1 {
		t.Errorf("combo team2 = %+v, want pos 1280,0 and the unprefixed counter.font 2,0,1", c1)
	}
	if c0.Counter.Font.No != -1 {
		t.Errorf("combo team1 counter font = %d, want -1: its counter.font key has no team1. prefix", c0.Counter.Font.No)
	}
}

func TestParse_ReadsTheMFJ2ClassicMugenLifebar(t *testing.T) {
	lb, warnings := Parse(fixture(t, "mfj2-classic-mugen-fixture.def"))

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if lb.Localcoord != [2]int{640, 480} {
		t.Errorf("Localcoord = %v, want 640x480 (the commented 1280,720 is ignored)", lb.Localcoord)
	}
	if lb.Files.Sprites != "sff/fight.sff" || lb.Files.Fonts[8].File != "font/18x18-2.fnt" ||
		lb.Files.FightFxSounds != "snd/common.snd" {
		t.Errorf("Files = %+v", lb.Files)
	}
	p1, p2 := lb.Life[0], lb.Life[1]
	if p1.Pos != [2]int{0, 20} || p2.Pos != [2]int{320, 20} {
		t.Errorf("life pos = %v %v", p1.Pos, p2.Pos)
	}
	if p1.Bg0.Sprite != (SpriteRef{10, 0}) || p2.Bg0.Sprite != (SpriteRef{10, 1}) {
		t.Errorf("life bg0 = %v %v", p1.Bg0.Sprite, p2.Bg0.Sprite)
	}
	if !p1.Bg1.HasAnim || p1.Bg1.Anim != 111 || p1.Bg1.Layout.Offset != [2]float32{2, 6} {
		t.Errorf("life p1 bg1 = %+v", p1.Bg1)
	}
	if p1.Bg2.Layout.LayerNo != 1 || p1.Front.Sprite != (SpriteRef{14, 0}) || p1.RangeX != [2]int{283, 66} {
		t.Errorf("life p1 bg2/front/range = %+v", p1)
	}
	c0, c1 := lb.Combo[0], lb.Combo[1]
	if c0.Pos != [2]int{60, 172} || c0.StartX != -80 || c0.Text.Font.No != 6 || c0.Text.Text != "%iH" {
		t.Errorf("combo team1 = %+v", c0)
	}
	if c1.Pos != [2]int{580, 172} || c1.StartX != -80 {
		t.Errorf("combo team2 pos/start.x = %v %v, want 580,172 and 640-720", c1.Pos, c1.StartX)
	}
	if wi := lb.WinIcon[0]; wi.Pos != [2]int{282, 80} || wi.IconOffset != [2]int{-10, 0} ||
		wi.Counter.Font.No != 7 || wi.Icons.Normal.Sprite != (SpriteRef{80, 0}) || wi.Icons.Normal.Layout.LayerNo != 2 {
		t.Errorf("winicon p1 = %+v", wi)
	}
	if len(lb.Animations) != 6 {
		t.Errorf("Animations = %d, want 6", len(lb.Animations))
	}
	if f := lb.Animations[132].Frames; len(f) != 1 || f[0].Group != 13 || f[0].Image != 1 || f[0].Blend != "A1" {
		t.Errorf("Animations[132] frames = %+v", f)
	}
}

func TestParse_ReadsTheGameDefaultLifebar(t *testing.T) {
	lb, warnings := Parse(fixture(t, "game-default-fight.def"))

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if lb.Localcoord != [2]int{320, 240} || lb.Files.Sprites != "fight.sff" {
		t.Errorf("Localcoord/Files = %v %+v", lb.Localcoord, lb.Files)
	}
	if lb.Life[0].Pos != [2]int{8, 8} || lb.Life[1].RangeX != [2]int{2, 129} {
		t.Errorf("life = %+v %+v", lb.Life[0], lb.Life[1])
	}
	if m := lb.Life[0].Mid; m.Sprite != (SpriteRef{3, 0}) || m.Layout.Offset != [2]float32{2, 2} {
		t.Errorf("life p1 mid = %+v", m)
	}
	if lb.Power[1].Pos != [2]int{208, 22} || lb.Power[1].RangeX != [2]int{101, 2} {
		t.Errorf("power p2 = %+v", lb.Power[1])
	}
	if lb.Name[1].Pos != [2]int{312, 42} || lb.Name[1].Name.Font.Align != -1 {
		t.Errorf("name p2 = %+v", lb.Name[1])
	}
	if lb.Time.Pos != [2]int{160, 22} || lb.Time.Counter.Font.No != 1 {
		t.Errorf("time = %+v", lb.Time)
	}
}

func TestParse_ReadsTheGMSFixtureWithoutWarningOnIkemenOnlySections(t *testing.T) {
	lb, warnings := Parse(fixture(t, "gms-ikemen-go-fixture.def"))

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none for [FightFx], [Simul_3P Lifebar], [Guardbar], [Stunbar]", warnings)
	}
	if lb.Localcoord != [2]int{320, 240} || lb.Files.FightFx[0] != "gofx.def" || lb.Files.Fonts[3].File != "font/round.fnt" {
		t.Errorf("Localcoord/Files = %v %+v", lb.Localcoord, lb.Files)
	}
	if lb.Life[0].Pos != [2]int{142, 12} || lb.Life[0].Bg0.Layout.Scale != [2]float32{0.5, 0.5} {
		t.Errorf("life p1 = %+v", lb.Life[0])
	}
	if c := lb.Combo[0]; c.Pos != [2]int{10, 98} || c.StartX != -40 || c.Text.Text != "H%iC" ||
		c.Text.Font != (TextFont{No: 4, Bank: 0, Align: 1, Color: [3]int{-1, -1, -1}, Alpha: 255}) {
		t.Errorf("combo team1 = %+v", c)
	}
}

func TestParse_DropsKeysAfterATruncatedHeader(t *testing.T) {
	lb, warnings := Parse(fixture(t, "truncated-real-fixture.def"))

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none: a line that is not a header is not a section", warnings)
	}
	if lb.Localcoord != [2]int{320, 240} || lb.Files.Sprites != "fight.sff" || lb.Files.FightFxSounds != "common.snd" {
		t.Errorf("Localcoord/Files = %v %+v", lb.Localcoord, lb.Files)
	}
	if lb.Combo[0].Pos != [2]int{} {
		t.Errorf("combo team1 pos = %v, want 0,0: [Combo has no closing bracket", lb.Combo[0].Pos)
	}
}

func TestParse_GivesIkemenDefaultsOnEmptyInput(t *testing.T) {
	lb, warnings := Parse("")

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if lb.Localcoord != [2]int{320, 240} {
		t.Errorf("Localcoord = %v, want 320x240", lb.Localcoord)
	}
	if lb.Animations == nil || len(lb.Animations) != 0 {
		t.Errorf("Animations = %v, want empty and non-nil", lb.Animations)
	}
	front := lb.Life[0].Front
	if front.HasSprite || front.HasAnim || front.Layout.Facing != 1 || front.Layout.VFacing != 1 ||
		front.Layout.Scale != [2]float32{1, 1} || front.Layout.LayerNo != 0 {
		t.Errorf("life front = %+v, want an empty element with facing 1 and scale 1", front)
	}
	if l := lb.Life[1]; !l.MidFreeze || l.MidDelay != 30 || l.MidMult != 1 || l.MidSteps != 8 {
		t.Errorf("life trail = %+v, want freeze, delay 30, mult 1, steps 8", l)
	}
	if f := lb.Face[0]; f.FaceSprite != (SpriteRef{Group: -1}) || !f.FacePalShare || f.FacePalFxShare ||
		f.Teammate.FaceSprite != (SpriteRef{Group: -1}) || !f.Teammate.FacePalShare {
		t.Errorf("face = %+v", f)
	}
	if lb.WinIcon[0].UseIconUpTo != 4 {
		t.Errorf("UseIconUpTo = %d, want 4", lb.WinIcon[0].UseIconUpTo)
	}
	if lb.Time.FramesPerCount != 60 {
		t.Errorf("FramesPerCount = %d, want 60", lb.Time.FramesPerCount)
	}
	if lb.Time.Counter.Font != (TextFont{No: -1, Color: [3]int{-1, -1, -1}, Alpha: 255}) {
		t.Errorf("time counter font = %+v", lb.Time.Counter.Font)
	}
	if c := lb.Combo[1]; c.DisplayTime != 90 || c.ShowSpeed != 8 || c.HideSpeed != 4 || !c.AutoAlign ||
		c.Counter.Text != "%i" || c.Counter.Layout.LayerNo != 2 || c.Counter.Font.Align != -1 {
		t.Errorf("combo team2 = %+v", c)
	}
	if a := lb.Action[0]; a.DisplayTime != 90 || a.ShowSpeed != 8 || a.HideSpeed != 4 || a.Max != 8 {
		t.Errorf("action = %+v", a)
	}
	r := lb.Round
	if r.StartWaitTime != 30 || r.CtrlTime != 30 || r.SlowTime != 60 || r.SlowFadeTime != 45 || r.SlowSpeed != 0.25 {
		t.Errorf("round times = %+v", r)
	}
	if r.OverHitTime != 10 || r.OverWaitTime != 45 || r.OverWinTime != 45 || r.OverForceWinTime != 900 ||
		r.OverTime != 210 || r.ShutterTime != 15 || r.ClutchThreshold != 10 || r.CallFightTime != 60 {
		t.Errorf("round over times = %+v", r)
	}
	if r.Ko.Sound != (SoundRef{Group: -1}) || r.Ko.DisplayTime != -2 || r.Ko.Element.Layout.LayerNo != 1 ||
		r.Round[0].Element.Layout.LayerNo != 2 {
		t.Errorf("round announcements: ko = %+v, round1 = %+v", r.Ko, r.Round[0])
	}
}

func TestParse_StripsAByteOrderMarkAndReadsCRLFAndFFFDComments(t *testing.T) {
	lb, warnings := Parse("\xef\xbb\xbf[Info]\r\nlocalcoord = 640,480 ; ��\r\n\r\n[Time]\r\npos = 3,4\r\n")

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if lb.Localcoord != [2]int{640, 480} || lb.Time.Pos != [2]int{3, 4} {
		t.Errorf("Localcoord = %v, Time.Pos = %v", lb.Localcoord, lb.Time.Pos)
	}
}

func TestParse_ReadsAnIndentedHeaderAsAHeaderBecauseIkemenTrimsEveryLine(t *testing.T) {
	lb, _ := Parse("[Time]\npos = 1,2\n  [Lifebar]\n  p1.pos = 3,4\n")

	if lb.Time.Pos != [2]int{1, 2} || lb.Life[0].Pos != [2]int{3, 4} {
		t.Errorf("Time.Pos = %v, Life[0].Pos = %v, want 1,2 and 3,4", lb.Time.Pos, lb.Life[0].Pos)
	}
}

func TestParse_EndsASectionAtAHeaderWithoutClosingBracket(t *testing.T) {
	lb, warnings := Parse("[Time]\npos = 1,2\n[Lifebar\npos = 9,9\np1.pos = 3,4\n")

	if lb.Time.Pos != [2]int{1, 2} {
		t.Errorf("Time.Pos = %v, want 1,2", lb.Time.Pos)
	}
	if lb.Life[0].Pos != [2]int{} {
		t.Errorf("Life[0].Pos = %v, want 0,0: [Lifebar has no closing bracket", lb.Life[0].Pos)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
}

func TestParse_ReadsAKeyWithoutEqualsAsEmptyAndHidesALaterOne(t *testing.T) {
	withBare, _ := Parse("[Time]\nframespercount bar\nframespercount = 30\n")
	plain, _ := Parse("[Time]\nframespercount = 30\n")

	if withBare.Time.FramesPerCount != 60 {
		t.Errorf("FramesPerCount = %d, want the default 60: the bare key stored an empty value first", withBare.Time.FramesPerCount)
	}
	if plain.Time.FramesPerCount != 30 {
		t.Errorf("FramesPerCount = %d, want 30", plain.Time.FramesPerCount)
	}
}

func TestParse_FirstSectionWinsWhenASectionRepeats(t *testing.T) {
	lb, _ := Parse("[Lifebar]\np1.pos = 1,1\n[Lifebar]\np1.pos = 2,2\np2.pos = 3,3\n[Info]\nlocalcoord = 100,200\n[Info]\nlocalcoord = 300,400\n")

	if lb.Life[0].Pos != [2]int{1, 1} {
		t.Errorf("Life[0].Pos = %v, want 1,1", lb.Life[0].Pos)
	}
	if lb.Life[1].Pos != [2]int{} {
		t.Errorf("Life[1].Pos = %v, want 0,0: the second [Lifebar] is ignored whole", lb.Life[1].Pos)
	}
	if lb.Localcoord != [2]int{100, 200} {
		t.Errorf("Localcoord = %v, want 100,200", lb.Localcoord)
	}
}

func TestParse_UsesOnlyTheFirstFilesSection(t *testing.T) {
	lb, _ := Parse("[Files]\nsff = a.sff\n[Files]\nsff = b.sff\nsnd = b.snd\n")

	if lb.Files.Sprites != "a.sff" || lb.Files.Sounds != "" {
		t.Errorf("Files = %+v, want only the first section", lb.Files)
	}
}

func TestParse_ReadsFontsWithTheirHeightsAndSkipsFontsWithoutAFile(t *testing.T) {
	lb, _ := Parse("[Files]\nfont2 = a.def\nfont2.height = 16\nfont3.height = 12\nfont4 = \nfont x = z\n")

	if got := lb.Files.Fonts[2]; got != (Font{File: "a.def", Height: 16}) {
		t.Errorf("Fonts[2] = %+v", got)
	}
	if len(lb.Files.Fonts) != 1 {
		t.Errorf("Fonts = %+v, want only font 2", lb.Files.Fonts)
	}
}

func TestParse_LocalcoordWithOneValueKeepsTheDefaultHeight(t *testing.T) {
	lb, _ := Parse("[Info]\nlocalcoord = 640\n")

	if lb.Localcoord != [2]int{640, 240} {
		t.Errorf("Localcoord = %v, want 640,240", lb.Localcoord)
	}
}

func TestParse_ResolvesPlayerSelectors(t *testing.T) {
	lb, _ := Parse("[Lifebar]\n" +
		"p*.front.spr = 1,1\n" +
		"p1.front.spr = 2,2\n" +
		"p1|p3.pos = 5,5\n" +
		"p2.pos = 7,7\n" +
		"p*.range.x = 1,2\n" +
		"p1|p2.range.x = 3,4\n")

	if lb.Life[0].Front.Sprite != (SpriteRef{2, 2}) || lb.Life[1].Front.Sprite != (SpriteRef{1, 1}) {
		t.Errorf("front sprites = %v %v, want 2,2 (explicit) and 1,1 (p*)",
			lb.Life[0].Front.Sprite, lb.Life[1].Front.Sprite)
	}
	if lb.Life[0].Pos != [2]int{5, 5} || lb.Life[1].Pos != [2]int{7, 7} {
		t.Errorf("pos = %v %v, want 5,5 (p1|p3) and 7,7 (explicit p2)", lb.Life[0].Pos, lb.Life[1].Pos)
	}
	if lb.Life[0].RangeX != [2]int{1, 2} || lb.Life[1].RangeX != [2]int{1, 2} {
		t.Errorf("range.x = %v %v, want 1,2 for both: the first selector in the file wins", lb.Life[0].RangeX, lb.Life[1].RangeX)
	}
}

func TestParse_ResolvesTeamSelectorsInComboWhenAnIkemenVersionIsDeclared(t *testing.T) {
	lb, _ := Parse("[Info]\nikemenversion = 1.0\n[Combo]\nteam*.pos = 10,20\nteam2.displaytime = 33\n")

	if lb.Combo[0].Pos != [2]int{10, 20} || lb.Combo[1].Pos != [2]int{10, 20} {
		t.Errorf("combo pos = %v %v, want 10,20 for both (no mirroring with prefixes)", lb.Combo[0].Pos, lb.Combo[1].Pos)
	}
	if lb.Combo[0].DisplayTime != 90 || lb.Combo[1].DisplayTime != 33 {
		t.Errorf("combo displaytime = %d %d, want 90 and 33", lb.Combo[0].DisplayTime, lb.Combo[1].DisplayTime)
	}
}

func TestParse_ReadsUnprefixedComboKeysAndMirrorsTeam2WithoutAnIkemenVersion(t *testing.T) {
	lb, _ := Parse("[Info]\nlocalcoord = 320,240\n[Combo]\nteam*.pos = 10,20\npos = 100,5\nstart.x = -30\n")

	if lb.Combo[0].Pos != [2]int{100, 5} || lb.Combo[0].StartX != -30 {
		t.Errorf("combo team1 = %+v, want pos 100,5 start.x -30 (team*.pos is not a selector here)", lb.Combo[0])
	}
	if lb.Combo[1].Pos != [2]int{220, 5} || lb.Combo[1].StartX != -30 {
		t.Errorf("combo team2 = %+v, want pos 320-100,5 and the raw start.x", lb.Combo[1])
	}
}

func TestParse_KeepsTheSecondTeamStartXMovingWhenItsMirrorIsPositive(t *testing.T) {
	lb, _ := Parse("[Info]\nlocalcoord = 320,240\n[Action]\nteam2.pos = 250,10\nteam2.start.x = 200\n")

	// 320-200 = 120 >= 0, so Ikemen GO uses min(pos.x - start.x, 0) = min(50, 0) = 0.
	if got := lb.Action[1].StartX; got != 0 {
		t.Errorf("StartX = %v, want 0", got)
	}
	lb, _ = Parse("[Info]\nlocalcoord = 320,240\n[Action]\nteam2.pos = 100,10\nteam2.start.x = 200\n")
	if got := lb.Action[1].StartX; got != -100 {
		t.Errorf("StartX = %v, want -100 (min(100-200, 0))", got)
	}
}

func TestParse_ReadsNumbersAndLayoutAsIkemenDoes(t *testing.T) {
	lb, _ := Parse("[Lifebar]\n" +
		"p1.pos = 12abc, -3x\n" +
		"p1.bg0.layerno = 7\n" +
		"p1.bg0.facing = -3\n" +
		"p2.bg0.facing = 0\n" +
		"p2.bg0.vfacing = -1\n" +
		"p2.bg0.scale = .5, 2.\n")

	if lb.Life[0].Pos != [2]int{12, -3} {
		t.Errorf("Pos = %v, want 12,-3", lb.Life[0].Pos)
	}
	if bg := lb.Life[0].Bg0.Layout; bg.LayerNo != 2 || bg.Facing != -1 {
		t.Errorf("p1 bg0 layout = %+v, want layerno capped at 2 and facing -1", bg)
	}
	if bg := lb.Life[1].Bg0.Layout; bg.Facing != 1 || bg.VFacing != -1 || bg.Scale != [2]float32{0.5, 2} {
		t.Errorf("p2 bg0 layout = %+v, want facing 1, vfacing -1, scale 0.5,2", bg)
	}
}

func TestParse_IgnoresAnAnimKeyWhoseActionDoesNotExist(t *testing.T) {
	lb, _ := Parse("[Lifebar]\np1.front.spr = 4,5\np1.front.anim = 999\np1.mid.anim = 7\n[Begin Action 7]\n7,0, 0,0, 3\n")

	if f := lb.Life[0].Front; !f.HasSprite || f.Sprite != (SpriteRef{4, 5}) || f.HasAnim {
		t.Errorf("front = %+v, want its sprite kept and no anim", f)
	}
	if m := lb.Life[0].Mid; !m.HasAnim || m.Anim != 7 || m.HasSprite {
		t.Errorf("mid = %+v, want anim 7", m)
	}
}

func TestParse_ClampsAndDefaultsRoundTimes(t *testing.T) {
	lb, _ := Parse("[Round]\nround.time = 12\nko.time = 5\nko.sndtime = 6\nto.sndtime = 9\n" +
		"slow.time = 100\nslow.speed = 5\nover.hittime = 0\nstart.waittime = -4\nctrl.time = 0\n")
	r := lb.Round

	if r.RoundSndTime != 12 {
		t.Errorf("RoundSndTime = %d, want round.time", r.RoundSndTime)
	}
	if r.KoTime != 5 || r.KoSndTime != 6 || r.DkoTime != 5 || r.DkoSndTime != 6 || r.ToTime != 5 || r.ToSndTime != 9 {
		t.Errorf("ko/dko/to times = %+v", r)
	}
	if r.SlowFadeTime != 75 || r.SlowSpeed != 1 {
		t.Errorf("slow = %d %v, want fadetime 75 and speed capped at 1", r.SlowFadeTime, r.SlowSpeed)
	}
	if r.OverHitTime != 1 || r.StartWaitTime != 1 || r.CtrlTime != 1 {
		t.Errorf("minimums = %d %d %d, want 1 1 1", r.OverHitTime, r.StartWaitTime, r.CtrlTime)
	}
	lb, _ = Parse("[Round]\nslow.time = 100\nslow.fadetime = 200\n")
	if lb.Round.SlowFadeTime != 100 {
		t.Errorf("SlowFadeTime = %d, want it capped at slow.time", lb.Round.SlowFadeTime)
	}
}

func TestParse_WarnsOnlyOnSectionsIkemenIgnores(t *testing.T) {
	src := "[Info]\n" + // 1
		"\n" + // 2
		"[Foo]\n" + // 3
		"[GuardBar]\n" + // 4
		"[Simul Lifebar]\n" + // 5
		"[Begin Action 5]\n" + // 6
		"5,0, 0,0, 1\n" + // 7
		"[P1 Life Bar]\n" + // 8
		"[Turns Whatever]\n" + // 9
		"[Tag_4P Name]\n" + // 10
		"[Timer]\n" + // 11
		"[Lifebar ]\n" // 12
	_, warnings := Parse(src)

	want := []Warning{{Section: "Foo", Line: 3}, {Section: "P1 Life Bar", Line: 8}, {Section: "Lifebar ", Line: 12}}
	if !reflect.DeepEqual(warnings, want) {
		t.Errorf("warnings = %v, want %v", warnings, want)
	}
}

// Ikemen GO reads a frame only from a line starting with a digit or '-'
// and skips the others (anim.go): the action of a line that is not a frame
// is kept, without warning; empty, it holds the next block's frames.
func TestParse_SkipsALineThatIsNotAFrame(t *testing.T) {
	lb, warnings := Parse("[Begin Action 1]\n1,0, 0,0, 5\n" +
		"[Begin Action 2]\nthis is not a frame\n" +
		"[Begin Action 3]\n3,0, 0,0, 7\n3,1, 0,0, 8\n")

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if a := lb.Animations[1]; len(a.Frames) != 1 || a.Frames[0].Time != 5 {
		t.Errorf("Animations[1] = %+v", a)
	}
	if a := lb.Animations[2]; len(a.Frames) != 2 || a.Frames[0].Group != 3 {
		t.Errorf("Animations[2] = %+v, want the next block's frames", a)
	}
	if a := lb.Animations[3]; len(a.Frames) != 2 {
		t.Errorf("Animations[3] = %+v", a)
	}
}

func TestParse_KeepsTheFirstOfTwoActionsWithTheSameNumber(t *testing.T) {
	lb, _ := Parse("[Begin Action 1]\n1,0, 0,0, 5\n[Begin Action 1]\n1,0, 0,0, 9\n")

	if a := lb.Animations[1]; len(a.Frames) != 1 || a.Frames[0].Time != 5 {
		t.Errorf("Animations[1] = %+v, want the first block", a)
	}
}

func TestParse_ReadsAnimationsWhateverTheCaseOfTheHeader(t *testing.T) {
	lb, _ := Parse("[BEGIN action 4] ; c\n4,0, 0,0, 2\n")

	if a := lb.Animations[4]; len(a.Frames) != 1 {
		t.Errorf("Animations[4] = %+v", a)
	}
}

func FuzzParse(f *testing.F) {
	for _, name := range []string{
		"ikemen-default-fight.def", "vhd-ikemen-go-fixture.def", "mfj2-classic-mugen-fixture.def",
		"gms-ikemen-go-fixture.def", "truncated-real-fixture.def", "game-default-fight.def",
	} {
		f.Add(fixture(f, name))
	}
	f.Add("[Lifebar\n[]\n[ ]\n[a b]\np*.x|.y = \n= \n\x00\xff")
	f.Fuzz(func(t *testing.T, src string) {
		lb, _ := Parse(src)
		if lb.Animations == nil {
			t.Fatal("Animations is nil")
		}
	})
}

func TestParse_GivesAnEmptyActionTheFramesOfTheNextActionInTheFile(t *testing.T) {
	lb, _ := Parse("[Lifebar]\np1.front.anim = 10\n" +
		"[Begin Action 10]\n" +
		"[Begin Action 11]\n" +
		"[Begin Action 12]\n11,0, 0,0, 7\n" +
		"[Begin Action 13]\n")

	for _, number := range []int{10, 11, 12} {
		a := lb.Animations[number]
		if len(a.Frames) != 1 || a.Frames[0].Group != 11 || a.Frames[0].Time != 7 {
			t.Errorf("Animations[%d] = %+v, want the frame 11,0 of action 12", number, a)
		}
		if a.Number != number {
			t.Errorf("Animations[%d].Number = %d", number, a.Number)
		}
	}
	if a := lb.Animations[13]; len(a.Frames) != 0 {
		t.Errorf("Animations[13] = %+v, want it empty at the end of the file", a)
	}
	if !lb.Life[0].Front.HasAnim || lb.Life[0].Front.Anim != 10 {
		t.Errorf("front = %+v, want anim 10", lb.Life[0].Front)
	}
}

// Copy Action (Ikemen GO, anim.go:359-370): the action copies action 1.
func TestParse_CopyActionCopiesTheNamedAction(t *testing.T) {
	lb, warnings := Parse("[Begin Action 1]\n1,0, 0,0, 5\n" +
		"[Begin Action 2]\nCopy Action 1\n" +
		"[Begin Action 3]\n3,0, 0,0, 7\n")

	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if a := lb.Animations[2]; len(a.Frames) != 1 || a.Frames[0].Group != 1 {
		t.Errorf("Animations[2] = %+v, want action 1's frame", a)
	}
	if len(lb.Animations[1].Frames) != 1 || len(lb.Animations[3].Frames) != 1 {
		t.Errorf("Animations = %+v, want actions 1 and 3 kept", lb.Animations)
	}
}

func TestParse_ReadsTeammateFaceAndNameKeys(t *testing.T) {
	lb, _ := Parse("[Face]\n" +
		"p1.teammate.pos = 5,6\n" +
		"p1.teammate.spacing = 7,8\n" +
		"p1.teammate.face.spr = 9,3\n" +
		"p1.teammate.ko.hide = 1\n" +
		"p1.teammate.face.palshare = 0\n" +
		"[Name]\n" +
		"p1.teammate.pos = 1,2\n" +
		"p1.teammate.spacing = 3,4\n" +
		"p1.teammate.name.text = \"Bob\"\n" +
		"p1.teammate.ko.hide = 1\n")

	f := lb.Face[0].Teammate
	if f.Pos != [2]int{5, 6} || f.Spacing != [2]int{7, 8} || f.FaceSprite != (SpriteRef{9, 3}) ||
		!f.KoHide || f.FacePalShare {
		t.Errorf("face teammate = %+v", f)
	}
	if lb.Face[1].Teammate.KoHide || !lb.Face[1].Teammate.FacePalShare {
		t.Errorf("player 2 face teammate = %+v, want defaults", lb.Face[1].Teammate)
	}
	n := lb.Name[0].Teammate
	if n.Pos != [2]int{1, 2} || n.Spacing != [2]int{3, 4} || n.Name.Text != "Bob" || !n.KoHide {
		t.Errorf("name teammate = %+v", n)
	}
}

func TestParse_ReadsTheTrailingMeterAndBarFlags(t *testing.T) {
	lb, _ := Parse("[Lifebar]\n" +
		"p1.range.y = 2,9\n" +
		"leaderontop = 1\n" +
		"mid.shift = 1\n" +
		"mid.delay = 12\n" +
		"mid.mult = 1.5\n" +
		"mid.steps = 3\n" +
		"[Powerbar]\n" +
		"p1.levelbars = 1\n")

	l := lb.Life[0]
	if l.RangeY != [2]int{2, 9} || !l.LeaderOnTop || !l.MidShift || l.MidDelay != 12 ||
		l.MidMult != 1.5 || l.MidSteps != 3 {
		t.Errorf("life = %+v", l)
	}
	if !lb.Life[1].LeaderOnTop || lb.Life[1].RangeY != [2]int{0, 0} {
		t.Errorf("player 2 life = %+v, want the shared leaderontop and no range.y", lb.Life[1])
	}
	if !lb.Power[0].LevelBars || lb.Power[1].LevelBars {
		t.Errorf("levelbars = %v %v, want true false", lb.Power[0].LevelBars, lb.Power[1].LevelBars)
	}
}

func TestParse_RaisesTheTrailingMeterStepsToOne(t *testing.T) {
	lb, _ := Parse("[Lifebar]\nmid.steps = 0.5\n")

	if lb.Life[0].MidSteps != 1 {
		t.Errorf("MidSteps = %v, want 1", lb.Life[0].MidSteps)
	}
}

func TestParse_ReadsRoundTimesAndAnnouncementsOfEachOutcome(t *testing.T) {
	lb, _ := Parse("[Round]\n" +
		"fight.time = 15\n" +
		"dko.time = 11\n" +
		"to.time = 12\n" +
		"draw.displaytime = 13\n" +
		"round.single.displaytime = 14\n" +
		"win.time = 16\n")

	r := lb.Round
	if r.FightTime != 15 || r.FightSndTime != 15 || r.DkoTime != 11 || r.ToTime != 12 ||
		r.Draw.DisplayTime != 13 || r.RoundSingle.DisplayTime != 14 ||
		r.WinTime != 16 || r.WinSndTime != 16 {
		t.Errorf("round = %+v", r)
	}
	if r.KoTime != 0 || r.DkoSndTime != 0 || r.RoundDefault.DisplayTime != -2 {
		t.Errorf("round = %+v, want the other times at their defaults", r)
	}
}

func TestParse_ClampsTheSlowSpeedAndOverTimesOnBothSides(t *testing.T) {
	low, _ := Parse("[Round]\nslow.speed = 0.001\n" +
		"over.hittime = 0\nover.waittime = -5\nover.wintime = 0\nover.forcewintime = -1\nover.time = 0\n")
	r := low.Round
	if r.SlowSpeed != 0.01 {
		t.Errorf("SlowSpeed = %v, want the 0.01 floor", r.SlowSpeed)
	}
	if r.OverHitTime != 1 || r.OverWaitTime != 1 || r.OverWinTime != 1 || r.OverForceWinTime != 1 || r.OverTime != 1 {
		t.Errorf("over times = %d %d %d %d %d, want 1 each",
			r.OverHitTime, r.OverWaitTime, r.OverWinTime, r.OverForceWinTime, r.OverTime)
	}

	ok, _ := Parse("[Round]\nslow.speed = 0.5\n" +
		"over.hittime = 3\nover.waittime = 4\nover.wintime = 5\nover.forcewintime = 6\nover.time = 7\n")
	r = ok.Round
	if r.SlowSpeed != 0.5 {
		t.Errorf("SlowSpeed = %v, want 0.5 kept", r.SlowSpeed)
	}
	if r.OverHitTime != 3 || r.OverWaitTime != 4 || r.OverWinTime != 5 || r.OverForceWinTime != 6 || r.OverTime != 7 {
		t.Errorf("over times = %d %d %d %d %d, want 3 4 5 6 7",
			r.OverHitTime, r.OverWaitTime, r.OverWinTime, r.OverForceWinTime, r.OverTime)
	}
}

func TestParse_ReadsComboAutoAlignAndClampsShowSpeedOnBothSides(t *testing.T) {
	lb, _ := Parse("[Combo]\n" +
		"team1.pos = 1,2\n" +
		"team1.autoalign = 0\n" +
		"team1.showspeed = 5\n" +
		"showspeed = 0.2\n" +
		"[Action]\n" +
		"team1.showspeed = 0.3\n" +
		"team2.showspeed = 6\n")

	c1, c2 := lb.Combo[0], lb.Combo[1]
	if c1.AutoAlign || c1.ShowSpeed != 5 {
		t.Errorf("combo 1 = autoalign %v showspeed %v, want false 5", c1.AutoAlign, c1.ShowSpeed)
	}
	if !c2.AutoAlign || c2.ShowSpeed != 1 {
		t.Errorf("combo 2 = autoalign %v showspeed %v, want true 1 (shared key 0.2 raised)", c2.AutoAlign, c2.ShowSpeed)
	}
	if lb.Action[0].ShowSpeed != 1 || lb.Action[1].ShowSpeed != 6 {
		t.Errorf("action showspeed = %v %v, want 1 6", lb.Action[0].ShowSpeed, lb.Action[1].ShowSpeed)
	}
}

func TestIsVersionZero_FollowsIkemenVersionParsing(t *testing.T) {
	cases := map[string]bool{
		"": true, "0": true, "0.0": true, "0.0.7": true, "x": true,
		"0.5": false, "1": false, "1.0": false,
	}
	for version, want := range cases {
		if got := isVersionZero(version); got != want {
			t.Errorf("isVersionZero(%q) = %v, want %v", version, got, want)
		}
	}
}

func TestParse_PicksTheComboSyntaxFromTheLastInfoBeforeTheCombo(t *testing.T) {
	combo := "[Combo]\nteam*.pos = 10,20\n"
	cases := []struct {
		name, src string
		wantPos   [2]int
	}{
		{"info before combo, Ikemen syntax", "[Info]\nikemenversion = 1.0\n" + combo, [2]int{10, 20}},
		{"info after combo, legacy syntax", combo + "[Info]\nikemenversion = 1.0\n", [2]int{0, 0}},
		{"two infos, the last one is legacy", "[Info]\nikemenversion = 1.0\n[Info]\nikemenversion = 0.0\n" + combo, [2]int{0, 0}},
		{"two infos, the last one is Ikemen", "[Info]\nikemenversion = 0.0\n[Info]\nikemenversion = 1.0\n" + combo, [2]int{10, 20}},
	}
	for _, c := range cases {
		lb, _ := Parse(c.src)
		if lb.Combo[0].Pos != c.wantPos {
			t.Errorf("%s: Combo[0].Pos = %v, want %v", c.name, lb.Combo[0].Pos, c.wantPos)
		}
	}
}

// fightscreen.go:3484-3520: each side's win call reads p1. / p2. keys over
// the shared win. ones.
// fightscreen.go:3457-3468 readATS: a side whose pN.win.text, spr or anim
// exists reads its whole call from pN.win.*, otherwise wholly from win.*.
func TestParse_ReadsEachSidesWinCall(t *testing.T) {
	lb, _ := Parse("[Round]\n" +
		"win.text = \"%s Wins\"\n" +
		"win.displaytime = 90\n" +
		"win.snd = 5,0\n" +
		"p2.win.text = \"%s Wins!\"\n" +
		"p2.win.offset = 3,4\n")
	w1, w2 := lb.Round.Win[0], lb.Round.Win[1]
	if w1.Text.Text != "%s Wins" || w1.DisplayTime != 90 || w1.Sound != (SoundRef{5, 0}) {
		t.Errorf("p1 win = %+v", w1)
	}
	if w2.Text.Text != "%s Wins!" || w2.Text.Layout.Offset != [2]float32{3, 4} {
		t.Errorf("p2 win = %+v, want its own text and offset", w2)
	}
	if w2.DisplayTime != -2 || w2.Sound.Group != -1 {
		t.Errorf("p2 win = %+v, want the defaults, not the shared win.* keys", w2)
	}
	none, _ := Parse("[Round]\n")
	if none.Round.Win[0].DisplayTime != -2 || none.Round.Win[0].Sound.Group != -1 {
		t.Errorf("no win keys: %+v, want the defaults", none.Round.Win[0])
	}
}

func TestParse_APlayerKeyWithoutTextSprOrAnimIsIgnored(t *testing.T) {
	lb, _ := Parse("[Round]\n" +
		"win.text = \"%s Wins\"\n" +
		"win.snd = 5,0\n" +
		"p1.win.snd = 7,0\n")
	if got := lb.Round.Win[0].Sound; got != (SoundRef{5, 0}) {
		t.Errorf("p1 win sound = %+v, want win.snd 5,0 (no p1.win.text/spr/anim)", got)
	}
}

func TestParse_APlayerSpriteSelectsThePlayerKeys(t *testing.T) {
	lb, _ := Parse("[Round]\n" +
		"win.text = \"%s Wins\"\n" +
		"p2.win.spr = 3,1\n")
	w2 := lb.Round.Win[1]
	if w2.Text.Text != "" || w2.Element.Sprite != (SpriteRef{3, 1}) {
		t.Errorf("p2 win = %+v, want sprite 3,1 and no shared text", w2)
	}
}

// fightscreen.go:378-396, 640-655: a life bar's front50 shows from 50 % of
// life on, the plain front below; a power bar's frontmax once power is full
// (readPBKeys, resolvePBKey).
func TestParse_PerValueBarElements(t *testing.T) {
	lb, _ := Parse("[Lifebar]\n" +
		"p1.front.spr = 1,0\n" +
		"p1.front50.spr = 1,50\n" +
		"p1.front80.spr = 1,80\n" +
		"[Powerbar]\n" +
		"p1.front.spr = 2,0\n" +
		"p1.front1000.spr = 2,1\n" +
		"p1.frontmax.spr = 2,9\n")
	life := lb.Life[0].Bar
	for pct, want := range map[float64]int{10: 0, 50: 50, 79: 50, 100: 80} {
		if got := Pick(life.Front, life.FrontVariants, pct, 100).Sprite.Image; got != want {
			t.Errorf("life %v%%: front image %d, want %d", pct, got, want)
		}
	}
	power := lb.Power[0].Bar
	for value, want := range map[float64]int{500: 0, 1500: 1, 3000: 9} {
		if got := Pick(power.Front, power.FrontVariants, value, 3000).Sprite.Image; got != want {
			t.Errorf("power %v: front image %d, want %d", value, got, want)
		}
	}
	if len(lb.Life[1].FrontVariants) != 0 {
		t.Errorf("p2 variants = %+v, want none", lb.Life[1].FrontVariants)
	}
}
