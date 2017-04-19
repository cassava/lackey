// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

// Package filetype categorizes files based on their filetype,
// with the exception of directories, which it identifies directly.
//
// Among others, the following files were found in a directory of
// media files.
//
//   err    -> Unknown
//   flac   -> Audio
//   htm    -> Text
//   jpg    -> Image
//   lyrics -> Text
//   m4a    -> Audio
//   mkv    -> Video
//   mov    -> Video
//   mp3    -> Audio
//   mpg    -> Video
//   pdf    -> Text
//   png    -> Image
//   stfolder   -> Unknown
//   svg    -> Image
//   wav    -> Audio
//
package filetype

import (
	"path/filepath"

	"github.com/goulash/osutil"
)

type Type int

const (
	Directory Type = 1 << iota
	Text
	Image
	Audio
	Video
	Archive

	SourceCode
	Backup
	Binary

	Unknown Type = 0
	Error   Type = -1
	Ignored Type = -2
)

var types = map[string]Type{
	".txt":      Text,
	".md":       Text,
	".markdown": Text,
	".lyrics":   Text,
	".pdf":      Text,
	".chords":   Text,
	".chr":      Text,
	".tex":      Text,
	".html":     Text,
	".htm":      Text,

	".7z":   Archive,
	".Z":    Archive,
	".ace":  Archive,
	".arj":  Archive,
	".bz":   Archive,
	".bz2":  Archive,
	".cpio": Archive,
	".deb":  Archive,
	".dz":   Archive,
	".gz":   Archive,
	".jar":  Archive,
	".lzh":  Archive,
	".lzma": Archive,
	".rar":  Archive,
	".rpm":  Archive,
	".rz":   Archive,
	".svgz": Archive,
	".tar":  Archive,
	".taz":  Archive,
	".tbz2": Archive,
	".tgz":  Archive,
	".tz":   Archive,
	".xz":   Archive,
	".z":    Archive,
	".zip":  Archive,
	".zoo":  Archive,

	".asf":  Image,
	".bmp":  Image,
	".dl":   Image,
	".flc":  Image,
	".fli":  Image,
	".gif":  Image,
	".gl":   Image,
	".ico":  Image,
	".jpeg": Image,
	".jpg":  Image,
	".mng":  Image,
	".nuv":  Image,
	".ogm":  Image,
	".pbm":  Image,
	".pcx":  Image,
	".pgm":  Image,
	".png":  Image,
	".ppm":  Image,
	".qt":   Image,
	".rm":   Image,
	".rmvb": Image,
	".svg":  Image,
	".tga":  Image,
	".tif":  Image,
	".tiff": Image,
	".xbm":  Image,
	".xcf":  Image,
	".xpm":  Image,
	".xwd":  Image,
	".yuv":  Image,

	".avi":  Video,
	".dv":   Video,
	".flv":  Video,
	".m2v":  Video,
	".m4v":  Video,
	".mkv":  Video,
	".mov":  Video,
	".mp4v": Video,
	".mpeg": Video,
	".mpg":  Video,
	".mp4":  Video,
	".vob":  Video,
	".webm": Video,
	".wmv":  Video,

	".aax":  Audio,
	".aac":  Audio,
	".au":   Audio,
	".flac": Audio,
	".mid":  Audio,
	".midi": Audio,
	".mka":  Audio,
	".mp3":  Audio,
	".m4b":  Audio,
	".mpc":  Audio,
	".ogg":  Audio,
	".ra":   Audio,
	".wav":  Audio,

	".d":     SourceCode,
	".c":     SourceCode,
	".cc":    SourceCode,
	".cpp":   SourceCode,
	".cxx":   SourceCode,
	".i":     SourceCode,
	".ii":    SourceCode,
	".ipp":   SourceCode,
	".ixx":   SourceCode,
	".h":     SourceCode,
	".hh":    SourceCode,
	".hpp":   SourceCode,
	".hxx":   SourceCode,
	".py":    SourceCode,
	".sh":    SourceCode,
	".bash":  SourceCode,
	".zsh":   SourceCode,
	".vim":   SourceCode,
	".hs":    SourceCode,
	".pl":    SourceCode,
	".go":    SourceCode,
	".php":   SourceCode,
	".js":    SourceCode,
	".css":   SourceCode,
	".scss":  SourceCode,
	".inc":   SourceCode,
	".xml":   SourceCode,
	".sql":   SourceCode,
	".pgsql": SourceCode,

	".a":   Binary,
	".ko":  Binary,
	".o":   Binary,
	".gch": Binary,
	".so":  Binary,
	".pyc": Binary,
	".pyo": Binary,
	".s":   Binary,
	".zwc": Binary,

	".tmp":  Backup,
	".temp": Backup,
	".bak":  Backup,
	".swp":  Backup,
}

func Identify(path string) Type {
	if ex, _ := osutil.Exists(path); !ex {
		return Error
	} else if ex, _ := osutil.DirExists(path); ex {
		return Directory
	}

	ext := filepath.Ext(path)
	return types[ext]
}
