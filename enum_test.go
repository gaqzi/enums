package enums_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gaqzi/enums"
)

func TestAll(t *testing.T) {
	t.Run("returns an empty collection when no matches found", func(t *testing.T) {
		matches, err := enums.All("./testdata/nomatch", "Flag")
		require.NoError(t, err, "error when scanning testdata/nomatches")

		require.Empty(t, matches, "expected to not have found any matches")
	})

	t.Run("when one match found return it", func(t *testing.T) {
		matches, err := enums.All("./testdata/singlematch", "singlematch.Flag")
		require.NoError(t, err, "error when scanning testdata/singlematch")

		require.Equal(
			t,
			enums.Collection{
				Type: "github.com/gaqzi/enums/testdata/singlematch.Flag",
				Enums: []enums.Enum{
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
		matches, err := enums.All("./testdata/multimatch", "multimatch.Flag")
		require.NoError(t, err, "error when scanning testdata/multimatch")

		require.Equal(
			t,
			enums.Collection{
				Type: "github.com/gaqzi/enums/testdata/multimatch.Flag",
				Enums: []enums.Enum{
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
		// Fixes https://github.com/gaqzi/enums/issues/38.
		matches, err := enums.All("./testdata/falsepositiveinfunc", "Flag")
		require.NoError(t, err, "error when scanning testdata/falsepositiveinfunc")

		require.Equal(
			t,
			enums.Collection{
				Type: "github.com/gaqzi/enums/testdata/falsepositiveinfunc.Flag",
				Enums: []enums.Enum{
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
			enums.Diff{Missing: enums.Collection{Type: "enums_test.val"}},
			enums.Collection{
				Type: "enums_test.val",
				Enums: []enums.Enum{
					{"test", `"hello"`},
				},
			}.Diff([]val{test}),
			"expected the same values to have no diff",
		)
	})

	t.Run("missing values are shown", func(t *testing.T) {
		require.Equal(
			t,
			enums.Diff{
				Missing: enums.Collection{
					Type: "enums_test.val",
					Enums: []enums.Enum{
						{"test", `"hello"`},
					},
				},
			},
			enums.Collection{
				Type: "enums_test.val",
				Enums: []enums.Enum{
					{"test", `"hello"`},
				},
			}.Diff([]val{}),
			"expected a diff message",
		)
	})

	t.Run("extra values are shown", func(t *testing.T) {
		require.Equal(
			t,
			enums.Diff{Extra: []string{`"m000"`}},
			enums.Collection{}.Diff([]val{"m000"}),
			"expected a diff message",
		)
	})

	t.Run("handles structs", func(t *testing.T) {
		type testStruct struct {
			FieldA string `enums:"identifier"`
		}
		test := testStruct{FieldA: "Hello"}
		collection := enums.Collection{
			Type:      "enums_test.testStruct",
			FieldName: "FieldA",
			Enums: []enums.Enum{
				{"test", `"Hello"`},
			},
		}

		t.Run("uses the first field that has `enum:\"identifier\"` as a tag", func(t *testing.T) {
			require.Equal(
				t,
				enums.Diff{
					Missing: enums.Collection{
						Type:      "enums_test.testStruct",
						FieldName: "FieldA",
					},
				},
				collection.Diff([]testStruct{test}),
				"expected no differences",
			)
		})

		t.Run("doesn't match fields for other struct types", func(t *testing.T) {
			type otherStruct struct {
				FieldA string `enums:"identifier"`
			}
			test := otherStruct{FieldA: "Hello"}

			require.Equal(
				t,
				enums.Diff{
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
		typ := reflect.TypeOf(enums.Diff{})
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
		diff     enums.Diff
		expected string
		isZero   bool
	}{
		{
			name:     "Diff is zero",
			diff:     enums.Diff{},
			expected: "<Diff{}>",
			isZero:   true,
		},
		{
			name: "Missing is set",
			diff: enums.Diff{Missing: enums.Collection{
				Type: "full.Flag",
				Enums: []enums.Enum{
					{
						Name:  "FlagSomething",
						Value: "flag-something",
					},
				},
			}},
			expected: "Enums declared but not part of actual:\n" +
				"\tFlagSomething = flag-something\n",
		},
		{
			name: "Extra is set",
			diff: enums.Diff{Extra: []string{"hello"}},
			expected: "Extra values provided but not part of Enums:\n" +
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
