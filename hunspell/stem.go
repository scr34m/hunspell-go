package hunspell

import (
	"fmt"
	"log"
)

var debug = false

/*
 * Generates a list of stems for the provided word
 *
 * @param word Word to generate the stems for
 * @param previous previous affix that was removed (so we dont remove same one twice)
 * @param prevFlag Flag from a previous stemming step that need to be cross-checked with any affixes in this recursive step
 * @param prefixFlag flag of the most inner removed prefix, so that when removing a suffix, it's also checked against the word
 * @param recursionDepth current recursiondepth
 * @param doPrefix true if we should remove prefixes
 * @param doSuffix true if we should remove suffixes
 * @param previousWasPrefix true if the previous removal was a prefix:
 *        if we are removing a suffix, and it has no continuation requirements, it's ok.
 *        but two prefixes (COMPLEXPREFIXES) or two suffixes must have continuation requirements to recurse.
 * @param circumfix true if the previous prefix removal was signed as a circumfix
 *        this means inner most suffix must also contain circumfix flag.
 * @param caseVariant true if we are searching for a case variant. if the word has KEEPCASE flag it cannot succeed.
 * @return List of stems, or empty list if no stems are found
 */
func (hs *hunSpell) stem(word []rune, length int, previous int, prevFlag rune, prefixFlag rune, recursionDepth int, doPrefix bool, doSuffix bool, previousWasPrefix bool, circumfix bool, caseVariant bool) []string {
	items := []string{}

	if doPrefix && len(hs.prefixes) > 0 {
		/*
			if (doPrefix && dictionary.prefixes != null) {
			  FST<IntsRef> fst = dictionary.prefixes;
			  Outputs<IntsRef> outputs = fst.outputs;
			  FST.BytesReader bytesReader = prefixReaders[recursionDepth];
			  FST.Arc<IntsRef> arc = prefixArcs[recursionDepth];
			  fst.getFirstArc(arc);
			  IntsRef NO_OUTPUT = outputs.getNoOutput();
			  IntsRef output = NO_OUTPUT;
			  int limit = dictionary.fullStrip ? length : length-1;
			  for (int i = 0; i < limit; i++) {
				if (i > 0) {
				  int ch = word[i-1];
				  if (fst.findTargetArc(ch, arc, arc, bytesReader) == null) {
					break;
				  } else if (arc.output != NO_OUTPUT) {
					output = fst.outputs.add(output, arc.output);
				  }
				}
				IntsRef prefixes = null;
				if (!arc.isFinal()) {
				  continue;
				} else {
				  prefixes = fst.outputs.add(output, arc.nextFinalOutput);
				}

				for (int j = 0; j < prefixes.length; j++) {
				  int prefix = prefixes.ints[prefixes.offset + j];
				  if (prefix == previous) {
					continue;
				  }
				  affixReader.setPosition(8 * prefix);
				  char flag = (char) (affixReader.readShort() & 0xffff);
				  char stripOrd = (char) (affixReader.readShort() & 0xffff);
				  int condition = (char) (affixReader.readShort() & 0xffff);
				  boolean crossProduct = (condition & 1) == 1;
				  condition >>>= 1;
				  char append = (char) (affixReader.readShort() & 0xffff);

				  final boolean compatible;
				  if (recursionDepth == 0) {
					if (dictionary.onlyincompound == -1) {
					  compatible = true;
					} else {
					  // check if affix is allowed in a non-compound word
					  dictionary.flagLookup.get(append, scratch);
					  char appendFlags[] = Dictionary.decodeFlags(scratch);
					  compatible = !Dictionary.hasFlag(appendFlags, (char) dictionary.onlyincompound);
					}
				  } else if (crossProduct) {
					// cross check incoming continuation class (flag of previous affix) against list.
					dictionary.flagLookup.get(append, scratch);
					char appendFlags[] = Dictionary.decodeFlags(scratch);
					assert prevFlag >= 0;
					boolean allowed = dictionary.onlyincompound == -1 ||
									  !Dictionary.hasFlag(appendFlags, (char) dictionary.onlyincompound);
					compatible = allowed && hasCrossCheckedFlag((char)prevFlag, appendFlags, false);
				  } else {
					compatible = false;
				  }

				  if (compatible) {
					int deAffixedStart = i;
					int deAffixedLength = length - deAffixedStart;

					int stripStart = dictionary.stripOffsets[stripOrd];
					int stripEnd = dictionary.stripOffsets[stripOrd+1];
					int stripLength = stripEnd - stripStart;

					if (!checkCondition(condition, dictionary.stripData, stripStart, stripLength, word, deAffixedStart, deAffixedLength)) {
					  continue;
					}

					char strippedWord[] = new char[stripLength + deAffixedLength];
					System.arraycopy(dictionary.stripData, stripStart, strippedWord, 0, stripLength);
					System.arraycopy(word, deAffixedStart, strippedWord, stripLength, deAffixedLength);

					List<CharsRef> stemList = applyAffix(strippedWord, strippedWord.length, prefix, -1, recursionDepth, true, circumfix, caseVariant);

					stems.addAll(stemList);
				  }
				}
			  }
			}
		*/
	}

	if doSuffix && len(hs.suffixes) > 0 {
		limit := 1
		if hs.fullStrip {
			limit = 0
		}

		lookup := ""
		for i := length; i >= limit; i-- {
			var suffixes []int
			if i < length {
				lookup = lookup + string(word[i])
				if debug && recursionDepth == 0 {
					fmt.Println(lookup)
				}
				list, ok := hs.suffixes[lookup]
				if !ok {
					break
				}
				suffixes = list
			}

			if suffixes == nil {
				continue
			}

			for j := 0; j < len(suffixes); j++ {

				suffix := suffixes[j]
				if suffix == previous {
					continue
				}

				affix := hs.affixData[suffix]

				stripOrd := affix.stripOrd
				condition := affix.patternOrd
				crossProduct := affix.crossProduct
				appendAff := affix.append

				var compatible bool

				if recursionDepth == 0 {
					if hs.onlyincompound == -1 {
						compatible = true
					} else {
						// check if affix is allowed in a non-compound word
						appendFlags := decodeFlags(hs.flags[appendAff])
						compatible = !hasFlag(appendFlags, hs.onlyincompound)
					}
				} else if crossProduct {
					// cross check incoming continuation class (flag of previous affix) against list.
					appendFlags := decodeFlags(hs.flags[appendAff])
					allowed := hs.onlyincompound == -1 || !hasFlag(appendFlags, hs.onlyincompound)
					compatible = allowed && hasCrossCheckedFlag(prevFlag, appendFlags, previousWasPrefix)
				} else {
					compatible = false
				}

				if compatible {
					appendLength := length - i
					deAffixedLength := length - appendLength

					stripStart := hs.stripOffsets[stripOrd]
					stripEnd := hs.stripOffsets[stripOrd+1]
					stripLength := stripEnd - stripStart

					if debug {
						if suffix == 18153 {
							fmt.Print()
						}
						fmt.Printf("stem: depth%dsuffix: %d", recursionDepth, suffix)
					}
					if !hs.checkCondition(condition, word, 0, deAffixedLength, hs.stripData, stripStart, stripLength) {
						if debug {
							fmt.Println(" FAIL")
						}
						continue
					}
					if debug {
						fmt.Println(" OK")
					}

					strippedWord := make([]rune, stripLength+deAffixedLength)

					for i := 0; i < deAffixedLength; i++ {
						strippedWord[i] = word[i]
					}

					for i := 0; i < stripLength; i++ {
						strippedWord[deAffixedLength+i] = hs.stripData[stripStart+i]
					}

					items2 := hs.applyAffix(strippedWord, len(strippedWord), suffix, prefixFlag, recursionDepth, false, circumfix, caseVariant)
					items = append(items, items2...)
				}

			}
		}
	}

	if debug && len(items) > 0 {
		fmt.Println("match in stem")
	}
	return items
}

