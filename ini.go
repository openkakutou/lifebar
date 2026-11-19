package lifebar

import (
	"fmt"
	"math"
	"strings"
)

// The tokenizer below copies Ikemen GO's INI reading (src/common.go @
// 54d4148a: SectionName :862, ReadIniSection :1054, IniSection.Parse :1071),
// quirks included, because fight.def files in the wild depend on them.

// iniKeys maps a lowercased key to its value.
type iniKeys map[string]string

// iniSection is one section header with the keys that follow it.
type iniSection struct {
	// inner is the text between the brackets, as written.
	inner string
	// name is the lowercased text up to and including the first space;
	// subname is what follows that space.
	name, subname string
	// line is the 1-based line of the header.
	line  int
	keys  iniKeys
	order []string // keys in file order
	lines []string // header line first
}

// sectionName reads a header line. name is empty when line is not a header.
func sectionName(line string) (inner, name, subname string) {
	if line == "" || line[0] != '[' {
		return "", "", ""
	}
	line = cutComment(line)
	if line[len(line)-1] != ']' {
		return "", "", ""
	}
	inner = line[1:strings.Index(line, "]")]
	name, subname = inner, ""
	if i := strings.Index(inner, " "); i >= 0 {
		name, subname = inner[:i+1], inner[i+1:]
	}
	return inner, strings.ToLower(name), subname
}

func cutComment(line string) string {
	before, _, _ := strings.Cut(line, ";")
	return strings.TrimSpace(before)
}

// tokenize splits src into sections. Ikemen GO trims every line before it
// looks for headers, so an indented "[Lifebar]" is a header. A line that
// starts with "[" ends the keys of the section above it even when it is not
// a valid header; its own keys then belong to no section.
func tokenize(src string) []iniSection {
	src = strings.TrimPrefix(src, "\xef\xbb\xbf")
	lines := strings.Split(src, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	var sections []iniSection
	for i := 0; i < len(lines); {
		inner, name, subname := sectionName(lines[i])
		if name == "" {
			i++
			continue
		}
		section := iniSection{inner: inner, name: name, subname: subname, line: i + 1, keys: iniKeys{}}
		start := i
		for i++; i < len(lines) && !strings.HasPrefix(lines[i], "["); i++ {
			section.addKey(lines[i])
		}
		section.lines = lines[start:i]
		sections = append(sections, section)
	}
	return sections
}

// addKey reads "key = value". The key ends at the first '=', space or tab,
// so "foo bar" stores foo with an empty value, which hides a later "foo = 1".
// The first value of a key wins.
func (s *iniSection) addKey(line string) {
	line = cutComment(line)
	end := strings.IndexAny(line, "= \t")
	if end <= 0 {
		return
	}
	key := strings.ToLower(line[:end])
	if _, seen := s.keys[key]; seen {
		return
	}
	value := ""
	if _, after, found := strings.Cut(line, "="); found {
		value = strings.TrimSpace(after)
	}
	s.keys[key] = value
	s.order = append(s.order, key)
}

// forPlayer returns the keys of the section as player n (prefixBase "p", as
// in "p1.pos") or team n (prefixBase "team") reads them: "p*.pos" and
// "p1|p3.pos" become "p1.pos" (Ikemen GO resolveLifebarPrefixOperators,
// common.go:988). An explicit "p1.pos" beats any selector. Between selectors
// the first in the file wins; Ikemen GO iterates a map there, so its winner
// is random.
func (s iniSection) forPlayer(prefixBase string, n int) iniKeys {
	target := fmt.Sprintf("%s%d", prefixBase, n)
	wildcard := prefixBase + "*"
	resolved := iniKeys{}
	var selected []string
	for _, key := range s.order {
		if isSelector(key, target, wildcard) {
			selected = append(selected, key)
			continue
		}
		resolved[key] = s.keys[key]
	}
	for _, key := range selected {
		_, rest, _ := strings.Cut(key, ".")
		dest := target + "." + rest
		if _, taken := resolved[dest]; !taken {
			resolved[dest] = s.keys[key]
		}
	}
	return resolved
}

// isSelector reports whether the part of key before its first dot is the
// wildcard or a "|" group that includes the target.
func isSelector(key, target, wildcard string) bool {
	selector, _, found := strings.Cut(key, ".")
	if !found || selector == "" {
		return false
	}
	if selector == wildcard {
		return true
	}
	if !strings.Contains(selector, "|") {
		return false
	}
	for _, token := range strings.Split(selector, "|") {
		token = strings.TrimSpace(token)
		if token == target || token == wildcard {
			return true
		}
	}
	return false
}

// csv returns the comma-separated fields of a key, trimmed. ok is false for
// a missing or empty key.
func (k iniKeys) csv(name string) (fields []string, ok bool) {
	value := k[name]
	if value == "" {
		return nil, false
	}
	fields = strings.Split(value, ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}
	return fields, true
}

// ints fills out from the comma-separated values of a key, leaving the
// entries whose value is empty or missing as they are. It reports whether the
// key has a value.
func (k iniKeys) ints(name string, out ...*int) bool {
	fields, ok := k.csv(name)
	for i, field := range fields {
		if i < len(out) && field != "" {
			*out[i] = atoi(field)
		}
	}
	return ok
}

// floats is ints for float32 values.
func (k iniKeys) floats(name string, out ...*float32) bool {
	fields, ok := k.csv(name)
	for i, field := range fields {
		if i < len(out) && field != "" {
			*out[i] = atof(field)
		}
	}
	return ok
}

// flag sets out to whether the first value of a key is a non-zero number.
func (k iniKeys) flag(name string, out *bool) {
	fields, _ := k.csv(name)
	if len(fields) > 0 && fields[0] != "" {
		*out = atoi(fields[0]) != 0
	}
}

// quoted returns the value of a key with its surrounding quotes removed; a
// value without quotes is returned as written (Ikemen GO getText,
// common.go:1305).
func (k iniKeys) quoted(name string) (value string, ok bool) {
	value, ok = k[name]
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		value = value[1 : len(value)-1]
	}
	return value, ok
}

