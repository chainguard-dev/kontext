/*
Copyright 2023 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package kontext

import (
	"context"
	"testing"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

type descriptorImpl struct {
	i   v1.Image
	ii  v1.ImageIndex
	err error
}

func (di *descriptorImpl) ImageIndex() (v1.ImageIndex, error) {
	return di.ii, di.err
}

func (di *descriptorImpl) Image() (v1.Image, error) {
	return di.i, di.err
}

func TestBundleIndex(t *testing.T) {
	want := int64(5)
	remoteGet = func(name.Reference, ...remote.Option) (types.MediaType, descriptor, error) {
		ii, err := random.Index(3, 4, want)
		return types.OCIImageIndex, &descriptorImpl{ii: ii}, err
	}
	remoteWriteIndex = func(name.Reference, v1.ImageIndex, ...remote.Option) error {
		return nil
	}
	remoteWrite = func(name.Reference, v1.Image, ...remote.Option) error {
		t.Error("Unexpected call to remoteWrite")
		return nil
	}

	source, _ := name.NewTag("ghcr.io/blah/blurg")
	tag, _ := name.NewTag("gcr.io/buffoon/banana")

	got := int64(0)

	_, err := Map(context.Background(), source, tag,
		func(ctx context.Context, img v1.Image) (v1.Image, error) {
			got++
			return img, nil
		})
	if err != nil {
		t.Error("Map() =", err)
	}

	if got != want {
		t.Errorf("callback called %d times, wanted %d", got, want)
	}
}

func TestBundleImage(t *testing.T) {
	remoteGet = func(name.Reference, ...remote.Option) (types.MediaType, descriptor, error) {
		i, err := random.Image(3, 4)
		return types.OCIManifestSchema1, &descriptorImpl{i: i}, err
	}
	remoteWriteIndex = func(name.Reference, v1.ImageIndex, ...remote.Option) error {
		t.Error("Unexpected call to remoteWriteIndex")
		return nil
	}
	remoteWrite = func(name.Reference, v1.Image, ...remote.Option) error {
		return nil
	}

	source, _ := name.NewTag("ghcr.io/blah/blurg")
	tag, _ := name.NewTag("gcr.io/buffoon/banana")

	got := int64(0)

	_, err := Map(context.Background(), source, tag,
		func(ctx context.Context, img v1.Image) (v1.Image, error) {
			got++
			return img, nil
		})
	if err != nil {
		t.Error("Map() =", err)
	}

	if want := int64(1); got != want {
		t.Errorf("callback called %d times, wanted %d", got, want)
	}
}
