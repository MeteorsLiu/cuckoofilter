package cuckoo_test

import (
	"fmt"

	cuckoo "github.com/panmari/cuckoofilter"
)

func Example() {
	cf := cuckoo.NewFilter(1000)

	cf.Insert([]byte("pizza"))
	cf.Insert([]byte("tacos"))
	cf.Insert([]byte("tacos")) // Re-insertion is possible.

	fmt.Println(cf.Lookup([]byte("pizza")))
	fmt.Println(cf.Lookup([]byte("missing")))

	cf.Reset()
	fmt.Println(cf.Lookup([]byte("pizza")))
	// Output:
	// true
	// false
	// false
}

func ExampleLookup() {
	cf := cuckoo.NewFilter(1000)

	cf.Insert([]byte("pizza"))
	cf.Insert([]byte("tacos"))

	fmt.Println(cf.Lookup([]byte("pizza")))
	fmt.Println(cf.Lookup([]byte("missing")))
	// Output:
	// true
	// false
}

func ExampleDelete() {
	cf := cuckoo.NewFilter(1000)

	cf.Insert([]byte("pizza"))
	cf.Insert([]byte("tacos"))

	fmt.Println(cf.Lookup([]byte("pizza")))

	cf.Delete([]byte("pizza"))
	fmt.Println(cf.Lookup([]byte("pizza")))
	// Output:
	// true
	// false
}
