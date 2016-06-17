// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/goulash/audio"
	"github.com/goulash/osutil"
)

type AudioOperation int

const (
	SkipAudio AudioOperation = iota
	IgnoreAudio
	TranscodeAudio
	UpdateAudio
	CopyAudio
)

type Audio interface {
	audio.Metadata
	os.FileInfo
}

func entryAudio(e *Entry) Audio {
	if e == nil {
		return nil
	}
	type aud struct {
		audio.Metadata
		os.FileInfo
	}
	return &aud{
		Metadata: e.Data().(audio.Metadata),
		FileInfo: e.FileInfo(),
	}
}

type Operator interface {
	// WhichExt takes the source metadata and returns the
	// expected destination extension of the file, such as ".mp3".
	// If "" is returned, the extension remains unchanged.
	WhichExt(src Audio) string

	// Which returns an audio operation that should be
	// performed, based on src and dst (possibly nil).
	Which(src, dst Audio) AudioOperation

	// Feedback
	Ok(dst string) error
	Ignore(dst string) error
	Error(err error) error
	Warn(err error) error

	// Operations
	RemoveDir(dst string) error
	CreateDir(dst string) error

	RemoveFile(dst string) error
	CopyFile(src, dst string) error
	Transcode(src, dst string, md Audio) error
	Update(src, dst string, md Audio) error
}

type Planner struct {
	IgnoreData   bool
	DeleteBefore bool
	TranscodeAll bool

	op  Operator
	src *Database
	dst *Database
}

func NewPlanner(src, dst *Database, op Operator) *Planner {
	return &Planner{
		op:  op,
		src: src,
		dst: dst,
	}
}

func (p *Planner) Plan() error {
	if p.src == nil || p.dst == nil || p.op == nil {
		return errors.New("planner contains nil fields")
	}
	src := p.src.Root()
	if !src.IsDir() {
		return errors.New("src must be a directory")
	}
	dst := p.dst.Root()
	if !dst.IsDir() {
		return errors.New("dst must be a directory")
	}
	return p.planDir(src, dst)
}

func (p *Planner) planDir(src, dst *Entry) error {
	// We know that both src and dst are directories, or dst doesn't exist.
	if dst != nil && p.DeleteBefore {
		// Delete extra files on destination first, if dst exists.
		expect := make(map[string]bool)
		for _, e := range src.Children() {
			expect[p.dkey(e)] = true
		}

		for _, e := range dst.Children() {
			if !expect[e.Key()] {
				p.remove(e)
			}
		}
	} else {
		// Create the directory if it doesn't exist.
		path := p.dpath(src.Key())
		ex, err := osutil.DirExists(path)
		if err != nil {
			return err
		}
		if !ex {
			err := p.op.CreateDir(p.dpath(src.Key()))
			if err != nil {
				return err
			}
		}
	}

	// Sync source to destination
	for _, s := range src.Children() {
		d := p.dst.Get(p.dkey(s))

		// Eliminate the possibility of a mismatch
		if d != nil && (s.IsDir() != d.IsDir() || s.IsMusic() != d.IsMusic()) {
			err := p.remove(dst)
			if err != nil {
				return err
			}
			dst = nil
		}

		var err error
		if s.IsDir() {
			err = p.planDir(s, d)
		} else {
			err = p.planFile(s, d)
		}
		if err != nil {
			err = p.op.Warn(err)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// planFile synchronizes src to dst, which may be nil.
func (p *Planner) planFile(src, dst *Entry) error {
	path := p.dpath(src.RelPath())
	if src.IsMusic() {
		sa := entryAudio(src)
		da := entryAudio(dst)

		path = p.dpath(p.dkey(src))
		switch p.op.Which(sa, da) {
		case SkipAudio:
			return p.op.Ok(path)
		case CopyAudio:
			return p.op.CopyFile(src.AbsPath(), path)
		case TranscodeAudio:
			return p.op.Transcode(src.AbsPath(), path, sa)
		case UpdateAudio:
			return p.op.Update(src.AbsPath(), path, da)
		case IgnoreAudio:
			return p.op.Ignore(path)
		default:
			panic("unknown audio operation")
		}
	} else if !p.IgnoreData {
		return p.op.CopyFile(src.AbsPath(), path)
	}
	return p.op.Ignore(path)
}

// dpath returns the absolute destination path, given the key.
func (p *Planner) dpath(key string) string {
	return filepath.Join(p.dst.Path(), key)
}

// dkey returns the destination key, which also takes into account whether the
// file should be transcoded or not.
func (p *Planner) dkey(src *Entry) string {
	if !src.IsMusic() {
		return src.Key()
	}

	ext := p.op.WhichExt(entryAudio(src))
	key := src.Key()
	_, oxt := src.FilenameExt()
	return key[:len(key)-len(oxt)] + ext // this might not work
}

func (p *Planner) pathWithExt(src *Entry, ext string) string {
	path := filepath.Join(p.dst.Path(), src.RelPath())
	if ext == "" {
		return path
	}
	_, oxt := src.FilenameExt()
	return path[:len(path)-len(oxt)] + ext // this might not work
}

func (p *Planner) remove(dst *Entry) error {
	var err error
	if dst.IsDir() {
		err = p.op.RemoveDir(dst.AbsPath())
	} else {
		err = p.op.RemoveFile(dst.AbsPath())
	}
	if err != nil {
		return p.op.Warn(err)
	}
	return nil
}
