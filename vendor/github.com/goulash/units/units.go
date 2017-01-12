// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package units

import "fmt"

const (
	KB = 1000
	MB = 1000 * 1000
	GB = 1000 * 1000 * 1000
	TB = 1000 * 1000 * 1000 * 1000
	PB = 1000 * 1000 * 1000 * 1000 * 1000

	KiB = 1024
	MiB = 1024 * 1024
	GiB = 1024 * 1024 * 1024
	TiB = 1024 * 1024 * 1024 * 1024
	PiB = 1024 * 1024 * 1024 * 1024 * 1024
)

func Bytes10(sz int64) string {
	z := float64(sz)
	if z < KB {
		return fmt.Sprintf("%d B", sz)
	} else if z < MB {
		return fmt.Sprintf("%.3f kB", z/KB)
	} else if z < GB {
		return fmt.Sprintf("%.3f MB", z/MB)
	} else if z < TB {
		return fmt.Sprintf("%.3f GB", z/GB)
	} else if z < PB {
		return fmt.Sprintf("%.3f TB", z/TB)
	} else {
		return fmt.Sprintf("%.3f PB", z/PB)
	}
}

func Bytes2(sz int64) string {
	z := float64(sz)
	if z < KiB {
		return fmt.Sprintf("%d B", sz)
	} else if z < MiB {
		return fmt.Sprintf("%.3f kB", z/KiB)
	} else if z < GiB {
		return fmt.Sprintf("%.3f MB", z/MiB)
	} else if z < TiB {
		return fmt.Sprintf("%.3f GB", z/GiB)
	} else if z < PiB {
		return fmt.Sprintf("%.3f TB", z/TiB)
	} else {
		return fmt.Sprintf("%.3f PB", z/PiB)
	}
}
