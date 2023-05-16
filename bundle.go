/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kontext

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

var (
	// BaseImageString holds a reference to a built image of ./cmd/expander
	BaseImageString = "ghcr.io/chainguard-dev/kontext:latest"
	// BaseImage is where we publish ./cmd/expander
	BaseImage, _ = name.ParseReference(BaseImageString)
)

func bundle(directory string) (v1.Layer, error) {
	buf := bytes.NewBuffer(nil)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := filepath.Walk(directory,
		func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip anything in the .git directory
			// TODO(mattmoor): expand this to .gitignore / .dockerignore?
			if fi.IsDir() && filepath.Base(path) == ".git" {
				return filepath.SkipDir
			}

			// Chase symlinks.
			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			// Compute the path relative to the base path
			relativePath, err := filepath.Rel(directory, path)
			if err != nil {
				return err
			}

			newPath := filepath.Join(StoragePath, relativePath)

			if info.Mode().IsDir() {
				return tw.WriteHeader(&tar.Header{
					Name:     newPath,
					Typeflag: tar.TypeDir,
					Mode:     0555,
				})
			}

			// Open the file to copy it into the tarball.
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// Copy the file into the image tarball.
			if err := tw.WriteHeader(&tar.Header{
				Name:     newPath,
				Size:     info.Size(),
				Typeflag: tar.TypeReg,
				// Use a fixed Mode, so that this isn't sensitive to the directory and umask
				// under which it was created. Additionally, windows can only set 0222,
				// 0444, or 0666, none of which are executable.
				Mode: 0555,
			}); err != nil {
				return err
			}
			_, err = io.Copy(tw, file)
			return err
		})
	if err != nil {
		return nil, err
	}

	return tarball.LayerFromOpener(func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(buf.Bytes())), nil
	})
}

// Bundle packages up the given directory as a self-extracting container image based
// on BaseImage and publishes it to tag.
func Bundle(ctx context.Context, directory string, tag name.Tag) (name.Digest, error) {
	layer, err := bundle(directory)
	if err != nil {
		return name.Digest{}, err
	}

	return Map(ctx, BaseImage, tag, func(ctx context.Context, img v1.Image) (v1.Image, error) {
		return mutate.AppendLayers(img, layer)
	})
}
