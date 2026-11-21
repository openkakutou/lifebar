package lifebar

import "github.com/openkakutou/character/air"

// Lifebar is a parsed fight.def, as Ikemen GO's loadFightScreen reads it
// (src/fightscreen.go @ 54d4148a). Every modelled section is always filled:
// when the file has none, the section is read from empty text, so each field
// holds the default Ikemen GO gives it. Index 0 is player 1 (or team 1) and
// index 1 is player 2 (or team 2).
//
// Keys that only matter to a mode other than single (the [Simul ...],
// [Turns ...] and [Tag ...] families), and the sections listed in
// unmodelledSections, are not read yet.
type Lifebar struct {
	// Localcoord is [Info] localcoord, 320x240 when absent.
	Localcoord [2]int
	Files      Files
	Life       [2]PlayerLife
	Power      [2]PowerBar
	Face       [2]Face
	Name       [2]Name
	WinIcon    [2]WinIcon
	Time       Time
	Combo      [2]Combo
	Round      Round
	Action     [2]Action
	// Animations holds the [Begin Action N] blocks by number. A block that
	// does not parse is left out and reported as a Warning; the others are
	// kept. A block without frames holds the frames of the next block.
	Animations map[int]air.Animation
}

// Warning reports a section Ikemen GO's loader skips, or a [Begin Action N]
// block that could not be read.
type Warning struct {
	// Section is the text between the brackets, as written.
	Section string
	// Line is the 1-based line of the section header.
	Line int
	// Reason says why a [Begin Action N] block was dropped (the error of the
	// animation parser). It is empty for a section Ikemen GO's loader skips.
	Reason string
}

// Files lists the files [Files] references, as written. They are not loaded.
type Files struct {
	Sprites           string // sff
	Sounds            string // snd
	FightFxSprites    string // fightfx.sff
	FightFxAnimations string // fightfx.air
	FightFxSounds     string // common.snd
	FightFx           [3]string
	// Fonts is keyed by the font number of the fontN key.
	Fonts map[int]Font
}

// Font is a fontN entry of [Files].
type Font struct {
	File string
	// Height is fontN.height, 0 when absent.
	Height int
}

// SpriteRef names a sprite of an sff by group and image number.
type SpriteRef struct {
	Group int `json:"group"`
	Image int `json:"image"`
}

// SoundRef names a sound of a snd by group and sound number.
type SoundRef struct {
	Group  int `json:"group"`
	Number int `json:"number"`
}

// Layout is the placement every element key set shares.
type Layout struct {
	Offset  [2]float32
	Facing  int // 1 or -1
	VFacing int // 1 or -1
	// LayerNo is capped at 2, as Ikemen GO caps it.
	LayerNo int
	Scale   [2]float32
}

// Element is a sprite or animation with its layout (Ikemen GO's AnimLayout).
type Element struct {
	// HasSprite reports whether the spr key was written; Sprite is its group
	// and number.
	HasSprite bool
	Sprite    SpriteRef
	// HasAnim reports whether the anim key names an action of Animations;
	// Ikemen GO ignores an anim key whose action does not exist. An
	// animation replaces the sprite when both are set.
	HasAnim bool
	Anim    int
	Layout  Layout
}

// TextFont is the font key of a text: font number, bank, alignment, then
// the optional color and alpha.
type TextFont struct {
	No    int // -1 when unset
	Bank  int
	Align int
	Color [3]int // -1 when unset
	Alpha int
}

// Text is a text element (Ikemen GO's FSText).
type Text struct {
	Font TextFont
	// Text is the text key with its quotes removed.
	Text   string
	Layout Layout
}

// Bar holds what the life and power bars share.
type Bar struct {
	Pos             [2]int
	RangeX, RangeY  [2]int
	Bg0, Bg1, Bg2   Element
	Mid, Front, Top Element
	LeaderOnTop     bool
}

// PlayerLife is a [Lifebar] player.
type PlayerLife struct {
	Bar
	// The trailing meter settings are read without a player prefix.
	MidShift  bool
	MidFreeze bool // true by default
	MidDelay  int  // 30 by default
	MidMult   float32
	MidSteps  float32 // at least 1
}

// PowerBar is a [Powerbar] player.
type PowerBar struct {
	Bar
	Counter   Text
	LevelBars bool
}

// Face is a [Face] player.
type Face struct {
	Pos                        [2]int
	Bg, Bg0, Bg1, Bg2, Top, Ko Element
	FaceSprite                 SpriteRef // face.spr, -1,0 by default
	FaceLayout                 Layout
	FacePalShare               bool // true by default
	FacePalFxShare             bool
	FaceDarkenShare            bool
	LeaderOnTop                bool
	Teammate                   FaceTeammate
}

