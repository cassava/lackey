// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Jeffail/tunny"
	"github.com/goulash/osutil"
)

type Planner struct {
	IgnoreData bool
	DataExcept map[string]bool

	DeleteBefore bool
	TranscodeAll bool
	Concurrent   int

	DownscaleCover bool
	CoverSource    string
	CoverTarget    string

	op  Operator
	src *Database
	dst *Database

	pool *tunny.WorkPool
	errs chan error
	wg   sync.WaitGroup
	quit error
}

func NewPlanner(src, dst *Database, op Operator) *Planner {
	return &Planner{
		DataExcept: make(map[string]bool),
		Concurrent: runtime.NumCPU(),

		op:  op,
		src: src,
		dst: dst,
	}
}

func (p *Planner) Plan() error {
	var err error

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

	p.pool, err = tunny.CreatePoolGeneric(p.Concurrent).Open()
	if err != nil {
		return err
	}
	defer p.pool.Close()

	p.errs = make(chan error, 1)
	go func() {
		for e := range p.errs {
			err := p.op.Warn(e)
			if err != nil {
				p.quit = err
				break
			}
		}
	}()

	err = p.planDir(src, dst)
	p.wg.Wait()
	return err
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
		// Check for errors from the workers
		if p.quit != nil {
			return p.quit
		}

		d := p.dst.Get(p.dkey(s))

		// Eliminate the possibility of a mismatch
		if d != nil && (s.IsDir() != d.IsDir() || s.IsMusic() != d.IsMusic()) {
			err := p.remove(d)
			if err != nil {
				return err
			}
			d = nil
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
	path := p.dpath(p.dkey(src))
	// Check our assumption, that dst is the data for path:
	if dst != nil && path != p.dpath(dst.Key()) {
		fmt.Printf("warn: destination path %q != stat data from %q", path, p.dpath(dst.Key()))
	}

	if src.IsMusic() {
		switch p.op.Which(src, dst) {
		case SkipAudio:
			return p.op.Ok(path)
		case CopyAudio:
			return p.op.CopyFile(src.AbsPath(), path)
		case TranscodeAudio:
			p.wg.Add(1)
			p.pool.SendWorkAsync(func() {
				err := p.op.Transcode(src.AbsPath(), path, src)
				if err != nil {
					p.errs <- err
				}
				p.wg.Done()
			}, nil)
			return nil
		case UpdateAudio:
			p.wg.Add(1)
			p.pool.SendWorkAsync(func() {
				err := p.op.Update(src.AbsPath(), path, dst)
				if err != nil {
					p.errs <- err
				}
				p.wg.Done()
			}, nil)
			return nil
		case IgnoreAudio:
			return p.op.Ignore(path)
		default:
			panic("unknown audio operation")
		}
	} else if src.IsIgnored() {
		return p.op.Ignore(path)
	} else {
		if p.IgnoreData != p.DataExcept[src.Filename()] && src.Filename() != p.CoverSource {
			// We land in here when:
			// - Ignore all data (true) and there is no exception (false)
			// - Do not ignore all data (false) and there is an exception (true)
			return p.op.Ignore(path)
		}

		if dst != nil && dst.FileInfo().ModTime().After(src.FileInfo().ModTime()) {
			return p.op.Ok(path)
		}

		if p.DownscaleCover && src.Filename() == p.CoverSource {
			return p.op.DownscaleCover(src.AbsPath(), path)
		}
		return p.op.CopyFile(src.AbsPath(), path)
	}
}

// dpath returns the absolute destination path, given the key.
func (p *Planner) dpath(key string) string {
	return filepath.Join(p.dst.Path(), key)
}

// dkey returns the destination key, which also takes into account whether the
// file should be transcoded or not.
func (p *Planner) dkey(src *Entry) string {
	if src.Filename() == p.CoverSource && p.CoverSource != p.CoverTarget {
		if p.CoverTarget == "" {
			return src.Key()
		}
		return strings.Replace(src.Key(), p.CoverSource, p.CoverTarget, 1)
	}

	if !src.IsMusic() {
		return src.Key()
	}

	ext := p.op.WhichExt(src)
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
	//debug
	if dst.parent == nil {
		panic("why?")
	}
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