/**
 * Applies the affix rule to the given word, producing a list of stems if any are found
 *
 * @param strippedWord Word the affix has been removed and the strip added
 * @param length valid length of stripped word
 * @param affix HunspellAffix representing the affix rule itself
 * @param prefixFlag when we already stripped a prefix, we cant simply recurse and check the suffix, unless both are compatible
 *                   so we must check dictionary form against both to add it as a stem!
 * @param recursionDepth current recursion depth
 * @param prefix true if we are removing a prefix (false if it's a suffix)
 * @return List of stems for the word, or an empty list if none are found
 */
func (hs *hunSpell) applyAffix(strippedWord []rune, length int, affix int, prefixFlag rune, recursionDepth int, prefix bool, circumfix bool, caseVariant bool) []string {

	affixData := hs.affixData[affix]

	flag := affixData.flag
	crossProduct := affixData.crossProduct
	appendAff := affixData.append

	items := []string{}

	form := hs.lookupWord(strippedWord, 0, length)
	if form != nil {
		// XXX we only have one forms always?!
		// XXX, for i := 0; i < len(forms); i += hs.formStep {
		for i := 0; i < 1; i++ {
			wordFlags := form.flags
			if hasFlag(wordFlags, flag) {
				// confusing: in this one exception, we already chained the first prefix against the second,
				// so it doesnt need to be checked against the word
				chainedPrefix := hs.complexPrefixes && recursionDepth == 1 && prefix
				if !chainedPrefix && prefixFlag >= 0 && !hasFlag(wordFlags, prefixFlag) {
					// see if we can chain prefix thru the suffix continuation class (only if it has any!)
					scratch := hs.flags[appendAff]
					appendFlags := decodeFlags(scratch)
					if !hasCrossCheckedFlag(prefixFlag, appendFlags, false) {
						continue
					}
				}

				// if circumfix was previously set by a prefix, we must check this suffix,
				// to ensure it has it, and vice versa
				if hs.circumfix != -1 {
					scratch := hs.flags[appendAff]
					appendFlags := decodeFlags(scratch)
					suffixCircumfix := hasFlag(appendFlags, hs.circumfix)
					if circumfix != suffixCircumfix {
						continue
					}
				}

				// we are looking for a case variant, but this word does not allow it
				if caseVariant && hs.keepcase != -1 && hasFlag(wordFlags, hs.keepcase) {
					continue
				}

				// we aren't decompounding (yet)
				if hs.onlyincompound != -1 && hasFlag(wordFlags, hs.onlyincompound) {
					continue
				}

				items = append(items, hs.newStem(strippedWord, length))
			}
		}
	}

	// if a circumfix flag is defined in the dictionary, and we are a prefix, we need to check if we have that flag
	if hs.circumfix != -1 && !circumfix && prefix {
		scratch := hs.flags[appendAff]
		appendFlags := decodeFlags(scratch)
		circumfix = hasFlag(appendFlags, hs.circumfix)
	}

	if crossProduct {
		if recursionDepth == 0 {
			if prefix {
				// we took away the first prefix.
				// COMPLEXPREFIXES = true:  combine with a second prefix and another suffix
				// COMPLEXPREFIXES = false: combine with a suffix
				recursionDepth += 1
				items2 := hs.stem(strippedWord, length, affix, flag, flag, recursionDepth, hs.complexPrefixes && hs.twoStageAffix, true, true, circumfix, caseVariant)
				items = append(items, items2...)
			} else if !hs.complexPrefixes && hs.twoStageAffix {
				// we took away a suffix.
				// COMPLEXPREFIXES = true: we don't recurse! only one suffix allowed
				// COMPLEXPREFIXES = false: combine with another suffix
				recursionDepth += 1
				items2 := hs.stem(strippedWord, length, affix, flag, prefixFlag, recursionDepth, false, true, false, circumfix, caseVariant)
				items = append(items, items2...)
			}
		} else if recursionDepth == 1 {
			if prefix && hs.complexPrefixes {
				// we took away the second prefix: go look for another suffix
				recursionDepth += 1
				items2 := hs.stem(strippedWord, length, affix, flag, flag, recursionDepth, false, true, true, circumfix, caseVariant)
				items = append(items, items2...)
			} else if !prefix && !hs.complexPrefixes && hs.twoStageAffix {
				// we took away a prefix, then a suffix: go look for another suffix
				recursionDepth += 1
				items2 := hs.stem(strippedWord, length, affix, flag, prefixFlag, recursionDepth, false, true, false, circumfix, caseVariant)
				items = append(items, items2...)
			}
		}
	}

	if debug && len(items) > 0 {
		fmt.Println("match in applyAffix")
	}
	return items
}

