// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package mp3

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cassava/lackey/music"
	"github.com/dhowden/tag"
)

var ErrNotMP3 = errors.New("file is not an MP3")

func Assert(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, ft, err := tag.Identify(f)
	if err != nil {
		return err
	}
	if ft != tag.MP3 {
		return ErrNotMP3
	}
	return nil
}

// ReadMetadata {{{

var metaReader = func(_ string) (int, time.Duration, error) {
	return 0, 0, errors.New("system program to read bitrate and song length missing")
}

func ReadMetadata(file string) (*Metadata, error) {
	if err := Assert(file); err != nil {
		return nil, err
	}

	// Read length and bitrate
	b, d, err := metaReader(file)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tm, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}

	return &Metadata{
		Metadata: tm,
		length:   d,
		bitrate:  b,
		codec:    music.MP3,
	}, nil
}

func init() {
	// Find out what program we should use to find out various information.
	readers := []struct {
		name string
		fn   func(string) (int, time.Duration, error)
	}{
		{"mp3info", mp3info},
		{"exiftool", exiftool},
	}
	var found bool
	for _, r := range readers {
		if _, err := exec.LookPath(r.name); err == nil {
			found = true
			metaReader = r.fn
			break
		}
	}
	if found {
		// Ensure that the programs we need for getting the metadata are available.
		music.MetadataReaders[music.MP3] = func(file string) (music.Metadata, error) {
			return ReadMetadata(file)
		}
	}
}

func mp3info(file string) (r int, d time.Duration, err error) {
	cmd := exec.Command("mp3info", "-r", "m", "-p", "%r\t%Ss", file)
	bs, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	perr := func(err error) (int, time.Duration, error) {
		return 0, 0, fmt.Errorf("cannot parse mp3info output %q: %s", string(bs), err)
	}

	rs := strings.Split(string(bs), "\t")
	if len(rs) != 2 {
		return perr(errors.New("incorrect number of fields"))
	}

	ri, err := strconv.ParseInt(rs[0], 10, 0)
	if err != nil {
		return perr(err)
	}
	r = int(ri)
	d, err = time.ParseDuration(rs[1])
	if err != nil {
		return perr(err)
	}
	return r, d, nil
}

func exiftool(_ string) (int, time.Duration, error) {
	return 0, 0, errors.New("not implemented")
}

// }}}

// WriteMetadata {{{

func WriteMetadata(file string, m music.Metadata) error {
	return errors.New("not implemented")
}

// }}}

// Metadata {{{

var _ = music.Metadata(new(Metadata))

type Metadata struct {
	// Metadata is an interface that already implements:
	//
	//  Format() tag.Format
	//  FileType() tag.FileType
	//
	//  Title() string
	//  Album() string
	//  Artist() string
	//  AlbumArtist() string
	//  Composer() string
	//  Year() int
	//  Genre() string
	//  Track() (int, int)
	//  Disc() (int, int)
	//
	//  Picture() *tag.Picture
	//  Lyrics() string
	//  Raw() map[string]interface{}
	tag.Metadata

	length  time.Duration
	bitrate int
	codec   music.Codec
}

func (m *Metadata) Year() int {
	// Easytag writes Year to the TDRC/TDAT field. Retrieve it from there if
	// necessary (in other words, if the normal year is not set).
	year := m.Metadata.Year()
	if year != 0 {
		return year
	}

	if y := m.rawInt("TDRC"); y != 0 {
		return y
	}
	if y := m.rawInt("TDAT"); y != 0 {
		return y
	}
	return 0
}

func (m *Metadata) Length() time.Duration    { return m.length }
func (m *Metadata) Comment() string          { return m.rawString("CXXX") }
func (m *Metadata) Website() string          { return m.rawComment("WXXX") }
func (m *Metadata) Copyright() string        { return m.rawString("TCOP") }
func (m *Metadata) Encoding() music.Codec    { return m.codec }
func (m *Metadata) EncodedBy() string        { return m.rawString("TENC") }
func (m *Metadata) EncodingBitrate() int     { return m.bitrate }
func (m *Metadata) EncoderSettings() string  { return m.rawString("TSSE") }
func (m *Metadata) OriginalFilename() string { return m.rawString("TOFN") }
func (m *Metadata) PrivateData() []byte      { return m.rawBytes("PRIV") }

func (m *Metadata) rawBytes(key string) []byte {
	if v, ok := m.Raw()[key]; ok {
		s, ok := v.([]byte)
		if !ok {
			panic(fmt.Sprintf("expecting []byte, got %#v", v))
		}
		return s
	}
	return []byte{}
}

func (m *Metadata) rawString(key string) string {
	if v, ok := m.Raw()[key]; ok {
		s, ok := v.(string)
		if !ok {
			panic(fmt.Sprintf("expecting string, got %#v", v))
		}
		return s
	}
	return ""
}

func (m *Metadata) rawInt(key string) int {
	if v, ok := m.Raw()[key]; ok {
		if i, ok := v.(int); ok {
			return i
		} else if s, ok := v.(string); ok {
			i, _ := strconv.ParseInt(s, 10, 0)
			return int(i)
		}
	}
	return 0
}

func (m *Metadata) rawComment(key string) string {
	if v, ok := m.Raw()[key]; ok {
		s, ok := v.(*tag.Comm)
		if !ok {
			panic(fmt.Sprintf("expecting *tag.Comm, got %#v", v))
		}
		return s.Text
	}
	return ""
}

// }}}
