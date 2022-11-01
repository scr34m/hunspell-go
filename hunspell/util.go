package hunspell

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
	"unicode"
)

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func parseFlag(s string) rune {
	runes := parseFlags(s)
	return runes[0]
}

func parseFlags(s string) []rune {
	return []rune(s)
}

func (hs *hunSpell) getAliasValue(idx int) string {
	return hs.aliases[idx-1]
}

func escapeDash(re string) string {
	runes := []rune(re)
	escaped := ""
	for i := 0; i < len(runes); i++ {
		if runes[i] == '-' {
			escaped = escaped + "\\-"
		} else {
			escaped = escaped + string(runes[i])
			if runes[i] == '\\' && i+1 < len(runes) {
				escaped = escaped + string(runes[i+1])
				i++
			}
		}
	}
	return escaped
}

func encodeFlags(flags []rune) []byte {
	b := []byte{}
	for i := 0; i < len(flags); i++ {
		flag := flags[i]
		b = append(b, (byte)((flag>>8)&0xff))
		b = append(b, (byte)(flag&0xff))
	}
	return b
}

func (hs *hunSpell) flagLookup_add(b []byte) int {
	h := hash(b)
	pos, ok := hs.flagLookup[h]
	if !ok {
		pos = len(hs.flags)
		hs.flags = append(hs.flags, b)
		hs.flagLookup[h] = pos
	}
	return pos
}

func hash(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func decodeFlags(b []byte) []rune {
	if len(b) == 0 {
		return nil
	}

	var flags []rune

	for i := 0; i < len(b); i += 2 {
		r := (b[i] << 8) + (b[i+1] & 0xff)
		flags = append(flags, rune(r))
	}
	return flags
}

func hasFlag(flags []rune, flag rune) bool {
	for i := 0; i < len(flags); i++ {
		if flags[i] == flag {
			return true
		}
	}
	return false
}

func hasCrossCheckedFlag(flag rune, flags []rune, matchEmpty bool) bool {
	return (len(flags) == 0 && matchEmpty) || hasFlag(flags, flag)
}

const (
	ExactCase = iota
	TileCase
	UpperCase
)

func (hs *hunSpell) caseOf(word string) int {
	runes := []rune(word)

	if hs.ignoreCase || len(runes) == 0 || !unicode.IsUpper(runes[0]) {
		return ExactCase
	}

	// determine if we are title or lowercase (or something funky, in which it's exact)
	seenUpper := false
	seenLower := false
	for _, r := range runes {
		if unicode.IsLetter(r) {
			if unicode.IsUpper(r) {
				seenUpper = true
			} else {
				seenLower = true
			}
		}
	}

	if !seenLower {
		return UpperCase
	} else if !seenUpper {
		return TileCase
	} else {
		return ExactCase
	}
}

func (hs *hunSpell) cleanInput(s string) string {
	if hs.needsInputCleaning {
		return strings.ToLower(s)
	} else {
		return s
	}
}
