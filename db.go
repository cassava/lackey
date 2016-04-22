// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cassava/lackey/music"
)

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
	typ   EntryType   // entry type (dir|file|music)
	bytes int64       // cumulative size of entry
	data  interface{} // any extra data stored with this entry
}

func (e *Entry) Key() string {
	return e.path
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

func (e *Entry) Data() interface{} {
	return e.data
}

// Size returns the size of the entry in bytes.
// A directory's size is the size of all its children.
func (e *Entry) Size() int64 {
	return e.bytes
}

func (e *Entry) SizeString() string {
	const (
		kB = 1000
		MB = 1000 * 1000
		GB = 1000 * 1000 * 1000
		TB = 1000 * 1000 * 1000 * 1000
	)

	z := float64(e.bytes)
	if z < kB {
		return fmt.Sprintf("%d B", e.bytes)
	} else if z < MB {
		return fmt.Sprintf("%.3f kB", z/kB)
	} else if z < GB {
		return fmt.Sprintf("%.3f MB", z/MB)
	} else if z < TB {
		return fmt.Sprintf("%.3f GB", z/GB)
	} else {
		return fmt.Sprintf("%.3f TB", z/TB)
	}
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
	base = name[:len(base)-len(ext)]
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
}

func (db *Database) Path() string {
	return db.path
}

func (db *Database) Size() int64 {
	return db.root.Size()
}

func (db *Database) SizeString() string {
	return db.root.SizeString()
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
		filepath.Walk(abs, func(path string, fi os.FileInfo, err error) error {
			if path == abs {
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

	if c, err := music.Identify(abs); c == music.Unknown {
		e.typ = FileEntry
		e.data = err
	}

	e.bytes = fi.Size()
	e.typ = MusicEntry
	m, err := music.ReadMetadata(abs)
	if err != nil {
		e.data = err
		return
	}
	e.data = m
}

func ReadLibrary(path string) (*Database, error) {
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
		path:    abs,
		entries: make(map[string]*Entry),
	}
	db.root = &Entry{db: db}
	db.root.init(".", fi, nil)
	return db, nil
}

// }}}
