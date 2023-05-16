/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

// Package kontext is a set of utilities for bundling up a local directory
// context within CLIs into a self-extracting container image that when run
// will expand its payload into the working directory.
// It is (originally) based on github.com/mattmoor/kontext
package kontext
