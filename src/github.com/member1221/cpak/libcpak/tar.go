/*
Modified code from archiver by mholt. https://github.com/mholt/archiver
MIT License

Copyright (c) 2016 Matthew Holt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.*/

package libcpak

import (
	"os"
	"io"
	"bytes"
	"strconv"
	"archive/tar"
	"fmt"
	"path/filepath"
	"strings"
	"runtime"
)

const tarBlockSize int = 512

// isTar checks the file has the Tar format header by reading its beginning
// block.
func IsTar(tarPath string) bool {
	f, err := os.Open(tarPath)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, tarBlockSize)
	if _, err = io.ReadFull(f, buf); err != nil {
		return false
	}

	return hasTarHeader(buf)

}

// hasTarHeader checks passed bytes has a valid tar header or not. buf must
// contain at least 512 bytes and if not, it always returns false.
func hasTarHeader(buf []byte) bool {
	if len(buf) < tarBlockSize {
		return false
	}

	b := buf[148:156]
	b = bytes.Trim(b, " \x00") // clean up all spaces and null bytes
	if len(b) == 0 {
		return false // unknown format
	}
	hdrSum, err := strconv.ParseUint(string(b), 8, 64)
	if err != nil {
		return false
	}

	// According to the go official archive/tar, Sun tar uses signed byte
	// values so this calcs both signed and unsigned
	var usum uint64
	var sum int64
	for i, c := range buf {
		if 148 <= i && i < 156 {
			c = ' ' // checksum field itself is counted as branks
		}
		usum += uint64(uint8(c))
		sum += int64(int8(c))
	}

	if hdrSum != usum && int64(hdrSum) != sum {
		return false // invalid checksum
	}

	return true
}

// tarball writes all files listed in filePaths into tarWriter, which is
// writing into a file located at dest.
func Tarball(filePaths []string, tarWriter *tar.Writer, dest string) error {
	for _, fpath := range filePaths {
		err := TarFile(tarWriter, fpath, dest)
		if err != nil {
			return err
		}
	}
	return nil
}

// tarFile writes the file at source into tarWriter. It does so
// recursively for directories.
func TarFile(tarWriter *tar.Writer, source, dest string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("%s: stat: %v", source, err)
	}

	var baseDir string
	if sourceInfo.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking to %s: %v", path, err)
		}

		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return fmt.Errorf("%s: making header: %v", path, err)
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if header.Name == dest {
			// our new tar file is inside the directory being archived; skip it
			return nil
		}

		if info.IsDir() {
			header.Name += "/"
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("%s: writing header: %v", path, err)
		}

		if info.IsDir() {
			return nil
		}

		if header.Typeflag == tar.TypeReg {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("%s: open: %v", path, err)
			}
			defer file.Close()

			_, err = io.CopyN(tarWriter, file, info.Size())
			if err != nil && err != io.EOF {
				return fmt.Errorf("%s: copying contents: %v", path, err)
			}
		}
		return nil
	})
}

// untar un-tarballs the contents of tr into destination.
func Untar(tr *tar.Reader, destination string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := UntarFile(tr, header, destination); err != nil {
			return err
		}
	}
	return nil
}

// untarFile untars a single file from tr with header header into destination.
func UntarFile(tr *tar.Reader, header *tar.Header, destination string) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return mkdir(filepath.Join(destination, header.Name))
	case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
		return writeNewFile(filepath.Join(destination, header.Name), tr, header.FileInfo().Mode())
	case tar.TypeSymlink:
		return writeNewSymbolicLink(filepath.Join(destination, header.Name), header.Linkname)
	case tar.TypeLink:
		return writeNewHardLink(filepath.Join(destination, header.Name), filepath.Join(destination, header.Linkname))
	default:
		return fmt.Errorf("%s: unknown type flag: %c", header.Name, header.Typeflag)
	}
}

func mkdir(dir string) error {
	return os.Mkdir(dir, 0755)
}

func writeNewFile(fpath string, in io.Reader, fm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer out.Close()

	err = out.Chmod(fm)
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("%s: changing file mode: %v", fpath, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

func writeNewSymbolicLink(fpath string, target string) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	err = os.Symlink(target, fpath)
	if err != nil {
		return fmt.Errorf("%s: making symbolic link for: %v", fpath, err)
	}

	return nil
}

func writeNewHardLink(fpath string, target string) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	err = os.Link(target, fpath)
	if err != nil {
		return fmt.Errorf("%s: making hard link for: %v", fpath, err)
	}

	return nil
}



