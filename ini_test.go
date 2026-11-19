package lifebar

import (
	"reflect"
	"testing"
)

func TestSectionName_SplitsHeaderAtFirstSpace(t *testing.T) {
	cases := []struct {
		line, inner, name, subname string
	}{
		{"[Lifebar]", "Lifebar", "lifebar", ""},
		{"[Simul Lifebar]", "Simul Lifebar", "simul ", "Lifebar"},
		{"[Begin Action 5]", "Begin Action 5", "begin ", "Action 5"},
		{"[Info] ; comment", "Info", "info", ""},
		{"[Lifebar ]", "Lifebar ", "lifebar ", ""},
	}
	for _, c := range cases {
		inner, name, subname := sectionName(c.line)
		if inner != c.inner || name != c.name || subname != c.subname {
			t.Errorf("sectionName(%q) = (%q, %q, %q), want (%q, %q, %q)",
				c.line, inner, name, subname, c.inner, c.name, c.subname)
		}
	}
}

func TestSectionName_RejectsLinesThatAreNotHeaders(t *testing.T) {
	for _, line := range []string{"", "Lifebar]", "[Lifebar", "[]", "[", "x = [y]", " [Lifebar]"} {
		if _, name, _ := sectionName(line); name != "" {
			t.Errorf("sectionName(%q) name = %q, want no header", line, name)
		}
	}
}

func TestTokenize_KeyEndsAtFirstEqualsSpaceOrTab(t *testing.T) {
	sections := tokenize("[S]\nA b = 1\nc\t= 2\nd=3 ; x\n=4\nE\nf g\n")
	want := iniKeys{"a": "1", "c": "2", "d": "3", "f": ""}
	if !reflect.DeepEqual(sections[0].keys, want) {
		t.Errorf("keys = %v, want %v", sections[0].keys, want)
	}
}

func TestTokenize_KeepsFirstValueOfARepeatedKey(t *testing.T) {
	sections := tokenize("[S]\nk = 1\nK = 2\n")
	if got := sections[0].keys["k"]; got != "1" {
		t.Errorf("k = %q, want 1", got)
	}
	if !reflect.DeepEqual(sections[0].order, []string{"k"}) {
		t.Errorf("order = %v, want [k]", sections[0].order)
	}
}

func TestTokenize_NumbersHeadersByTheirOneBasedLine(t *testing.T) {
	sections := tokenize("; c\n\n[A]\nx=1\n[B]\n")
	if len(sections) != 2 || sections[0].line != 3 || sections[1].line != 5 {
		t.Errorf("sections = %+v, want lines 3 and 5", sections)
	}
}

func TestTokenize_DropsKeysBetweenABrokenHeaderAndTheNextHeader(t *testing.T) {
	sections := tokenize("[A]\nx=1\n[B\ny=2\n[C]\nz=3\n")
	if len(sections) != 2 {
		t.Fatalf("got %d sections, want 2 (A and C)", len(sections))
	}
	if _, ok := sections[0].keys["y"]; ok || sections[0].keys["x"] != "1" {
		t.Errorf("A keys = %v, want only x", sections[0].keys)
	}
	if sections[1].name != "c" || sections[1].keys["z"] != "3" {
		t.Errorf("second section = %+v, want C with z", sections[1])
	}
}

func TestTokenize_StripsTheUTF8ByteOrderMark(t *testing.T) {
	sections := tokenize("\xef\xbb\xbf[Info]\nx=1\n")
	if len(sections) != 1 || sections[0].name != "info" {
		t.Errorf("sections = %+v, want one info section", sections)
	}
}

func TestForPlayer_ExplicitKeyBeatsSelectorWhateverTheOrder(t *testing.T) {
	for _, src := range []string{
		"[S]\np*.pos = 1\np1.pos = 2\n",
		"[S]\np1.pos = 2\np*.pos = 1\n",
		"[S]\np1|p3.pos = 1\np1.pos = 2\n",
	} {
		keys := tokenize(src)[0].forPlayer("p", 1)
		if got := keys["p1.pos"]; got != "2" {
			t.Errorf("%q: p1.pos = %q, want 2", src, got)
		}
	}
}

func TestForPlayer_FirstSelectorInFileWins(t *testing.T) {
	keys := tokenize("[S]\np1|p2.pos = A\np*.pos = B\n")[0].forPlayer("p", 2)
	if got := keys["p2.pos"]; got != "A" {
		t.Errorf("p2.pos = %q, want A", got)
	}
	keys = tokenize("[S]\np*.pos = B\np1|p2.pos = A\n")[0].forPlayer("p", 2)
	if got := keys["p2.pos"]; got != "B" {
		t.Errorf("p2.pos = %q, want B", got)
	}
}

func TestForPlayer_LeavesSelectorsOfOtherPlayersOut(t *testing.T) {
	keys := tokenize("[S]\np1|p3.pos = 1\n")[0].forPlayer("p", 2)
	if _, ok := keys["p2.pos"]; ok {
		t.Errorf("p2.pos = %q, want it unset", keys["p2.pos"])
	}
}

func TestIntsFloats_ReadLeadingNumbersLikeIkemen(t *testing.T) {
	keys := iniKeys{"i": " 12abc, -3x ,,7", "f": ".5, 2.", "big": "99999999999, -99999999999", "neg": "-2147483648"}
	var a, b, c int
	keys.ints("i", &a, &b, &c)
	if a != 12 || b != -3 || c != 0 {
		t.Errorf("ints = %d %d %d, want 12 -3 0", a, b, c)
	}
	var x, y float32
	keys.floats("f", &x, &y)
	if x != 0.5 || y != 2 {
		t.Errorf("floats = %v %v, want 0.5 2", x, y)
	}
	keys.ints("big", &a, &b)
	if a != 2147483647 || b != -2147483648 {
		t.Errorf("saturated ints = %d %d, want 2147483647 -2147483648", a, b)
	}
	keys.ints("neg", &a)
	if a != -2147483648 {
		t.Errorf("neg = %d, want -2147483648", a)
	}
}

func TestInts_ReportMissingAndEmptyValuesAsAbsent(t *testing.T) {
	keys := iniKeys{"empty": ""}
	var a int
	if keys.ints("empty", &a) || keys.ints("nokey", &a) {
		t.Error("ints reported a value for an empty or missing key")
	}
}

func TestAtof_ReadsLeadingNumbersLikeIkemenAtof(t *testing.T) {
	cases := []struct {
		in   string
		want float32
	}{
		{"1e2", 100},
		{"1e-2", 0.01},
		{"-1e2", -100},
		{"-1e-2", -100}, // Ikemen GO reads the exponent sign from the signed text, so it stays positive
		{"+.5", 0.5},
		{"-.5", -0.5},
		{"1.2.3", 1.2},
		{"e5", 0},
		{"5e", 5},
		{"", 0},
	}
	for _, c := range cases {
		if got := atof(c.in); got != c.want {
			t.Errorf("atof(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
