// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/goulash/audio"
	"github.com/goulash/color"
	"github.com/goulash/osutil"
)

type WetRunner struct {
	Color *color.Colorizer

	BitrateThreshold int
	TargetQuality    int

	Verbose   bool
	Strip     bool
	SrcPrefix string
	DstPrefix string
}

func (o *WetRunner) WhichExt(src Audio) string {
	return ".mp3"
}

func (o *WetRunner) Which(src, dst Audio) AudioOperation {
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

func (o *WetRunner) Ok(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	if o.Verbose {
		o.Color.Printf("@{g.}ok:@|@.       %s\n", dst)
	}
	return nil
}

func (o *WetRunner) Ignore(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@{.y}ignoring:@|@. %s\n", dst)
	return nil
}

func (o *WetRunner) Error(err error) error {
	o.Color.Fprintf(os.Stderr, "@rerror:@|    %s\n", err)
	return err
}

func (o *WetRunner) Warn(err error) error {
	o.Color.Fprintf(os.Stderr, "@rwarning:@|  %s\n", err)
	return nil
}

func (o *WetRunner) RemoveDir(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm -r:@|    %s\n", dst)
	return os.RemoveAll(dst)
}

func (o *WetRunner) CreateDir(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gmkdir:@|    %s\n", dst)
	return os.MkdirAll(dst, 0777)
}

func (o *WetRunner) RemoveFile(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm:@|       %s\n", dst)
	return os.Remove(dst)
}

func (o *WetRunner) CopyFile(src, dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gcp:@|       %s\n", dst)
	return osutil.CopyFile(src, dst)
}

func (o *WetRunner) Transcode(src string, dst string, md Audio) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gencode:@|   %s\n", dst)

	if o.TargetQuality < 0 || 9 < o.TargetQuality {
		return errors.New("VBR quality must be between 0 and 9")
	}
	q := strconv.FormatInt(int64(o.TargetQuality), 10)
	return exec.Command("ffmmpeg", "-i", src, "-qscale:a", q, dst).Run()
}

func (o *WetRunner) Update(src string, dst string, md Audio) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gupdate:@|   %s\n", dst)
	o.Transcode(src, dst, md)
	return nil
}
