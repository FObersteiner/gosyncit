package tree_test

import (
	"log"
	"testing"

	"golang.org/x/exp/maps"
	"slices"

	"gosyncit/lib/tree"
)

func TestTest(t *testing.T) {
	s := tree.Treetest()
	log.Println(s)

	m := make(map[string]string)
	// fill the map
	m["hello"] = "mars"
	m["world"] = "goodbye"
	keys := maps.Keys(m)
	slices.Sort(keys)
	for _, k := range keys {
		log.Println(k, m[k])
	}
}
