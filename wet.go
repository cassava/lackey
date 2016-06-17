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

type Runner struct {
	Color *color.Colorizer

	BitrateThreshold int
	TargetQuality    int
	ForceTranscode   bool

	DryRun    bool
	Verbose   bool
	Strip     bool
	SrcPrefix string
	DstPrefix string
}

func (o *Runner) WhichExt(src Audio) string {
	return ".mp3"
}

func (o *Runner) Which(src, dst Audio) AudioOperation {
	switch src.Encoding() {
	case audio.FLAC:
		if dst == nil || o.ForceTranscode {
			return TranscodeAudio
		}
		if src.ModTime().After(dst.ModTime()) {
			return UpdateAudio
		}
		return SkipAudio
	case audio.MP3:
		if o.ForceTranscode {
			return TranscodeAudio
		}

		if dst == nil {
			if src.EncodingBitrate() > o.BitrateThreshold {
				return TranscodeAudio
			}
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

func (o *Runner) Ok(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	if o.Verbose {
		o.Color.Printf("@{g.}ok:@|@.       %s\n", dst)
	}
	return nil
}

func (o *Runner) Ignore(dst string) error {
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@{.y}ignoring:@|@. %s\n", dst)
	return nil
}

func (o *Runner) Error(err error) error {
	o.Color.Fprintf(os.Stderr, "@rerror:@|    %s\n", err)
	return err
}

func (o *Runner) Warn(err error) error {
	o.Color.Fprintf(os.Stderr, "@rwarning:@|  %s\n", err)
	return nil
}

func (o *Runner) RemoveDir(dst string) error {
	path := dst
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm -r:@|    %s\n", dst)
	if o.DryRun {
		return nil
	}

	return os.RemoveAll(path)
}

func (o *Runner) CreateDir(dst string) error {
	path := dst
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gmkdir:@|    %s\n", dst)
	if o.DryRun {
		return nil
	}

	return os.MkdirAll(path, 0777)
}

func (o *Runner) RemoveFile(dst string) error {
	path := dst
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@grm:@|       %s\n", dst)
	if o.DryRun {
		return nil
	}

	return os.Remove(path)
}

func (o *Runner) CopyFile(src, dst string) error {
	path := dst
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gcp:@|       %s\n", dst)
	if o.DryRun {
		return nil
	}

	return osutil.CopyFile(src, path)
}

func (o *Runner) Transcode(src string, dst string, md Audio) error {
	path := dst
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gencode:@|   %s\n", dst)

	if o.TargetQuality < 0 || 9 < o.TargetQuality {
		return errors.New("VBR quality must be between 0 and 9")
	}

	if o.DryRun {
		return nil
	}

	q := strconv.FormatInt(int64(o.TargetQuality), 10)
	return exec.Command("ffmpeg", "-i", src, "-qscale:a", q, path).Run()
}

func (o *Runner) Update(src string, dst string, md Audio) error {
	path := dst
	if o.Strip {
		dst = strings.TrimPrefix(dst, o.DstPrefix)
	}
	o.Color.Printf("@gupdate:@|   %s\n", dst)
	if o.DryRun {
		return nil
	}

	return o.Transcode(src, path, md)
}
