package zip

import (
	"archive/zip"
	"io"
	"os"
)

// WriteEntry writes a single file (src) into a zip archive on w,
// storing it under the given name inside the archive.
func WriteEntry(w io.Writer, name, src string) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	fh, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	fh.Name = name
	fh.Method = zip.Deflate

	entry, err := zw.CreateHeader(fh)
	if err != nil {
		return err
	}

	_, err = io.Copy(entry, f)
	return err
}
