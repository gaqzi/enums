package typedecl_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gaqzi/typedecl"
)

func TestAll(t *testing.T) {
	t.Run("returns an empty collection when no matches found", func(t *testing.T) {
		matches, err := typedecl.All("./testdata/nomatch", "Flag")
		require.NoError(t, err, "error when scanning testdata/nomatches")

		require.Empty(t, matches, "expected to not have found any matches")
	})

	t.Run("when one match found return it", func(t *testing.T) {
		matches, err := typedecl.All("./testdata/singlematch", "singlematch.Flag")
		require.NoError(t, err, "error when scanning testdata/singlematch")

		require.Equal(
			t,
			typedecl.Collection{
				Type: "github.com/gaqzi/typedecl/testdata/singlematch.Flag",
				Matches: []typedecl.Match{
					{
						Name:  "FlagSomethingCouldBe",
						Value: `"flag-whatever"`,
					},
				},
			},
			matches,
			"expected to have gotten back a single match",
		)
	})

	t.Run("returns all matches found", func(t *testing.T) {
		matches, err := typedecl.All("./testdata/multimatch", "multimatch.Flag")
		require.NoError(t, err, "error when scanning testdata/multimatch")

		require.Equal(
			t,
			typedecl.Collection{
				Type: "github.com/gaqzi/typedecl/testdata/multimatch.Flag",
				Matches: []typedecl.Match{
					{
						Name:  "FlagSomethingCouldBe",
						Value: `"flag-whatever"`,
					},
					{
						Name:  "FlagSomethingElse",
						Value: `"flag-whomever"`,
					},
				},
			},
			matches,
			"expected to have gotten back a single match",
		)
	})

	t.Run("does not match a variable declared in a func for the type", func(t *testing.T) {
		// Fixes https://github.com/gaqzi/typedecl/issues/38.
		matches, err := typedecl.All("./testdata/falsepositiveinfunc", "Flag")
		require.NoError(t, err, "error when scanning testdata/falsepositiveinfunc")

		require.Equal(
			t,
			typedecl.Collection{
				Type: "github.com/gaqzi/typedecl/testdata/falsepositiveinfunc.Flag",
				Matches: []typedecl.Match{
					{
						Name:  "AnotherExample",
						Value: `"hello-there"`,
					},
					{
						Name:  "DeployAllTheThings",
						Value: `"deploy-all-the-things"`,
					},
				},
			},
			matches,
			"expected to only have found the one declared flag and not a variable of the type declared anywhere else",
		)
	})
}

func TestCollection_Diff(t *testing.T) {
	type val string

	const (
		test = "hello"
	)

	t.Run("Same values returns an empty string", func(t *testing.T) {
		require.Equal(
			t,
			typedecl.Diff{Missing: typedecl.Collection{Type: "typedecl_test.val"}},
			typedecl.Collection{
				Type: "typedecl_test.val",
				Matches: []typedecl.Match{
					{"test", `"hello"`},
				},
			}.Diff([]val{test}),
			"expected the same values to have no diff",
		)
	})

	t.Run("missing values are shown", func(t *testing.T) {
		require.Equal(
			t,
			typedecl.Diff{
				Missing: typedecl.Collection{
					Type: "typedecl_test.val",
					Matches: []typedecl.Match{
						{"test", `"hello"`},
					},
				},
			},
			typedecl.Collection{
				Type: "typedecl_test.val",
				Matches: []typedecl.Match{
					{"test", `"hello"`},
				},
			}.Diff([]val{}),
			"expected a diff message",
		)
	})

	t.Run("extra values are shown", func(t *testing.T) {
		require.Equal(
			t,
			typedecl.Diff{Extra: []string{`"m000"`}},
			typedecl.Collection{}.Diff([]val{"m000"}),
			"expected a diff message",
		)
	})

	t.Run("handles structs", func(t *testing.T) {
		type testStruct struct {
			FieldA string `typedecl:"identifier"`
		}
		test := testStruct{FieldA: "Hello"}
		collection := typedecl.Collection{
			Type:      "typedecl_test.testStruct",
			FieldName: "FieldA",
			Matches: []typedecl.Match{
				{"test", `"Hello"`},
			},
		}

		t.Run("uses the first field that has `typedecl:\"identifier\"` as a tag", func(t *testing.T) {
			require.Equal(
				t,
				typedecl.Diff{
					Missing: typedecl.Collection{
						Type:      "typedecl_test.testStruct",
						FieldName: "FieldA",
					},
				},
				collection.Diff([]testStruct{test}),
				"expected no differences",
			)
		})

		t.Run("doesn't match fields for other struct types", func(t *testing.T) {
			type otherStruct struct {
				FieldA string `typedecl:"identifier"`
			}
			test := otherStruct{FieldA: "Hello"}

			require.Equal(
				t,
				typedecl.Diff{
					Extra:   []string{fmt.Sprintf("%#v", test)},
					Missing: collection,
				},
				collection.Diff([]otherStruct{test}),
				"expected to have a different value as incorrect type",
			)
		})
	})
}

func TestDiff(t *testing.T) {
	t.Run("Handles all members of the struct", func(t *testing.T) {
		typ := reflect.TypeOf(typedecl.Diff{})
		var allFields []string

		for i := 0; i < typ.NumField(); i++ {
			allFields = append(allFields, typ.Field(i).Name)
		}

		require.ElementsMatch(
			t,
			[]string{"Missing", "Extra"}, // All handled fields
			allFields,
			"when a need field is added to Diff remember to update the test cases below to handle them",
		)
	})

	testCases := []struct {
		name     string
		diff     typedecl.Diff
		expected string
		isZero   bool
	}{
		{
			name:     "Diff is zero",
			diff:     typedecl.Diff{},
			expected: "<Diff{}>",
			isZero:   true,
		},
		{
			name: "Missing is set",
			diff: typedecl.Diff{Missing: typedecl.Collection{
				Type: "full.Flag",
				Matches: []typedecl.Match{
					{
						Name:  "FlagSomething",
						Value: "flag-something",
					},
				},
			}},
			expected: "Matches declared but not part of actual:\n" +
				"\tFlagSomething = flag-something\n",
		},
		{
			name: "Extra is set",
			diff: typedecl.Diff{Extra: []string{"hello"}},
			expected: "Extra values provided but not part of Matches:\n" +
				"\thello\n",
		},
	}

	for _, tc := range testCases {
		t.Run("#String: "+tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.diff.String())
		})

		t.Run("#Zero: "+tc.name, func(t *testing.T) {
			require.Equal(t, tc.isZero, tc.diff.Zero())
		})
	}
}