// atoi reads the leading integer of s and saturates at the int32 range
// (Ikemen GO Atoi, common.go:191).
func atoi(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	negative := s[0] == '-'
	digits := s
	if negative || s[0] == '+' {
		digits = s[1:]
	}
	limit := int64(math.MaxInt32)
	if negative {
		limit++
	}
	var n int64
	for i := 0; i < len(digits) && digits[i] >= '0' && digits[i] <= '9'; i++ {
		n = n*10 + int64(digits[i]-'0')
		if n > limit && negative {
			return math.MinInt32
		}
		if n > limit {
			return math.MaxInt32
		}
	}
	if negative {
		n = -n
	}
	return int(n)
}

// atof reads the leading decimal number of s (Ikemen GO Atof, common.go:227).
func atof(s string) float32 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	digits := s
	if s[0] == '-' || s[0] == '+' {
		digits = s[1:]
	}
	value := 0.0
	i, dot := 0, 0
	for ; i < len(digits); i++ {
		if digits[i] == '.' {
			if dot != 0 {
				break
			}
			dot = i + 1
			continue
		}
		if digits[i] < '0' || digits[i] > '9' {
			break
		}
		value = value*10 + float64(digits[i]-'0')
	}
	exponent := 0.0
	if i+1 < len(digits) && (digits[i] == 'e' || digits[i] == 'E') {
		j := i + 1
		if digits[j] == '-' || digits[j] == '+' {
			j++
		}
		for ; j < len(digits) && digits[j] >= '0' && digits[j] <= '9'; j++ {
			exponent = exponent*10 + float64(digits[j]-'0')
		}
		if exponent != 0 {
			// Ikemen GO reads the sign from s, not digits: after a leading
			// sign it sees the 'e' and the exponent stays positive.
			if s[i+1] == '-' {
				exponent = -exponent
			}
			if dot == 0 {
				dot = i
			}
		}
	}
	if dot > 0 {
		value *= math.Pow10(dot - i + int(exponent))
	}
	if s[0] == '-' {
		value = -value
	}
	return float32(value)
}
