/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kontext

import (
	"context"
	"io"
	"io/fs"
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

	// In the first pass, expand all of the files as quickly as possible,
	// granting broad file permissions.
	{
		eg, ctx := errgroup.WithContext(ctx)
		eg.SetLimit(100)
		if err := filepath.WalkDir(base, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path == base {
				return nil
			}

			// Add each file to the backlog.
			eg.Go(func() (err error) {
				// If the context is canceled, then bail out early.
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				relativePath := path[len(base)+1:]
				target := filepath.Join(targetPath, relativePath)

				if err := os.MkdirAll(filepath.Dir(target), 0777); err != nil {
					return err
				}
				fi, err := info.Info()
				if err != nil {
					return err
				}
				if info.IsDir() {
					return os.MkdirAll(target, fi.Mode())
				} else if info.Type()&fs.ModeSymlink != 0 {
					// It is not practical to test this path because there is not
					// a portable way to change the mtime of the symlink
					// https://github.com/golang/go/issues/3951
					link, err := os.Readlink(path)
					if err != nil {
						return err
					}
					return os.Symlink(link, target)
				}
				return copyFile(path, target)
			})

			return nil
		}); err != nil {
			return err
		}
		if err := eg.Wait(); err != nil {
			return err
		}
	}

	// In the final pass, fixup permissions and mtimes
	{
		eg, ctx := errgroup.WithContext(ctx)
		eg.SetLimit(100)
		if err := filepath.WalkDir(base, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Add each file to the backlog.
			eg.Go(func() (err error) {
				// If the context is canceled, then bail out early.
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				target := targetPath
				if path != base {
					relativePath := path[len(base)+1:]
					target = filepath.Join(targetPath, relativePath)
				}

				fi, err := info.Info()
				if err != nil {
					return err
				}
				// Set the permissions and mtime
				if err := os.Chmod(target, fi.Mode()); err != nil {
					return err
				}
				// Skip symlinks due to:
				// https://github.com/golang/go/issues/3951
				if info.Type()&fs.ModeSymlink != 0 {
					return nil
				}
				return os.Chtimes(target, fi.ModTime(), fi.ModTime())
			})

			return nil
		}); err != nil {
			return err
		}

		// Wait for the work to be done.
		return eg.Wait()
	}
}

// Expand recursively copies the current working directory into StoragePath.
func Expand(ctx context.Context) error {
	return expand(ctx, StoragePath)
}
