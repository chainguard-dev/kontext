/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/chainguard-dev/kontext"
	"knative.dev/pkg/signals"
)

func main() {
	ctx := signals.NewContext()

	if err := kontext.Expand(ctx); err != nil {
		log.Fatal("Expand() =", err)
	}
}
