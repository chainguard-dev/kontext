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
	"io/fs"
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

	err := filepath.Walk(directory,
		func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// If it's a symlink, then determine where it points.
			var link string
			if fi.Mode()&fs.ModeSymlink != 0 {
				link, err = os.Readlink(path)
				if err != nil {
					return err
				}
			}

			hdr, err := tar.FileInfoHeader(fi, link)
			if err != nil {
				return err
			}
			// Give it the proper path.
			hdr.Name = filepath.Join(StoragePath, path)
			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			// If it's not a regular file, then return.
			if !fi.Mode().IsRegular() {
				return nil
			}
			// For regular files, copy the contexts to the tar writer.

			// Open the file to copy it into the tarball.
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		})
	if err != nil {
		tw.Close()
		return nil, err
	}

	tw.Close()
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
		// We run the container as root, to ensure it has permissions to chmod
		// the directory we are run in.
		cf, err := img.ConfigFile()
		if err != nil {
			return nil, err
		}
		cf.Config.User = "0"
		img, err = mutate.ConfigFile(img, cf)
		if err != nil {
			return nil, err
		}
		return mutate.AppendLayers(img, layer)
	})
}
