package hunspell

import (
	"io"
	"regexp"
)

type affixDataEntry struct {
	flag         rune
	stripOrd     int
	patternOrd   int // aka. condition
	crossProduct bool
	append       rune
}

type dictEntry struct {
	word  string
	flags []rune
}

type hunSpell struct {
	aliasCount int
	aliases    []string

	seenPatterns map[string]int
	patterns     []*regexp.Regexp

	seenStrips   map[string]int
	stripData    []rune
	stripOffsets []int

	prefixes      map[string][]int
	suffixes      map[string][]int
	flagToaffixes map[rune][]int
	flags         [][]byte
	flagLookup    map[string]int
	affixData     []affixDataEntry
	currentAffix  int

	dict map[string]dictEntry

	complexPrefixes     bool
	twoStageAffix       bool
	needsInputCleaning  bool
	needsOutputCleaning bool
	fullStrip           bool

	onlyincompound rune
	circumfix      rune
	keepcase       rune

	formStep          int
	hasStemExceptions bool

	ignoreCase      bool
	compoundVersion bool
	overrideMap     map[string]string
}

type HunSpell interface {
	Lookup(string) *dictEntry
	Stem(string) []string
	SetOverrideMap(map[string]string)
}

func NewHunSpellReader(aff, dic io.Reader, ignoreCase bool, compoundVersion bool) (HunSpell, error) {

	h := &hunSpell{
		seenPatterns:        map[string]int{},
		patterns:            []*regexp.Regexp{},
		seenStrips:          map[string]int{},
		prefixes:            map[string][]int{},
		suffixes:            map[string][]int{},
		flagToaffixes:       map[rune][]int{},
		flags:               [][]byte{},
		flagLookup:          map[string]int{},
		affixData:           []affixDataEntry{},
		currentAffix:        0,
		dict:                map[string]dictEntry{},
		complexPrefixes:     false,
		twoStageAffix:       false,
		needsInputCleaning:  ignoreCase,
		needsOutputCleaning: false,
		onlyincompound:      -1,
		circumfix:           -1,
		keepcase:            -1,
		ignoreCase:          ignoreCase,
		compoundVersion:     compoundVersion,
		overrideMap:         map[string]string{},
	}

	err := h.readAffixFile(aff)
	if err != nil {
		return nil, err
	}

	err = h.readDictionaryFile(dic)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (hs *hunSpell) Lookup(s string) *dictEntry {
	r := []rune(s)
	return hs.lookupWord(r, 0, len(r))
}

func (hs *hunSpell) Stem(s string) []string {

	s = hs.cleanInput(s)
	r := []rune(s)

	var stems []string
	if hs.compoundVersion {
		stems = hs.compoundStem(r, len(r))
	} else {
		stems = hs._stem(r, len(r))
	}

	return stems
}

func (hs *hunSpell) SetOverrideMap(m map[string]string) {
	hs.overrideMap = m
}

func (hs *hunSpell) _stem(word []rune, length int) []string {
	f := hs.lookupWord(word, 0, length)
	if f != nil {
		return []string{string(word)}
	}

	caseType := hs.caseOf(word)
	if caseType == UpperCase {
		// upper: union exact, title, lower
		panic("TODO")
	} else if caseType == TileCase {
		// title: union exact, lower
		panic("TODO")
	} else {
		// exact match only
		return hs.stem(word, length, -1, -1, -1, 0, true, true, false, false, false)
	}
}
