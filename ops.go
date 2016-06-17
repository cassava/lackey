// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"os"
	"strings"

	"github.com/goulash/audio"
	"github.com/goulash/color"
)

type DryRunner struct {
	Color *color.Colorizer

	Strip     bool
	SrcPrefix string
	DstPrefix string
}

func (o *DryRunner) ShouldTranscode(src audio.Metadata, dst audio.Metadata) string {
	switch src.Encoding() {
	case audio.FLAC:
		return ".mp3"
	case audio.MP3:
		return ".mp3"
	default:
		return ""
	}
}

func (o *DryRunner) Ok(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@{g.}ok:@|@.       %s", dst)
	return nil
}

func (o *DryRunner) Ignore(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@{.y}ignoring:@|@. %s", dst)
	return nil
}

func (o *DryRunner) Error(err error) error {
	o.Color.Fprintf(os.Stderr, "@rerror:@|    %s", err)
	return nil
}

func (o *DryRunner) Warn(err error) error {
	o.Color.Fprintf(os.Stderr, "@rwarning:@|  %s", err)
	return nil
}

func (o *DryRunner) RemoveDir(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm -r:@|    %s", dst)
	return nil
}

func (o *DryRunner) CreateDir(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gmkdir:@|    %s", dst)
	return nil
}

func (o *DryRunner) RemoveFile(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm:@|       %s", dst)
	return nil
}

func (o *DryRunner) CopyFile(src string, dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gcp:@|       %s", dst)
	return nil
}

func (o *DryRunner) Transcode(src string, dst string, md audio.Metadata) error {
	if o.Strip {
		src = strings.TrimPrefix(src, o.SrcPrefix)
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gtranscode:@|%s -> %s", src, dst)
	return nil
}

func (o *DryRunner) Update(src string, dst string, md audio.Metadata) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gupdate:@|   %s", dst)
	return nil
}
