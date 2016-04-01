// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/Jeffail/tunny"
	"github.com/goulash/color"
	"github.com/goulash/errs"
	"github.com/goulash/osutil"
)

var col = color.New()

type Walker struct {
	SourceMaxBitrate int
	TargetCodec      Codec
	TargetQuality    int
	LinkFiles        bool
	OnlyMusic        bool
	Parallel         int
}

func (w *Walker) Walk(src, dst string, h errs.Handler) error {
	pool, err := tunny.CreatePoolGeneric(w.Parallel).Open()
	if err != nil {
		return err
	}

	wlk := walker{
		opt:  w,
		src:  src,
		dst:  dst,
		uhoh: h,
		pool: pool,
	}
	return wlk.Walk()
}

type walker struct {
	opt *Walker
	src string
	dst string

	uhoh  errs.Handler
	pool  *tunny.WorkPool
	mutex sync.RWMutex
	wg    sync.WaitGroup
	err   error
}

func (w *walker) IsError() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.err != nil
}

func (w *walker) Walk() error {
	err := filepath.Walk(w.src, func(path string, fi os.FileInfo, err error) error {
		if w.IsError() {
			return w.err
		}
		if err != nil {
			return w.uhoh(err)
		}

		relpath, err := filepath.Rel(w.src, path)
		if err != nil {
			return w.uhoh(err)
		}

		if fi.IsDir() {
			err = w.doDir(relpath)
			if err != nil {
				return w.uhoh(err)
			}
		} else {
			err = w.doFile(relpath)
			if err != nil {
				return w.uhoh(err)
			}
		}

		return nil
	})
	w.wg.Wait()
	w.pool.Close()
	return err
}

func (w *walker) doDir(path string) error {
	return os.MkdirAll(filepath.Join(w.dst, path), 0777)
}

func (w *walker) doFile(path string) error {
	fi, err := ReadFileInfo(filepath.Join(w.src, path))
	if err != nil {
		return err
	}

	if fi.IsMusic() {
		return w.doMusic(path)
	}

	if w.opt.OnlyMusic {
		return nil
	}

	return w.copyFile(path)
}

func (w *walker) copyFile(path string) error {
	src := filepath.Join(w.src, path)
	dst := filepath.Join(w.dst, path)
	if w.opt.LinkFiles {
		col.Println("@gCopying   \t", path)
		return osutil.CopyFileLazy(src, dst)
	}
	col.Println("@yCopying   \t", path)
	return osutil.CopyFile(src, dst)
}

func (w *walker) doMusic(path string) error {
	src := filepath.Join(w.src, path)
	fi, err := ReadMusicInfo(src)
	if err != nil {
		return err
	}

	if fi.Encoding() == MP3 && fi.Bitrate() < w.opt.SourceMaxBitrate {
		return w.copyFile(path)
	}

	switch w.opt.TargetCodec {
	case MP3, OGG, AAC, M4A:
	case FLAC:
		// There are use-cases where this would come in handy, converting
		// from one lossless format to another. But it seems ultimately
		// pointless, so we return an error, it probably is one, and that
		// isn't really the point of this program.
		return errors.New("you may not convert to FLAC type")
	default:
		return errors.New("unknown target codec")
	}

	// Replace the current extension with the new one
	ext := w.opt.TargetCodec.Ext()
	ne := len(filepath.Ext(path))
	dst := filepath.Join(w.dst, path[:len(path)-ne]+ext)

	w.wg.Add(1)
	w.pool.SendWorkAsync(func() {
		if w.IsError() {
			w.wg.Done()
			return
		}

		col.Println("@rConverting\t", path)
		err := ConvertMusicFile(src, dst, w.opt.TargetQuality)
		if err != nil {
			err = w.uhoh(err)
			if err != nil {
				w.mutex.Lock()
				if w.err == nil {
					w.err = err
				}
				w.mutex.Unlock()
			}
		}
		w.wg.Done()
	}, nil)
	return nil
}
