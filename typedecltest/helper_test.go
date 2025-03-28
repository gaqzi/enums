package typedecltest_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gaqzi/typedecl/testdata/full"
	"github.com/gaqzi/typedecl/typedecltest"
)

type tLogger struct {
	failCalled   int
	helperCalled int
	log          []interface{}
}

func (t *tLogger) Fail() {
	t.failCalled++
}

func (t *tLogger) Helper() {
	t.helperCalled++
}

func (t *tLogger) Log(args ...interface{}) {
	t.log = append(t.log, args)
}

func TestNoDiff(t *testing.T) {
	t.Run("Does not call fail or log when there is no diff", func(t *testing.T) {
		tl := new(tLogger)

		require.True(t, typedecltest.NoDiff(
			tl,
			"../testdata/full",
			"full.Flag",
			full.AllFlags(),
			"expected no differences",
		))

		require.Equal(
			t,
			&tLogger{helperCalled: 1},
			tl,
			"expected to only have called helper since our test is passing",
		)
	})

	t.Run("Fails the case when diffs are found", func(t *testing.T) {
		tl := new(tLogger)

		require.False(t, typedecltest.NoDiff(
			tl,
			"../testdata/full",
			"full.Flag",
			full.MissingFlags(),
			"expected a missing difference",
		))

		require.Equal(
			t,
			&tLogger{
				failCalled:   1,
				helperCalled: 1,
				log: []interface{}{
					[]interface{}{
						"expected a missing difference\n" +
							"Matches declared but not part of actual:\n" +
							"\tDeployOneThing = \"deploy-one-thing\"\n",
					},
				},
			},
			tl,
			"expected to have called with a precise error and to have called fail",
		)
	})
}
