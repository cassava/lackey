// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cassava/lackey/audio/mp3"
	"github.com/goulash/audio"
	"github.com/goulash/color"
	"github.com/goulash/osutil"
)

type ExecError struct {
	Err    error
	Output string
}

func (err *ExecError) Error() string { return err.Err.Error() }

type Encoder interface {
	Ext() string
	CanCopy(src, dst Audio) bool
	Encode(src, dst string, md Audio) error
}

type Runner struct {
	Color *color.Colorizer

	Encoder        Encoder
	ForceTranscode bool
	CopyExtensions []string

	DryRun    bool
	Verbose   bool
	Strip     bool
	SrcPrefix string
	DstPrefix string
}

func (o *Runner) WhichExt(src Audio) string {
	// At the moment, we always have the same kind of output
	name := src.FileInfo().Name()
	for _, ext := range o.CopyExtensions {
		if strings.HasSuffix(name, ext) {
			return ext
		}
	}
	return o.Encoder.Ext()
}

// canEncode returns true if this runner can encode the codec.
func (o *Runner) canEncode(c audio.Codec) bool {
	return c == audio.FLAC || c == audio.MP3 || c == audio.M4A || c == audio.OGG
}

func (o *Runner) transcodeOrCopy(src, dst Audio) AudioOperation {
	if o.Encoder.CanCopy(src, dst) {
		return CopyAudio
	}
	return TranscodeAudio
}

func (o *Runner) Which(src, dst Audio) AudioOperation {
	// Files we should copy directly, we do so here
	name := src.FileInfo().Name()
	for _, ext := range o.CopyExtensions {
		if strings.HasSuffix(name, ext) {
			if dst.IsExists() {
				sfi, dfi := src.FileInfo(), dst.FileInfo()
				if dfi.Size() == 0 {
					return CopyAudio
				}
				if sfi.ModTime().After(dfi.ModTime()) {
					return CopyAudio
				}
				return SkipAudio
			}
			return CopyAudio
		}
	}

	if !o.canEncode(src.Encoding()) {
		return IgnoreAudio
	}

	if o.ForceTranscode {
		return TranscodeAudio
	}

	if !dst.IsExists() {
		return o.transcodeOrCopy(src, dst)
	}
	sfi, dfi := src.FileInfo(), dst.FileInfo()
	if dfi.Size() == 0 {
		return o.transcodeOrCopy(src, dst)
	}
	if sfi.ModTime().After(dfi.ModTime()) {
		return UpdateAudio
	}
	return SkipAudio
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

	if o.DryRun {
		return nil
	}

	if ex, _ := osutil.FileExists(path); ex {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}
	return o.Encoder.Encode(src, path, md)
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

type MP3Encoder struct {
	TargetQuality    int
	BitrateThreshold int
}

func (e *MP3Encoder) Ext() string { return ".mp3" }

func (e *MP3Encoder) CanCopy(src, dst Audio) bool {
	if src.Encoding() != audio.MP3 {
		return false
	}
	sm := src.Metadata()
	if sm.EncodingBitrate() > e.BitrateThreshold {
		return false
	}
	return true
}

func (e *MP3Encoder) Encode(src, dst string, md Audio) error {
	q := strconv.FormatInt(int64(e.TargetQuality), 10)
	var bs []byte
	var err error
	if md.Encoding() == audio.MP3 {
		// We are much more reliable using lame directly than over ffmpeg when downsampling
		// MP3 files directly.
		bs, err = exec.Command("lame", "--mp3input", "-h", "-V"+q, src, dst).CombinedOutput()
	} else if md.Encoding() == audio.FLAC {
		// Because ffmpeg is having some bugs, we avoid using it when possible.
		enc := mp3.NewEncoder()
		enc.Quality = e.TargetQuality
		dec := exec.Command("flac", "-c", "-d", src)
		bs, err = enc.EncodeFromStdin(dec, dst, md.Metadata())
	} else {
		bs, err = exec.Command("ffmpeg", "-i", src, "-vn", "-qscale:a", q, dst).CombinedOutput()
	}
	if err != nil {
		return &ExecError{
			Err:    err,
			Output: string(bs),
		}
	}
	return nil
}

type OPUSEncoder struct {
	Extension     string
	TargetBitrate string
}

func (e *OPUSEncoder) Ext() string {
	if e.Extension == "" {
		return ".opus"
	}
	return e.Extension
}

func (e *OPUSEncoder) CanCopy(src, dst Audio) bool {
	return false
}

func (e *OPUSEncoder) Encode(src, dst string, md Audio) error {
	bs, err := exec.Command("ffmpeg", "-i", src, "-vn", "-acodec", "libopus", "-vbr", "on",
		"-compression_level", "10", "-b:a", e.TargetBitrate, dst).CombinedOutput()
	if err != nil {
		return &ExecError{
			Err:    err,
			Output: string(bs),
		}
	}
	return nil
}
