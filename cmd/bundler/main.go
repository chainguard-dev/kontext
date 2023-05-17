/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chainguard-dev/kontext"
	"github.com/google/go-containerregistry/pkg/name"
	"knative.dev/pkg/signals"
)

func main() {
	ctx := signals.NewContext()

	directory := flag.String("directory", "", "the directory to turn into a kontext bundle.")
	rawTag := flag.String("tag", "", "the tag at which to publish the kontext bundle.")
	flag.Parse()

	if *directory == "" {
		log.Fatalf("-directory must be specified")
	}
	tag, err := name.NewTag(*rawTag)
	if err != nil {
		log.Fatalf("invalid value for -tag: %v", err)
	}

	dig, err := kontext.Bundle(ctx, *directory, tag)
	if err != nil {
		log.Fatal("Bundle() =", err)
	}
	fmt.Fprint(os.Stdout, dig.String())
}
