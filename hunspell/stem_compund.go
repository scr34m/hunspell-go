package hunspell

import (
	"fmt"
	"sort"
)

func (hs *hunSpell) compoundStem(r []rune, length int) []string {
	stems := []string{}

	if debug {
		fmt.Println(string(r))
	}
	buffer := hs.compoundStemSplit(r, length, false)
	if len(buffer) != 0 {
		temp := ""
		for _, word := range buffer {
			temp = temp + string(word)
		}

		// maradék
		temp_s := len([]rune(temp))
		residual := r[temp_s:]
		if len(residual) != 0 {
			buffer = append(buffer, string(residual))
		}

		elements := len(buffer)
		idx := 0
		for idx < elements {
			temp = ""
			for j := idx; j < elements; j++ {
				temp = temp + buffer[j]
			}
			if debug {
				fmt.Println("temp: " + temp)
			}

			residual_stem := hs.compoundCheck(temp)
			if debug {
				fmt.Println("residual_stem: " + residual_stem)
			}
			if residual_stem != "" {
				for j := elements - 1; j >= idx; j-- {
					buffer = buffer[:len(buffer)-1]
				}
				buffer = append(buffer, residual_stem)
				elements = len(buffer)
			}
			idx++
		}

		temp = ""
		for i := 0; i < len(buffer); i++ {
			temp = temp + buffer[i]
		}

		stems = append(stems, temp)
	} else {
		word := string(r)
		temp := hs.compoundCheck(word)
		if temp != "" {
			stems = append(stems, temp)
		} else {
			stems = append(stems, word)
		}
	}

	return stems
}

func (hs *hunSpell) compoundCheck(residual string) string {
	r := []rune(residual)
	residual_stems := hs.uniqueStems(r, len(r))

	// van szótő a maradékon
	if len(residual_stems) != 0 {
		if len(residual_stems) > 1 {
			sort.Slice(residual_stems, func(i, j int) bool {
				return len([]rune(residual_stems[i])) > len([]rune(residual_stems[j]))
			})
		}

		term := residual_stems[0]
		if stem, ok := hs.overrideMap[term]; ok {
			return stem
		}
		return string(term)
	}
	return ""
}

func (hs *hunSpell) compoundStemSplit(word []rune, length int, caseVariant bool) []string {
	stems := []string{}
	stripStart := 0
	stripEnd := 2

	// log.Println(string(word))
	if length > 2 {
		for i := 2; i <= length; i++ {
			strip := word[stripStart : stripStart+(stripEnd-stripStart)]
			// log.Println(string(strip) + ":")

			// look backward for stems + next strip
			if len(stems) != 0 {
				temp := strip
				idx := len(stems)
				found := false
				for idx > 0 {
					idx--
					temp = append([]rune(stems[idx]), temp...)
					// log.Println("temp: " + string(temp))
					l := hs.lookupWord(temp, 0, len(temp))
					if l != nil {
						found = true
						for j := len(stems) - 1; j >= idx; j-- {
							stems = stems[:len(stems)-1]
						}
						stems = append(stems, l.word)
						// log.Println("temp lookup: " + l.word)
						break
					}
				}
				if found {
					stripStart = stripEnd
					stripEnd++
					continue
				}
			}

			if stripEnd-stripStart > 1 && len(strip) > 2 {
				l := hs.lookupWord(strip, 0, stripEnd-stripStart)
				if l == nil {
					// log.Println("lookup: -")
				} else {
					stems = append(stems, l.word)
					// log.Println("lookup: " + l.word)
					stripStart = stripEnd
				}
			}
			stripEnd += 1
		}
	}
	return stems
}
