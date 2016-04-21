// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package flac

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

var ErrNotFLAC = errors.New("file is not an FLAC")

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
	if ft != tag.FLAC {
		return ErrNotFLAC
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
		codec:    music.FLAC,
	}, nil
}

func init() {
	readers := []struct {
		name string
		fn   func(string) (int, time.Duration, error)
	}{
		{"metaflac", metaflac},
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
		music.MetadataReaders[music.FLAC] = func(file string) (music.Metadata, error) {
			return ReadMetadata(file)
		}
	}
}

// for bitrate and time we need to use metaflac...
//
//  metaflac --show-total-samples --show-sample-rate
//
//   s = total/rate
//   bitrate = ((filesize - metadata) * 8) / (s * 1000)

func metaflac(file string) (r int, d time.Duration, err error) {
	fi, err := os.Stat(file)
	if err != nil {
		return 0, 0, err
	}

	cmd := exec.Command("metaflac", "--show-total-samples", "--show-sample-rate", file)
	bs, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	perr := func(err error) (int, time.Duration, error) {
		return 0, 0, fmt.Errorf("cannot parse metaflac output %q: %s", string(bs), err)
	}

	rs := strings.Split(strings.TrimSpace(string(bs)), "\n")
	if len(rs) != 2 {
		return perr(errors.New("incorrect number of fields"))
	}
	samples, err := strconv.ParseInt(rs[0], 10, 0)
	if err != nil {
		return perr(err)
	}
	rate, err := strconv.ParseInt(rs[1], 10, 0)
	if err != nil {
		return perr(err)
	}

	d = time.Duration((samples * int64(time.Second)) / rate)
	r = int((fi.Size() * 8) / (int64(d/time.Second) * 1000))
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

func (m *Metadata) Year() int                { return m.rawInt("date") }
func (m *Metadata) Length() time.Duration    { return m.length }
func (m *Metadata) Comment() string          { return m.rawString("comment") }
func (m *Metadata) Website() string          { return m.rawString("contact") }
func (m *Metadata) Copyright() string        { return m.rawString("copyright") }
func (m *Metadata) Encoding() music.Codec    { return m.codec }
func (m *Metadata) EncodedBy() string        { return m.rawString("encoded-by") }
func (m *Metadata) EncodingBitrate() int     { return m.bitrate }
func (m *Metadata) EncoderSettings() string  { return "" }
func (m *Metadata) OriginalFilename() string { return "" }
func (m *Metadata) PrivateData() []byte      { return []byte{} }

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

// }}}
