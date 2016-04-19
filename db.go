// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import "path/filepath"

type EntryType int

const (
	DirEntry EntryType = 1 << iota
	FileEntry
	MusicEntry

	UnknownEntry EntryType = 0
)

type Entry struct {
	db       *Database // pointer to database (root)
	parent   *Entry
	children []*Entry

	typ   EntryType   // entry type (dir|file|music)
	path  string      // path relative to database root
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

func (e *Entry) Type() EntryType {
	return e.typ
}

func (e *Entry) Filename() string {
	return filepath.Base(e.path)
}

func (e *Entry) FilenameExt() (base string, ext string) {
	name := filepath.Base(e.path)
	ext := filepath.Ext(e.path)
	base := name[:len(base)-len(ext)]
	return base, ext
}

func (e *Entry) RelPath() string {
	return e.path
}

func (e *Entry) AbsPath() string {
	return filepath.Join(e.db.path, e.path)
}

func (e *Entry) RootPath() string {
	return e.db.path
}

type Database struct {
	path    string
	root    *Entry
	entries map[string]*Entry
}

func (db *Database) Size() int64 {
	return db.root.Size()
}

// Get returns the entry that has the given key, or nil.
func (db *Database) Get(key string) *Entry {
	return db.entries[key]
}
