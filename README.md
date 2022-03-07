Solr's hunspell stemming implementation

Primary goal was to implement the stemming mechanism works in Solr.

```go
package main

import (
	"log"
	"os"

	"github.com/scr34m/hunspell-go/hunspell"
)

func main() {
	aff, err := os.Open("hu_HU.aff")
	if err != nil {
		log.Fatalf("Unable top open: %s", err)
	}

	dic, err := os.Open("hu_HU.dic")
	if err != nil {
		log.Fatalf("Unable top open: %s", err)
	}

	hs, err := hunspell.NewHunSpellReader(aff, dic)
	if err != nil {
		log.Fatalf("HunSpell error: %s", err)
	}

	log.Println(hs.Stem("barackos"))
}
```