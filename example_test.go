package enums_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gaqzi/enums"
	"github.com/gaqzi/enums/testdata/full"
)

func TestIntegration(t *testing.T) {
	matches, err := enums.All("./testdata/full", "full.Flag")
	require.NoError(t, err)

	t.Run("Show when there is no difference", func(t *testing.T) {
		require.Equal(
			t,
			enums.Diff{},
			matches.Diff(full.AllFlags()),
			"expected no differences",
		)
	})

	t.Run("Show when there is extra value in the actual", func(t *testing.T) {
		require.Equal(
			t,
			enums.Diff{Extra: []string{`"m000"`}},
			matches.Diff(full.ExtraFlags()),
			"expected to have an extra value indicated",
		)
	})

	t.Run("Show when there is a missing value in the actual", func(t *testing.T) {
		require.Equal(
			t,
			enums.Diff{Missing: enums.Collection{
				{
					Type:  "github.com/gaqzi/enums/testdata/full.Flag",
					Name:  "DeployOneThing",
					Value: `"deploy-one-thing"`,
				},
			}},
			matches.Diff(full.MissingFlags()),
			"expected to have an extra value indicated",
		)
	})
}
