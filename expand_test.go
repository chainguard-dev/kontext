/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kontext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExpand(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("os.Getwd() =", err)
	}
	defer os.Chdir(wd)

	// compute the source's bundle hash
	src := filepath.Join(wd, "testdata")
	lSrc, err := bundle(src)
	if err != nil {
		t.Fatal("bundle() =", err)
	}
	hSrc, err := lSrc.Digest()
	if err != nil {
		t.Fatal("lSrc.Digest() =", err)
	}

	fmt.Println("src:", hSrc.String())

	// "expand" testdata into a new temporary directory.
	dest, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal("ioutil.TempDir() =", err)
	}
	// t.Logf("tmp: %s", dest)
	defer os.RemoveAll(dest)
	if err := os.Chdir(dest); err != nil {
		t.Fatal("os.Chdir() =", err)
	}

	if err := expand(context.Background(), src); err != nil {
		t.Fatal("expand() =", err)
	}

	if err := os.Chdir(wd); err != nil {
		t.Fatal(err)
	}

	// Now compute the destination's bundle hash
	lDest, err := bundle(dest)
	if err != nil {
		t.Fatal("bundle() =", err)
	}
	hDest, err := lDest.Digest()
	if err != nil {
		t.Fatal("lDest.Digest() =", err)
	}
	t.Logf("dest digest: %s", hDest.String())

	// This was useful for debugging digest mismatches (with defer commented out!)
	// uc, _ := lDest.Uncompressed()
	// content, _ := io.ReadAll(uc)
	// os.WriteFile(filepath.Join(dest, "dest.tar"), content, os.ModePerm)
	// uc, _ = lSrc.Uncompressed()
	// content, _ = io.ReadAll(uc)
	// os.WriteFile(filepath.Join(dest, "src.tar"), content, os.ModePerm)

	// Make sure they match.
	if hSrc != hDest {
		t.Errorf("bundle() = %v, wanted %v", hDest, hSrc)
	}
}