// FaceTeammate is the teammate part of a Face.
type FaceTeammate struct {
	Pos, Spacing               [2]int
	Bg, Bg0, Bg1, Bg2, Top, Ko Element
	FaceSprite                 SpriteRef // -1,0 by default
	FaceLayout                 Layout
	KoHide                     bool
	FacePalShare               bool // true by default
}

// Name is a [Name] player.
type Name struct {
	Pos         [2]int
	LeaderOnTop bool
	Name        Text
	Bg, Top     Element
	Teammate    NameTeammate
}

// NameTeammate is the teammate part of a Name.
type NameTeammate struct {
	Pos, Spacing [2]int
	Name         Text
	Bg           Element
	KoHide       bool
}

// WinIcon is a [WinIcon] player.
type WinIcon struct {
	Pos, IconOffset [2]int
	UseIconUpTo     int // 4 by default, read without a player prefix
	Counter         Text
	Bg0, Top        Element
	Icons           WinIcons
}

// WinIcons holds one element per win type.
type WinIcons struct {
	Normal, Special, Hyper, Cheese, Time Element
	Throw, Suicide, Teammate             Element
	Perfect, Clutch                      Element
}

// Time is the [Time] section.
type Time struct {
	Pos            [2]int
	Counter        Text
	Bg, Top        Element
	FramesPerCount int // 60 by default
}

// Combo is a [Combo] team.
//
// Team 2 is mirrored across Localcoord as Ikemen GO derives it at load
// (fightscreen.go:2519-2531), depending on the keys it was read from. From
// team2. keys (always with an ikemenversion; in MUGEN syntax when team2.pos
// is written), StartX becomes Localcoord[0] - start.x, or, when that is not
// negative, min(Pos[0] - start.x, 0), and Pos is left as written. From the
// keys shared by both teams (MUGEN syntax without team2.pos), StartX is left
// as written and Pos[0] becomes Localcoord[0] - pos.x.
type Combo struct {
	Pos         [2]int
	StartX      float32
	Counter     Text
	Text        Text
	Bg, Top     Element
	DisplayTime int     // 90 by default
	ShowSpeed   float32 // at least 1, 8 by default
	HideSpeed   float32 // 4 by default
	AutoAlign   bool    // true by default
}

// Action is an [Action] team. Its message definitions are not read yet.
//
// The second team's StartX is mirrored across Localcoord as Ikemen GO derives
// it at load (fightscreen.go:3094-3100).
type Action struct {
	Pos, Spacing [2]int
	StartX       float32
	Text         Text
	DisplayTime  int     // 90 by default
	ShowSpeed    float32 // at least 1, 8 by default
	HideSpeed    float32 // 4 by default
	Max          int     // 8 by default
}

// Announcement is a sound, text or animation shown at a round step (Ikemen
// GO's AnimTextSnd).
type Announcement struct {
	Sound       SoundRef // -1,0 by default
	Text        Text
	Element     Element
	DisplayTime int // -2 by default
}

// Round is the [Round] section. Times are in ticks. Its win2-4, ai.win and
// ai.lose announcements, fade and shutter colors, per-win-type texts and
// bg0-31 layers are not read yet.
type Round struct {
	Pos           [2]int
	StartWaitTime int // at least 1, 30 by default
	RoundTime     int
	RoundSndTime  int // RoundTime by default
	Round         [9]Announcement
	RoundDefault  Announcement
	RoundSingle   Announcement
	RoundFinal    Announcement
	FightTime     int
	FightSndTime  int // FightTime by default
	Fight         Announcement
	CtrlTime      int // at least 1, 30 by default
	KoTime        int
	KoSndTime     int // KoTime by default
	Ko            Announcement
	DkoTime       int // KoTime by default
	DkoSndTime    int // KoSndTime by default
	DkoShowDraw   bool
	Dko           Announcement
	ToTime        int // KoTime by default
	ToSndTime     int // KoSndTime by default
	To            Announcement
	Draw          Announcement
	// Win is each side's win call (p1.win.* over win.*); its text's %s is
	// the winner's name.
	Win      [2]Announcement
	SlowTime int // 60 by default
	// SlowFadeTime is 75% of SlowTime by default and never above it.
	SlowFadeTime int
	// SlowSpeed is kept within 0.01 and 1.
	SlowSpeed        float32 // 0.25 by default
	OverHitTime      int     // at least 1, 10 by default
	OverWaitTime     int     // at least 1, 45 by default
	OverWinTime      int     // at least 1, 45 by default
	OverForceWinTime int     // at least 1, 900 by default
	OverTime         int     // at least 1, 210 by default
	WinTime          int
	WinSndTime       int // WinTime by default
	ShutterTime      int // 15 by default
	ClutchThreshold  int // 10 by default
	CallFightTime    int // 60 by default
}
