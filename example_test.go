package enums_test

import (
	"fmt"
	"testing"

	"github.com/gaqzi/enums"
	"github.com/gaqzi/enums/enumstest"
	"github.com/gaqzi/enums/testdata/full"
)

func Example() {
	// The simplest way of checking if all values are covered:
	var t *testing.T // so this file can compile
	enumstest.NoDiff(t, "./testdata/full", "full.Flag", full.AllFlags())

	// To control the process more you can use enums.All and Collection.Diff.
	// This example is equivalent to the one-liner above:
	collection, err := enums.All("./testdata/full", "full.Flag")
	if err != nil {
		panic("unexpected error in example" + err.Error())
	}

	if diff := collection.Diff(full.AllFlags()); !diff.Zero() {
		fmt.Println("expected to have all flags: " + diff.String())
	}

	// There are sometimes cases where you don't want to warn about certain values,
	// but you still want to be told when new ones are created so that you can make
	// a decision for what to do. So then you can modify the Diff result to suit
	// your needs.
	ignoreValues := []full.Flag{full.DeployOneThing}
	var kept []enums.Enum

	for _, e := range collection.Enums {
		var skip bool
		for _, v := range ignoreValues {
			if e.Value == fmt.Sprintf("%#v", v) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		kept = append(kept, e)
	}
	collection.Enums = kept

	if diff := collection.Diff(full.MissingFlags()); !diff.Zero() {
		fmt.Println("expected to have all flags: " + diff.String())
	}
}
