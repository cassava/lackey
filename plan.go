// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"path/filepath"

	"github.com/goulash/audio"
)

type Operator interface {
	// ShouldTranscode takes the source and destination (possibly nil)
	// metadata and returns an extension if the file described by src
	// should be transcoded, and "" otherwise.
	ShouldTranscode(src, dst audio.Metadata) string

	// Feedback
	Ok(dst string) error
	Ignore(dst string) error
	Error(format string, xs ...interface{}) error

	// Operations
	CreateDir(dst string) error
	RemoveDir(dst string) error

	Copy(src, dst string) error
	Transcode(src, dst string, md audio.Metadata) error
	Update(src, dst string, md audio.Metadata) error
	Remove(dst string) error
}

func Plan(src, dst *Entry, op Operator) error {
	if src.IsDir() {
		return planDir(src, dst, op)
	}
}

type planner struct {
	src *Entry
	dst *Entry
}

func (p *planner) path(src *Entry) string {
	return filepath.Join(p.dst.RootPath(), src.RelPath())
}

func (p *planner) pathWithExt(src *Entry, ext string) string {
	path := filepath.Join(p.dst.RootPath(), src.RelPath())
	if ext == "" {
		return path
	}
	_, oxt := str.FilenameExt()
	return path[:len(path)-len(oxt)] + ext // this might not work
}

func (p *planner) planDir(src, dst *Entry, op Operator) error {

}

func (p *planner) planFile(src, dst *Entry, op Operator) error {
	if src == nil {
		panic("source cannot be nil")
	}

	if dst != nil && (dst.IsDir() || src.IsMusic() != dst.IsMusic()) {
		err := op.RemoveDir(dst.AbsPath())
		if err != nil {
			return err
		}
		dst = nil
	}

	if src.IsMusic() {
		if src.IsMusic() {
			var mdIn, mdOut audio.Metadata

			mdIn, ok := src.Data().(audio.Metadata)
			if !ok {
				panic("filetype is audio but there is no metadata")
			}

			if dst != nil {
				mdOut, ok = dst.Data().(audio.Metadata)
				if !ok {
					panic("filetype is audio but there is no metadata")
				}
			}

			ext := op.ShouldTranscode(mdIn, mdOut)
			if ext != "" {
				return op.Transcode(src.AbsPath(), p.pathWithExt(src, ext), md)
			}
		}
	}
	return op.Copy(src.AbsPath(), p.path(src))
}

/*
func planSync(src, dst *Entry, p Operator) error {
		switch {
		case src.IsDir():
			if err := p.CreateDir(); err != nil {
				return err
			}
			src.
		case src.IsMusic():
			if
			fallthrough
		default:
			return p.Copy(src, dst)
		}
}

func planFile(


	if src == nil && dst == nil {
		panic("both source and destination nil")
	} else if src == nil {
		// Source file is missing, so delete the destination entirely
		if dst.IsDir() {
			return p.RemoveDir()
		}
		return p.Remove(dst)
	} else if dst == nil {
		// Destination file is missing, so sync everything over
		return planSync(src, dst, p)
	}

	if src.Key() != dst.Key() {
		panic("key must be identical")
	}
	if src.IsDir() != dst.IsDir() || src.IsMusic() != dst.IsMusic() {
		// File type mismatch, so delete destination and then sync.
		var err error
		if dst.IsDir() {
			err = p.RemoveDir(dst.AbsPath())
		} else {
			err = p.Remove(dst.AbsPath())
		}
		if err != nil {
			return err
		}
		return planSync(src, dst, p)
	}

	if src.IsMusic() {

	}
}

func (p *Planner) plan(src, dst *Entry) *Op {

	var f Flag
	if src.FileInfo().ModTime() > dst.FileInfo().ModTime() {
		f |= Time
	}
	if src.IsMusic() {
		sm := src.Data.(Metadata)
		dm := dst.Data.(Metadata)
		if !audio.MetadataEquals(sm, dm) {
			f |= Metadata
		}
		if qf(sm, dm) {
			f |= Quality
		}
	} else {
		if src.Size() != dst.Size() {
			f |= Size
		}
	}

	if f == Unknown {
		return Equal, nil
	} else {
		return Unequeal | f, nil
	}

}

*/
