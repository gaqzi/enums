package enums_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gaqzi/enums"
	"github.com/gaqzi/enums/testdata/full"
)

func TestIntegration(t *testing.T) {
	collection, err := enums.All("./testdata/full", "full.Flag")
	require.NoError(t, err)

	t.Run("No error when no difference", func(t *testing.T) {
		diff := collection.Diff(full.AllFlags())
		require.True(
			t,
			diff.Zero(),
			"expected no differences: %s", diff,
		)
	})

	t.Run("Indicates a difference when one is set", func(t *testing.T) {
		require.False(
			t,
			collection.Diff(full.MissingFlags()).Zero(),
			"expected to have indicated a diff",
		)
	})

	t.Run("Handles structs as enum type", func(t *testing.T) {
		fsCollection, err := enums.All("./testdata/full", "full.FlagStruct")
		require.NoError(t, err)

		t.Run("when all flags are present", func(t *testing.T) {
			diff := fsCollection.Diff(full.AllFlagStruct())
			require.Truef(
				t,
				diff.Zero(),
				"expected no differences: %s", diff,
			)
		})

		t.Run("when something is missing", func(t *testing.T) {
			require.Equal(
				t,
				enums.Diff{
					Missing: enums.Collection{
						enums.Enum{
							Type:      "github.com/gaqzi/enums/testdata/full.FlagStruct",
							Name:      "FlagDefaultOn",
							FieldName: "Name",
							Value:     `"flag-default-on"`,
						},
					},
				},
				fsCollection.Diff(full.MissingFlagStruct()),
			)
		})
	})
}
