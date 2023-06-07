/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kontext

import (
	"bytes"
	"context"
	"io"

	apkfs "github.com/chainguard-dev/go-apk/pkg/fs"
	apktar "github.com/chainguard-dev/go-apk/pkg/tarball"
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
	fsys := apkfs.DirFS(directory)

	tctx, err := apktar.NewContext(apktar.WithBaseDir(StoragePath))
	if err != nil {
		return nil, err
	}

	if err := tctx.WriteTar(context.Background(), buf, fsys); err != nil {
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
