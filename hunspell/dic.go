package hunspell

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func (hs *hunSpell) readDictionaryFile(file io.Reader) error {

	scanner := bufio.NewScanner(file)

	// get first line
	if !scanner.Scan() {
		return scanner.Err()
	}

	line := scanner.Text()
	_, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return err
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || line[0] == '/' || line[0] == '#' || line[0] == '\t' {
			continue
		}

		idx := strings.Index(line, "/")

		if idx == 0 || idx+1 == len(line) {
			return fmt.Errorf("Slash char found in first or last position")
		}

		tab := strings.IndexRune(line, '\t')
		if tab != -1 {
			line = line[:tab]
		}

		var flags []rune
		var word string
		if idx == -1 {
			word = line
		} else {
			if hs.aliasCount > 0 {
				flagPartInt, err := strconv.Atoi(line[idx+1:])
				if err != nil {
					return err
				}
				flags = []rune(hs.getAliasValue(flagPartInt))
			} else {
				flags = []rune(line[idx+1:])
			}
			word = line[:idx]
		}

		hs.dict[word] = dictEntry{word: word, flags: flags}
	}

	/*
	   TODO sort by word + checks below

	   		if (hasStemExceptions == false) {
	               int morphStart = line.indexOf(MORPH_SEPARATOR);
	               if (morphStart >= 0 && morphStart < line.length()) {
	                 hasStemExceptions = parseStemException(line.substring(morphStart+1)) != null;
	               }
	             }

	           // we possibly have morphological data
	           int stemExceptionID = 0;
	           if (hasStemExceptions && end+1 < line.length()) {
	             String stemException = parseStemException(line.substring(end+1));
	             if (stemException != null) {
	               if (stemExceptionCount == stemExceptions.length) {
	                 int newSize = ArrayUtil.oversize(stemExceptionCount+1, RamUsageEstimator.NUM_BYTES_OBJECT_REF);
	                 stemExceptions = Arrays.copyOf(stemExceptions, newSize);
	               }
	               stemExceptionID = stemExceptionCount+1; // we use '0' to indicate no exception for the form
	               stemExceptions[stemExceptionCount++] = stemException;
	             }
	           }
	*/

	if hs.hasStemExceptions {
		hs.formStep = 2
	} else {
		hs.formStep = 1
	}

	return nil
}

func (hs *hunSpell) lookupWord(word []rune, offset int, length int) *dictEntry {
	search := string(word[offset:length])
	if v, ok := hs.dict[string(search)]; ok {
		return &v
	} else {
		return nil
	}
}
