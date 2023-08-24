package compare

import (
	"bytes"
	"io"
	"os"
)

const BUFFERSIZE = 4096

// BasicEqual compares file modification time and file size as returend by os.Stat
func BasicEqual(srcInfo, dstInfo os.FileInfo) bool {
	return !srcInfo.ModTime().After(dstInfo.ModTime()) || (srcInfo.Size() != dstInfo.Size())
}

// DeepEqual returns if two files are equal on a byte-level
func DeepEqual(src, dst string) (bool, error) {
	// sanity check: file sizes must be equal
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false, err
	}
	if srcInfo.Size() != dstInfo.Size() {
		return false, nil
	}

	// sizes match, igonre mtime: compare byte by byte
	source, err := os.Open(src)
	if err != nil {
		return false, err
	}
	defer source.Close()

	destination, err := os.Open(dst)
	if err != nil {
		return false, err
	}
	defer destination.Close()

	bufSrc := make([]byte, BUFFERSIZE)
	bufDst := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(bufSrc)
		if err != nil && err != io.EOF {
			return false, err
		}
		if n == 0 {
			// it is sufficient to only check the bytes read from source,
			// since we made sure the sizes of source and destination match.
			break
		}
		n, err = destination.Read(bufDst)
		if err != nil && err != io.EOF {
			return false, err
		}
		if !bytes.Equal(bufSrc[:n], bufDst[:n]) {
			return false, nil
		}
	}
	return true, nil
}
