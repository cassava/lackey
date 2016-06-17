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

	BitrateThreshold int
	TargetQuality    int

	Strip     bool
	SrcPrefix string
	DstPrefix string
}

func (o *DryRunner) WhichExt(src Audio) string {
	return ".mp3"
}

func (o *DryRunner) Which(src, dst Audio) AudioOperation {
	switch src.Encoding() {
	case audio.FLAC:
		return TranscodeAudio
	case audio.MP3:
		if src.EncodingBitrate() > o.BitrateThreshold {
			return TranscodeAudio
		}
		if dst == nil {
			return CopyAudio
		}
		if src.ModTime().After(dst.ModTime()) {
			return UpdateAudio
		}
		return SkipAudio
	default:
		return IgnoreAudio
	}
}

func (o *DryRunner) Ok(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@{g.}ok:@|@.       %s\n", dst)
	return nil
}

func (o *DryRunner) Ignore(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@{.y}ignoring:@|@. %s\n", dst)
	return nil
}

func (o *DryRunner) Error(err error) error {
	o.Color.Fprintf(os.Stderr, "@rerror:@|    %s\n", err)
	return nil
}

func (o *DryRunner) Warn(err error) error {
	o.Color.Fprintf(os.Stderr, "@rwarning:@|  %s\n", err)
	return nil
}

func (o *DryRunner) RemoveDir(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm -r:@|    %s\n", dst)
	return nil
}

func (o *DryRunner) CreateDir(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gmkdir:@|    %s\n", dst)
	return nil
}

func (o *DryRunner) RemoveFile(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm:@|       %s\n", dst)
	return nil
}

func (o *DryRunner) CopyFile(src string, dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gcp:@|       %s\n", dst)
	return nil
}

func (o *DryRunner) Transcode(src string, dst string, md Audio) error {
	if o.Strip {
		src = strings.TrimPrefix(src, o.SrcPrefix)
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gencode:@|   %s\n", dst)
	return nil
}

func (o *DryRunner) Update(src string, dst string, md Audio) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gupdate:@|   %s\n", dst)
	return nil
}
