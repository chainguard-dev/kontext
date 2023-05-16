/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kontext

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"
)

const (
	// StoragePath is where in the container image the files are placed.
	StoragePath = "/var/run/kontext"
)

func copyFile(src, dest string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}

func expand(ctx context.Context, base string) error {
	targetPath, err := os.Getwd()
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(100)

	if err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == base {
			return nil
		}

		// Add each file to the backlog.
		eg.Go(func() error {
			// If the context is canceled, then bail out early.
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			relativePath := path[len(base)+1:]
			target := filepath.Join(targetPath, relativePath)

			if info.IsDir() {
				return os.MkdirAll(target, os.ModePerm)
			}
			if !info.Mode().IsRegular() {
				log.Printf("Skipping irregular file: %q", relativePath)
				return nil
			}
			if err := os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
				return err
			}
			return copyFile(path, target)
		})

		return nil
	}); err != nil {
		return err
	}

	// Wait for the work to be done.
	return eg.Wait()
}

// Expand recursively copies the current working directory into StoragePath.
func Expand(ctx context.Context) error {
	return expand(ctx, StoragePath)
}
