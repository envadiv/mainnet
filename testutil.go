package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func RequireDecEqual(t *testing.T, expected, actual Dec, msgAndArgs ...interface{}) {
	if !expected.IsEqual(actual) {
		assert.Fail(t, fmt.Sprintf("%s != %s", expected, actual), msgAndArgs...)
		t.FailNow()
	}
}

func RequireAccountEqual(t *testing.T, expected, actual Account) {
	RequireDecEqual(t, expected.TotalPassage, actual.TotalPassage, "TotalPassage")
	require.Equal(t, expected.Address, actual.Address, "Address")
	require.Equal(t, len(expected.Distributions), len(actual.Distributions), "len(Distributions)")
	for i := 0; i < len(expected.Distributions); i++ {
		require.Equal(t, expected.Distributions[i].Time, actual.Distributions[i].Time, "Distribution", i)
		RequireDecEqual(t, expected.Distributions[i].Passage, actual.Distributions[i].Passage, "Distribution", i)
	}
}
