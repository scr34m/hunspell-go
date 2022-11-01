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

	ignoreCase bool
}

type HunSpell interface {
	Lookup(string) *dictEntry
	Stem(string) []string
}

func NewHunSpellReader(aff, dic io.Reader, ignoreCase bool) (HunSpell, error) {

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
	r := []rune(hs.cleanInput(s))

	f := hs.lookupWord(r, 0, len(r))
	if f != nil {
		return []string{s}
	}

	// almás -> alma, buffer = stemmer.compoundStem(termAtt.buffer(), termAtt.length());

	caseType := hs.caseOf(s)
	if caseType == UpperCase {
		// upper: union exact, title, lower
		panic("TODO")
	} else if caseType == TileCase {
		// title: union exact, lower
		panic("TODO")
	} else {
		// exact match only
		return hs.stem(r, len(r), -1, -1, -1, 0, true, true, false, false, false)
	}
}
