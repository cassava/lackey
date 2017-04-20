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

type ExecError struct {
	Err    error
	Output string
}

func (err *ExecError) Error() string { return err.Err.Error() }

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

func (o *Runner) WhichExt(_ Audio) string {
	// At the moment, we assume that we always want MP3 output
	return ".mp3"
}

// canEncode returns true if this runner can encode the codec.
func (o *Runner) canEncode(c audio.Codec) bool {
	return c == audio.FLAC || c == audio.MP3
}

func (o *Runner) whichMP3(src, dst Audio) AudioOperation {
	if !dst.IsExists() {
		sm := src.Metadata()
		if sm.EncodingBitrate() > o.BitrateThreshold {
			return TranscodeAudio
		}
		return CopyAudio
	}
	sfi, dfi := src.FileInfo(), dst.FileInfo()
	if sfi.ModTime().After(dfi.ModTime()) {
		return UpdateAudio
	}
	return SkipAudio
}

func (o *Runner) which(src, dst Audio) AudioOperation {
	if !dst.IsExists() {
		return TranscodeAudio
	}
	sfi, dfi := src.FileInfo(), dst.FileInfo()
	if sfi.ModTime().After(dfi.ModTime()) {
		return UpdateAudio
	}
	return SkipAudio
}

func (o *Runner) Which(src, dst Audio) AudioOperation {
	// Find out when we can skip work.
	if !o.canEncode(src.Encoding()) {
		return IgnoreAudio
	}
	if o.ForceTranscode {
		return TranscodeAudio
	}

	switch src.Encoding() {
	case audio.MP3:
		return o.whichMP3(src, dst)
	default:
		return o.which(src, dst)
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
	if e, ok := err.(*ExecError); ok {
		o.Color.Fprintf(os.Stderr, "@routput:@|\n%s\n", e.Output)
	}
	return err
}

func (o *Runner) Warn(err error) error {
	o.Color.Fprintf(os.Stderr, "@rwarning:@|  %s\n", err)
	if e, ok := err.(*ExecError); ok {
		o.Color.Fprintf(os.Stderr, "@routput:@|\n%s\n", e.Output)
	}
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
	var bs []byte
	var err error
	if md.Encoding() == audio.MP3 {
		// We are much more reliable using lame directly than over ffmpeg when downsampling
		// MP3 files directly.
		bs, err = exec.Command("lame", "--mp3input", "-h", "-V"+q, src, path).CombinedOutput()
	} else {
		bs, err = exec.Command("ffmpeg", "-i", src, "-qscale:a", q, path).CombinedOutput()
	}
	if err != nil {
		return &ExecError{
			Err:    err,
			Output: string(bs),
		}
	}
	return nil
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

	err := os.Remove(path)
	if err != nil {
		return err
	}
	o.Color.Printf(" -> ")
	return o.Transcode(src, path, md)
}
