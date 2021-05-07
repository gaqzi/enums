package enums_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gaqzi/enums"
)

func TestAll(t *testing.T) {
	t.Run("returns an empty list when no matches found", func(t *testing.T) {
		matches, err := enums.All("./testdata/nomatch", "Flag")
		require.NoError(t, err, "error when scanning testdata/nomatches")

		require.Empty(t, matches, "expected to not have found any matches")
	})

	t.Run("when one match found return it", func(t *testing.T) {
		matches, err := enums.All("./testdata/singlematch", "singlematch.Flag")
		require.NoError(t, err, "error when scanning testdata/singlematch")

		require.ElementsMatch(
			t,
			enums.Collection{
				{
					Type:  "github.com/gaqzi/enums/testdata/singlematch.Flag",
					Name:  "FlagSomethingCouldBe",
					Value: `"flag-whatever"`,
				},
			},
			matches,
			"expected to have gotten back a single match",
		)
	})

	t.Run("returns all matches found on a single line", func(t *testing.T) {
		matches, err := enums.All("./testdata/multimatch", "multimatch.Flag")
		require.NoError(t, err, "error when scanning testdata/multimatch")

		require.ElementsMatch(
			t,
			enums.Collection{
				{
					Type:  "github.com/gaqzi/enums/testdata/multimatch.Flag",
					Name:  "FlagSomethingCouldBe",
					Value: `"flag-whatever"`,
				},
				{
					Type:  "github.com/gaqzi/enums/testdata/multimatch.Flag",
					Name:  "FlagSomethingElse",
					Value: `"flag-whomever"`,
				},
			},
			matches,
			"expected to have gotten back a single match",
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
			enums.Diff{},
			enums.Collection{
				{
					Type:  "enums_test.val",
					Name:  "test",
					Value: `"hello"`,
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
					{
						Type:  "enums_test.val",
						Name:  "test",
						Value: `"hello"`,
					},
				},
			},
			enums.Collection{
				{
					Type:  "enums_test.val",
					Name:  "test",
					Value: `"hello"`,
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
			diff: enums.Diff{Missing: []enums.Enum{
				{
					Type:  "full.Flag",
					Name:  "FlagSomething",
					Value: "flag-something",
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
