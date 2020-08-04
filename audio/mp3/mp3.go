// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package mp3

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/goulash/audio"
	"github.com/goulash/stat"
	"github.com/tcolgate/mp3"
)

var ErrNotMP3 = errors.New("file is not an MP3")

var Stats struct {
	Assert           stat.Run
	ReadMetadata     stat.Run
	ReadMetadataMeta stat.Run
	ReadMetadataBrDu stat.Run
	ToolMP3INFO      stat.Run
	ToolEXIFTOOL     stat.Run
}

func Assert(file string) error {
	start := time.Now()
	defer func() { Stats.Assert.Add(float64(time.Since(start))) }()

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

/*
Runtime stats:
  audio.Identify        μ=5.221595ms, σ=6.688962ms, n=12233
  audio.ReadMetadata    μ=69.826077ms, σ=89.871762ms, n=4694
  flac.Identify         μ=-2562047h47m16.854775808s, σ=-2562047h47m16.854775808s, n=0
  flac.ReadFileMetadata μ=2.412411ms, σ=4.72771ms, n=2399
  flac.ReadMetadata     μ=2.399813ms, σ=4.72703ms, n=2399
  mp3.Assert            μ=11.714µs, σ=2.509µs, n=2295
  mp3.ReadMetadata      μ=140.255738ms, σ=82.368328ms, n=2295

Runtime stats:
  audio.Identify        μ=5.497306ms, σ=6.697311ms, n=12233
  audio.ReadMetadata    μ=71.437403ms, σ=96.427961ms, n=4694
  flac.Identify         μ=-2562047h47m16.854775808s, σ=-2562047h47m16.854775808s, n=0
  flac.ReadFileMetadata μ=2.538888ms, σ=5.32524ms, n=2399
  flac.ReadMetadata     μ=2.526025ms, σ=5.324577ms, n=2399
  mp3.Assert            μ=11.556µs, σ=2.889µs, n=2295
  mp3.ReadMetadata      μ=143.417326ms, σ=94.047081ms, n=2295
  mp3.ReadMetadataMeta  μ=2.657978ms, σ=5.662125ms, n=2295
  mp3.ReadMetadataBrDu  μ=140.729045ms, σ=92.607449ms, n=2295
  mp3.ToolMP3INFO       μ=-2562047h47m16.854775808s, σ=-2562047h47m16.854775808s, n=0
  mp3.ToolEXIFTOOL      μ=-2562047h47m16.854775808s, σ=-2562047h47m16.854775808s, n=0
*/
func ReadMetadata(file string) (*Metadata, error) {
	start := time.Now()
	defer func() { Stats.ReadMetadata.Add(float64(time.Since(start))) }()

	if err := Assert(file); err != nil {
		return nil, err
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read metadata (quick 2ms)
	s1 := time.Now()
	tm, err := tag.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	Stats.ReadMetadataMeta.Add(float64(time.Since(s1)))

	// Read length and bitrate (slow 140ms)
	skipped := 0
	s2 := time.Now()
	f.Seek(0, 0)
	dec := mp3.NewDecoder(f)
	var (
		frame mp3.Frame
		dur   time.Duration
		bytes int64
	)
	for dec.Decode(&frame, &skipped) == nil {
		dur += frame.Duration()
		bytes += int64(frame.Size())
	}
	var kbps int64
	if dur != 0 {
		kbps = (bytes * 8) / int64(dur*1000/time.Second)
	}
	Stats.ReadMetadataBrDu.Add(float64(time.Since(s2)))

	return &Metadata{
		Metadata: tm,
		length:   dur,
		bitrate:  int(kbps),
		codec:    audio.MP3,
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
		audio.MetadataReaders[audio.MP3] = func(file string) (audio.Metadata, error) {
			return ReadMetadata(file)
		}
	}
}

func mp3info(file string) (r int, d time.Duration, err error) {
	start := time.Now()
	defer func() { Stats.ToolMP3INFO.Add(float64(time.Since(start))) }()

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
	start := time.Now()
	defer func() { Stats.ToolEXIFTOOL.Add(float64(time.Since(start))) }()

	return 0, 0, errors.New("not implemented")
}

// }}}

// WriteMetadata {{{

func WriteMetadata(file string, m audio.Metadata) error {
	return errors.New("not implemented")
}

// }}}

type Encoder struct {
	Path    string
	Quality int
}

func NewEncoder() *Encoder {
	return &Encoder{
		Path:    "lame",
		Quality: 4,
	}
}

// EncodeFromStdin takes a command that produces a decoded file on stdout, and
// pipes that through an mp3 encoder. It tags the MP3 with audio metadata received.
//
// For example:
//
//	enc := mp3.Encoder{q}
//  dec := exec.Command("flac", "-c", "-d", src)
//  bs, err = enc.EncodeFromStdin(dec, path, md.Metadata())
//
func (e *Encoder) EncodeFromStdin(dec *exec.Cmd, path string, md audio.Metadata) ([]byte, error) {
	slash := func(a, b int) string { return fmt.Sprintf("%d/%d", a, b) }
	q := strconv.FormatInt(int64(e.Quality), 10)
	enc := exec.Command(e.Path,
		"-h", "-V"+q,
		"--add-id3v2", "--pad-id3v2",
		"--tt", md.Title(),
		"--ta", md.Artist(),
		"--tv", fmt.Sprintf("TPE2=%s", md.AlbumArtist()),
		"--tl", md.Album(),
		"--tn", slash(md.Track()),
		"--ty", fmt.Sprintf("%d", md.Year()),
		"--tv", fmt.Sprintf("TPOS=%s", slash(md.Disc())),
		"--tg", md.Genre(),
		"--tc", md.Comment(),
		"--tv", fmt.Sprintf("TCOM=%s", md.Composer()),
		"--tv", fmt.Sprintf("WCOP=%s", md.Copyright()),
		"--tv", fmt.Sprintf("WXXX=%s", md.Website()),
		"--tv", fmt.Sprintf("TENC=%s", e.Path),
		"--tv", fmt.Sprintf("TSSE=%s", "-h -V"+q),
		"--tv", fmt.Sprintf("TOFN=%s", md.OriginalFilename()),
		// input, output
		"-", path,
	)
	// Set up the pipe
	var err error
	enc.Stdin, err = dec.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Set up the combined output
	var b bytes.Buffer
	dec.Stderr = &b
	enc.Stdout = &b
	enc.Stderr = &b

	// Run both commands
	if err := enc.Start(); err != nil {
		return b.Bytes(), err
	}
	if err := dec.Run(); err != nil {
		return b.Bytes(), err
	}
	return b.Bytes(), enc.Wait()
}

// Metadata {{{

var _ = audio.Metadata(new(Metadata))

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
	codec   audio.Codec
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
func (m *Metadata) Encoding() audio.Codec    { return m.codec }
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
