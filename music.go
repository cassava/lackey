// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/goulash/osutil"
)

type FileInfo interface {
	IsMusic() bool
}

type MusicInfo interface {
	Encoding() Codec
	Bitrate() int
}

type Codec int

const (
	Unknown Codec = iota
	FLAC
	MP3
	OGG
	AAC
	M4A
)

var codecExt = map[Codec]string{
	FLAC: ".flac",
	MP3:  ".mp3",
	OGG:  ".ogg",
	AAC:  ".aac",
	M4A:  ".m4a",
}

func (c Codec) Ext() string {
	s, _ := codecExt[c]
	return s
}

func IdentifyCodec(file string) Codec {
	ext := filepath.Ext(file)
	for k, v := range codecExt {
		if ext == v {
			return k
		}
	}
	return Unknown
}

type fileInfo struct {
	isMusic bool
}

func (fi fileInfo) IsMusic() bool { return fi.isMusic }

func ReadFileInfo(file string) (FileInfo, error) {
	return &fileInfo{IdentifyCodec(file) != Unknown}, nil
}

type musicInfo struct {
	codec   Codec
	bitrate int
}

func (mi *musicInfo) Encoding() Codec { return mi.codec }
func (mi *musicInfo) Bitrate() int    { return mi.bitrate }

func ReadMusicInfo(file string) (MusicInfo, error) {
	mi := &musicInfo{
		codec: IdentifyCodec(file),
	}
	switch mi.codec {
	case MP3:
		r, err := ReadBitrateFromMP3(file)
		if err != nil {
			return nil, err
		}
		mi.bitrate = r
	case FLAC:
		// TODO: get this right
		mi.bitrate = 1500
	default:
		return nil, errors.New("cannot read music info for this codec yet")
	}

	return mi, nil
}

func ConvertMusicFile(src, dst string, quality int) error {
	switch IdentifyCodec(dst) {
	case MP3:
		return ConvertMusicFileToMP3(src, dst, quality)
	default:
		return errors.New("cannot convert to this format yet")
	}
}

// MP3 {{{

func ReadBitrateFromMP3(file string) (int, error) {
	out, err := exec.Command("mp3info", "-r", "m", "-p", "%r", file).Output()
	if err != nil {
		return 0, err
	}
	br, err := strconv.ParseInt(string(out), 10, 0)
	if err != nil {
		return 0, err
	}
	return int(br), nil
}

func ConvertMusicFileToMP3(src, dst string, vbrQuality int) error {
	ex, err := osutil.Exists(dst)
	if err != nil {
		return err
	}
	if ex {
		return os.ErrExist
	}

	if vbrQuality < 0 || 9 < vbrQuality {
		return errors.New("VBR quality must be between 0 and 9")
	}
	q := strconv.FormatInt(int64(vbrQuality), 10)

	return exec.Command("ffmpeg", "-i", src, "-qscale:a", q, dst).Run()
}

// }}}
