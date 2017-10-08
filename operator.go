// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package lackey

import (
	"os"

	"github.com/goulash/audio"
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
	IsExists() bool
	FileInfo() os.FileInfo
	Encoding() audio.Codec
	Metadata() audio.Metadata
}

type Operator interface {
	// WhichExt takes the source metadata and returns the
	// expected destination extension of the file, such as ".mp3".
	// If "" is returned, the extension remains unchanged.
	//
	// Notes:
	// - This determines the destination filename that is constructed.
	//   Because the Operator takes care of encoding, it does not
	//   determine the encoding process.
	WhichExt(src Audio) string

	// Which returns an audio operation that should be
	// performed, based on src and dst (possibly nil).
	//
	// Notes:
	// - The Operator has full freedom in this decision.
	// - Parameter src is not nil.
	// - If dst == nil, then the destination file does not exist.
	// - If IgnoreAudio is returned, then Operator.Ignore is called.
	// - If UpdateAudio is returned, then Operator.Update is called.
	// - If CopyAudio is returned, then Operator.CopyFile is called.
	Which(src, dst Audio) AudioOperation

	// Feedback
	Ok(dst string) error
	Ignore(dst string) error
	Error(err error) error
	Warn(err error) error

	// Operations
	RemoveDir(dst string) error
	CreateDir(dst string) error

	// RemoveFile removes a file from the destination.
	// This occurs primarily when there is no corresponding source file or directory.
	RemoveFile(dst string) error
	CopyFile(src, dst string) error
	Transcode(src, dst string, md Audio) error
	Update(src, dst string, md Audio) error
}
