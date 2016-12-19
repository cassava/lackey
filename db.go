// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/facebookgo/symwalk"
	"github.com/goulash/audio"

	_ "github.com/cassava/lackey/audio/mp3"
	_ "github.com/goulash/audio/flac"
)

var Skip = errors.New("skip the directory or remaining files in directory")

// Entry and EntryType {{{

type EntryType int

const (
	DirEntry EntryType = 1 << iota
	FileEntry
	MusicEntry

	ErrorEntry EntryType = 0
)

type Entry struct {
	db       *Database // pointer to database (root)
	parent   *Entry
	children []*Entry

	path  string // path relative to database root
	fi    os.FileInfo
	typ   EntryType // entry type (dir|file|music)
	bytes int64     // cumulative size of entry
	codec audio.Codec
	data  interface{} // any extra data stored with this entry
}

func (e *Entry) Key() string {
	return e.path
}

func (e *Entry) FileInfo() os.FileInfo {
	return e.fi
}

func (e *Entry) IsExists() bool {
	return e != nil && e.fi != nil
}

func (e *Entry) IsDir() bool {
	return e.typ == DirEntry
}

func (e *Entry) IsMusic() bool {
	return e.typ == MusicEntry
}

func (e *Entry) Parent() *Entry {
	return e.parent
}

func (e *Entry) Children() []*Entry {
	return e.children
}

// Care should be taken with when Data is called, as the first call may involve
// reading the metadata of the associated file.
func (e *Entry) Data() interface{} {
	if e.data != nil {
		return e.data
	}

	// Get this data in a lazy fashion
	if e.typ == MusicEntry {
		abs := filepath.Join(e.db.Path(), e.path)
		m, err := audio.ReadMetadata(abs)
		if err != nil {
			e.data = err
			return e.data
		}
		e.data = m
	}

	return e.data
}

func (e *Entry) Encoding() audio.Codec {
	return e.codec
}

func (e *Entry) Metadata() audio.Metadata {
	md, ok := e.Data().(audio.Metadata)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", e.Key(), e.Data())
		return nil
	}
	return md
}

func (e *Entry) Walk(fn func(e *Entry) error) error {
	if e.IsDir() {
		err := fn(e)
		if err != nil {
			if err == Skip {
				return nil
			}
			return err
		}

		for _, c := range e.children {
			err = c.Walk(fn)
			if err != nil {
				if err == Skip {
					return nil
				}
				return err
			}
		}
		return nil
	} else {
		return fn(e)
	}
}

// Size returns the size of the entry in bytes.
// A directory's size is the size of all its children.
func (e *Entry) Size() int64 {
	return e.bytes
}

func (e *Entry) Type() EntryType {
	return e.typ
}

func (e *Entry) Filename() string {
	return filepath.Base(e.path)
}

func (e *Entry) FilenameExt() (base string, ext string) {
	name := filepath.Base(e.path)
	ext = filepath.Ext(e.path)
	base = name[:len(name)-len(ext)]
	return base, ext
}

func (e *Entry) RelPath() string {
	return e.path
}

func (e *Entry) AbsPath() string {
	return filepath.Join(e.db.Path(), e.path)
}

func (e *Entry) RootPath() string {
	return e.db.Path()
}

// }}}

// Database {{{

type Database struct {
	path    string
	root    *Entry
	entries map[string]*Entry

	// Options
	ignoreHidden   bool
	followSymlinks bool
	walker         func(string, filepath.WalkFunc) error
}

func (db *Database) Path() string {
	return db.path
}

func (db *Database) Size() int64 {
	return db.root.Size()
}

func (db *Database) Root() *Entry {
	return db.root
}

// Get returns the entry that has the given key, or nil.
func (db *Database) Get(key string) *Entry {
	return db.entries[key]
}

func (db *Database) Set(key string, e *Entry) {
	db.entries[key] = e
}

func (db *Database) Walk(fn func(e *Entry) error) error {
	return db.root.Walk(fn)
}

// init populates the entry with all the relevant informations.
// It is expected that e.parent and e.db are already set.
func (e *Entry) init(path string, fi os.FileInfo, err error) {
	defer e.db.Set(path, e)

	root := e.db.Path()
	abs := filepath.Join(root, path)
	e.path = path
	e.fi = fi

	if err != nil {
		e.typ = ErrorEntry
		e.data = err
		return
	}

	if fi.IsDir() {
		e.typ = DirEntry
		e.db.walker(abs, func(path string, fi os.FileInfo, err error) error {
			if path == abs {
				return nil
			}

			// Ignore hidden files if requested
			if e.db.ignoreHidden && filepath.HasPrefix(filepath.Base(path), ".") {
				return nil
			}

			path, _ = filepath.Rel(root, path)
			v := &Entry{
				db:     e.db,
				parent: e,
			}
			v.init(path, fi, err)
			e.children = append(e.children, v)
			e.bytes += v.bytes

			// filepath.Walk should not recurse, because v.init does that already.
			if fi != nil && fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		})
		return
	}

	e.bytes = fi.Size()
	e.codec, err = audio.Identify(abs)
	if e.codec == audio.Unknown {
		e.typ = FileEntry
		e.data = err
	} else {
		e.typ = MusicEntry
		// We don't set e.data now, rather when Data is called.
	}
}

type LibraryReader struct {
	FollowSymlinks bool
	IgnoreHidden   bool
}

func (r LibraryReader) ReadLibrary(path string) (*Database, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, errors.New("library path must be a directory")
	}

	db := &Database{
		path:           abs,
		entries:        make(map[string]*Entry),
		followSymlinks: r.FollowSymlinks,
		ignoreHidden:   r.IgnoreHidden,
		walker:         filepath.Walk,
	}
	if r.FollowSymlinks {
		db.walker = symwalk.Walk
	}
	db.root = &Entry{db: db}
	db.root.init(".", fi, nil)
	return db, nil
}

// ReadLibrary reads a library using the recommended options, namely:
//
//  FollowSymlinks = true
//  IgnoreHidden   = true
func ReadLibrary(path string) (*Database, error) {
	r := LibraryReader{
		FollowSymlinks: true,
		IgnoreHidden:   true,
	}
	return r.ReadLibrary(path)
}

// }}}
