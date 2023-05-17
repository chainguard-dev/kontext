/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kontext

import (
	"testing"
)

func TestBundleLayerIndex(t *testing.T) {
	// Check that if we bundle testdata it has the expected size.
	l, err := bundle("./testdata")
	if err != nil {
		t.Error("bundle() =", err)
	}
	sz, err := l.Size()
	if err != nil {
		t.Error("l.Size() =", err)
	}
	if got, want := sz, int64(211); got != want {
		t.Errorf("Size() = %d, wanted %d", got, want)
	}
}
