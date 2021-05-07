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
		require.True(
			t,
			collection.Diff(full.AllFlags()).Zero(),
			"expected no differences",
		)
	})

	t.Run("Indicates a difference when one is set", func(t *testing.T) {
		require.False(
			t,
			collection.Diff(full.MissingFlags()).Zero(),
			"expected to have indicated a diff",
		)
	})
}
