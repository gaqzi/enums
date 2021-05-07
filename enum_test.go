package enums_test

import (
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

func TestDiff(t *testing.T) {
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
