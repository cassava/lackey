// Copyright 2016 Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package flac

import (
	"io"
	"unsafe"
)

// TODO: We should see if this is all really as performant as
// I think it is...

func readBytes(r io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	n, err := r.Read(buf)
	if n != n || err != nil {
		return nil, ErrUnexpectedEOF
	}
	return buf, nil
}

func readString(r io.Reader, n int) (string, error) {
	buf := make([]byte, n)
	rn, err := r.Read(buf)
	if rn != n || err != nil {
		return "", ErrUnexpectedEOF
	}
	return string(buf), nil
}

func readUint8(r io.Reader) (uint8, error) {
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if n != 1 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	return uint8(buf[0]), nil
}

func readUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	n, err := r.Read(buf)
	if n != 2 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	buf[0], buf[1] = buf[1], buf[0]
	return *(*uint16)(unsafe.Pointer(&buf[0])), nil
}

func readUint24(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	n, err := r.Read(buf[1:])
	if n != 3 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	buf[0], buf[1], buf[2], buf[3] = buf[3], buf[2], buf[1], buf[0]
	return *(*uint32)(unsafe.Pointer(&buf[0])), nil
}

func readUint32(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	n, err := r.Read(buf)
	if n != 4 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	buf[0], buf[1], buf[2], buf[3] = buf[3], buf[2], buf[1], buf[0]
	return *(*uint32)(unsafe.Pointer(&buf[0])), nil
}

func readUint48(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	n, err := r.Read(buf[2:])
	if n != 6 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7] = buf[7], buf[6], buf[5], buf[4], buf[3], buf[2], buf[1], buf[0]
	return *(*uint64)(unsafe.Pointer(&buf[0])), nil
}

func readUint64(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if n != 8 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	buf[0], buf[1], buf[2], buf[3], buf[4], buf[5], buf[6], buf[7] = buf[7], buf[6], buf[5], buf[4], buf[3], buf[2], buf[1], buf[0]
	return *(*uint64)(unsafe.Pointer(&buf[0])), nil
}

func readUint16LE(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	n, err := r.Read(buf)
	if n != 2 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	return *(*uint16)(unsafe.Pointer(&buf[0])), nil
}

func readUint24LE(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	n, err := r.Read(buf[1:])
	if n != 3 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	return *(*uint32)(unsafe.Pointer(&buf[0])), nil
}

func readUint32LE(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	n, err := r.Read(buf)
	if n != 4 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	return *(*uint32)(unsafe.Pointer(&buf[0])), nil
}

func readUint48LE(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	n, err := r.Read(buf[2:])
	if n != 6 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	return *(*uint64)(unsafe.Pointer(&buf[0])), nil
}

func readUint64LE(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if n != 8 || err != nil {
		return 0, ErrUnexpectedEOF
	}
	return *(*uint64)(unsafe.Pointer(&buf[0])), nil
}
