package hunspell

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (hs *hunSpell) readAffixFile(file io.Reader) error {
	hs.seenPatterns[".*"] = 0
	hs.patterns = append(hs.patterns, nil)

	hs.seenStrips[""] = 0

	hs.flagLookup_add(nil) // no flags -> ord 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "AF") {
			err := hs.parseAlias(line)
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(line, "PFX") {
			err := hs.parseAffix(hs.prefixes, line, scanner, "^%s") // %s.*
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(line, "SFX") {
			err := hs.parseAffix(hs.suffixes, line, scanner, "%s$") // .*%s
			if err != nil {
				return err
			}
		} else if strings.HasPrefix(line, "ONLYINCOMPOUND") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return fmt.Errorf("Illegal ONLYINCOMPOUND declaration")
			}
			hs.onlyincompound = parseFlag(parts[1])
		} else if strings.HasPrefix(line, "CIRCUMFIX") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return fmt.Errorf("Illegal CIRCUMFIX declaration")
			}
			hs.circumfix = parseFlag(parts[1])
		} else if line == "COMPLEXPREFIXES" {
			hs.complexPrefixes = true
		} else if line == "FULLSTRIP" {
			hs.fullStrip = true
		} else if strings.HasPrefix(line, "KEEPCASE") {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				return fmt.Errorf("Illegal KEEPCASE declaration")
			}
			hs.keepcase = parseFlag(parts[1])
		} else if strings.HasPrefix(line, "ICONV") {
			/*
				parts := strings.Fields(line)
				if len(parts) != 2 {
					return fmt.Errorf("Illegal ICONV declaration")
				}
				hs.needsInputCleaning = true
				TODO parseConversions(reader, num);
			*/
		} else if strings.HasPrefix(line, "OCONV") {
			/*
				parts := strings.Fields(line)
				if len(parts) != 2 {
					return fmt.Errorf("Illegal OCONV declaration")
				}
				hs.needsOutputCleaning = true
				parseConversions(reader, num);
			*/
		}
	}

	totalChars := 0
	for strip, _ := range hs.seenStrips {
		totalChars += utf8.RuneCountInString(strip)
	}

	hs.stripData = make([]rune, totalChars)
	hs.stripOffsets = make([]int, len(hs.seenStrips)+1)

	currentOffset := 0
	currentIndex := 0

	// XXX map order not guarenteed, so walk over the map values and find specific index
	for i := 1; i < len(hs.seenStrips); i++ {
		for strip, ord := range hs.seenStrips {
			if ord == i {
				l := utf8.RuneCountInString(strip)
				currentIndex++
				hs.stripOffsets[currentIndex] = currentOffset
				runes := []rune(strip)
				for j := 0; j < l; j++ {
					hs.stripData[currentOffset+j] = runes[j]
				}
				currentOffset += l
				break
			}
		}
	}

	hs.stripOffsets[currentIndex] = currentOffset

	return nil
}

func (hs *hunSpell) parseAlias(line string) error {
	parts := strings.Fields(line)
	if hs.aliases == nil {
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}
		hs.aliases = make([]string, i)
		hs.aliasCount = 0
	} else {
		if len(parts) == 1 {
			hs.aliases[hs.aliasCount] = ""
		} else {
			hs.aliases[hs.aliasCount] = parts[1]
		}
		hs.aliasCount++
	}
	return nil
}

func (hs *hunSpell) parseAffix(affixes map[string][]int, line string, scanner *bufio.Scanner, conditionPattern string) error {

	parts := strings.Fields(line)

	crossProduct := parts[2] == "Y"
	isSuffix := conditionPattern == "%s$"

	numLines, err := strconv.Atoi(parts[3])
	if err != nil {
		return err
	}

	for i := 0; i < numLines; i++ {
		scanner.Scan()
		line = scanner.Text()

		parts = strings.Fields(line)
		if len(parts) < 4 {
			return fmt.Errorf("The affix file contains a rule with less than four elements: %s", line)
		}

		flag := parseFlag(parts[1])
		strip := ""
		if parts[2] != "0" {
			strip = parts[2]
		}
		affixArg := parts[3]

		var appendFlags []rune

		flagSep := strings.Index(affixArg, "/")
		if flagSep != -1 {
			flagPart := affixArg[flagSep+1:]
			affixArg = affixArg[0:flagSep]
			if hs.aliasCount > 0 {
				flagPartInt, err := strconv.Atoi(flagPart)
				if err != nil {
					return err
				}
				flagPart = hs.getAliasValue(flagPartInt)
			}
			appendFlags = parseFlags(flagPart)

			sort.Slice(appendFlags, func(i, j int) bool {
				return appendFlags[i] < appendFlags[j]
			})

			hs.twoStageAffix = true
		}

		condition := "."
		if len(parts) > 4 {
			condition = parts[4]
		}
		// at least the gascon affix file has this issue
		if strings.HasPrefix(condition, "[") && !strings.Contains(condition, "]") {
			condition = condition + "]"
		}
		// "dash hasn't got special meaning" (we must escape it)
		if strings.Contains(condition, "-") {
			condition = escapeDash(condition)
		}

		var regex string
		if condition == "." || condition == strip {
			regex = ".*" // Zero condition is indicated by dot
		} else {
			regex = fmt.Sprintf(conditionPattern, condition)
		}

		patternIndex, ok := hs.seenPatterns[regex]
		if !ok {
			patternIndex = len(hs.patterns)
			hs.seenPatterns[regex] = patternIndex

			matcher, err := regexp.Compile(regex)
			if err != nil {
				return fmt.Errorf("Unable to compile %s", regex)
			}
			hs.patterns = append(hs.patterns, matcher)
		}

		stripOrd, ok := hs.seenStrips[strip]
		if !ok {
			stripOrd = len(hs.seenStrips)
			hs.seenStrips[strip] = stripOrd
		}

		if strip == "t" && !ok {
			fmt.Print("")
		}

		scratch := encodeFlags(appendFlags)
		appendFlagsOrd := hs.flagLookup_add(scratch)

		hs.affixData = append(hs.affixData, affixDataEntry{
			flag:         flag,
			stripOrd:     stripOrd,
			patternOrd:   patternIndex,
			crossProduct: crossProduct,
			append:       rune(appendFlagsOrd)})

		if isSuffix {
			affixArg = reverse(affixArg)
		}

		if _, ok := affixes[affixArg]; !ok {
			affixes[affixArg] = []int{}
		}
		affixes[affixArg] = append(affixes[affixArg], hs.currentAffix)

		if _, ok := hs.flagToaffixes[flag]; !ok {
			hs.flagToaffixes[flag] = []int{}
		}
		hs.flagToaffixes[flag] = append(hs.flagToaffixes[flag], hs.currentAffix)

		hs.currentAffix++
	}

	return nil
}