/** checks condition of the concatenation of two strings */
// note: this is pretty stupid, we really should subtract strip from the condition up front and just check the stem
// but this is a little bit more complicated.
func (hs *hunSpell) checkCondition(condition int, c1 []rune, c1off int, c1len int, c2 []rune, c2off int, c2len int) bool {
	if condition != 0 {
		pattern := hs.patterns[condition]
		var r []rune
		r = append(r, c1[c1off:c1off+c1len]...)
		r = append(r, c2[c2off:c2off+c2len]...)
		return pattern.MatchString(string(r))
	}
	return true
}

// XXX not used parameters are: IntsRef forms, int formID
func (hs *hunSpell) newStem(buffer []rune, length int) string {
	var exception string
	if hs.hasStemExceptions {
		log.Panic("TODO newStem() hasStemExceptions")
		/*
		   int exceptionID = forms.ints[forms.offset + formID + 1];
		   if (exceptionID > 0) {
		     exception = dictionary.getStemException(exceptionID);
		   } else {
		     exception = null;
		   }
		*/
	}

	if hs.needsOutputCleaning {
		log.Panic("TODO newStem() needsOutputCleaning")
		/*
		   scratchSegment.setLength(0);
		   if (exception != null) {
		     scratchSegment.append(exception);
		   } else {
		     scratchSegment.append(buffer, 0, length);
		   }
		   try {
		     Dictionary.applyMappings(dictionary.oconv, scratchSegment);
		   } catch (IOException bogus) {
		     throw new RuntimeException(bogus);
		   }
		   char cleaned[] = new char[scratchSegment.length()];
		   scratchSegment.getChars(0, cleaned.length, cleaned, 0);
		   return new CharsRef(cleaned, 0, cleaned.length);
		*/
	}

	if exception != "" {
		return exception
	} else {
		return string(buffer)
	}
}
