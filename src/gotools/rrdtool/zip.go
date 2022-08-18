package rrdtool

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

const (
	Store uint16 = 0 // no compression
	// DeflateCustom uint16 = 16 // DEFLATE compressed // 自定义压缩
)

// IsZip checks to see if path is already a zip file
func IsZip(path string) bool {
	r, err := zip.OpenReader(path)
	if err == nil {
		r.Close()
		return true
	}
	return false
}

// Zip takes all the files (dirs) and zips them into path
func Zip(path string, dirs []string) (err error) {
	if IsZip(path) {
		return errors.New(path + " is already a zip file")
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	startoffset, err := f.Seek(0, os.SEEK_END)
	if err != nil {
		return
	}

	w := zip.NewWriter(f)
	w.SetOffset(startoffset)
	// 自定义压缩
	// w.RegisterCompressor(DeflateCustom, func(out io.Writer) (io.WriteCloser, error) {
	// 	return flate.NewWriter(out, flate.BestCompression)
	// })
	// 自定义解压缩
	// // func RegisterDecompressor(method uint16, dcomp Decompressor)

	for _, dir := range dirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			fh, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			fh.Name = path

			if info.IsDir() {
				fh.Name += "/"
			} else {
				fh.Method = zip.Deflate
				// fh.Method = DeflateCustom // 自定义压缩
			}

			p, err := w.CreateHeader(fh)
			if err != nil {
				return err
			}
			if !info.IsDir() {

				content, err := os.Open(path)
				if err != nil {
					return err
				}

				_, err = io.Copy(p, content)
				content.Close()
				if err != nil {
					return err
				}

			}
			return err
		})
	}
	err = w.Close()
	return
}

// Unzip unzips the file zippath and puts it in destination
func Unzip(zippath string, destination string) (err error) {
	r, err := zip.OpenReader(zippath)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		fullname := path.Join(destination, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fullname, f.FileInfo().Mode().Perm())
		} else {
			os.MkdirAll(filepath.Dir(fullname), 0755)
			perms := f.FileInfo().Mode().Perm()
			out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
			if err != nil {
				return err
			}
			rc, err := f.Open()
			if err != nil {
				return err
			}
			_, err = io.CopyN(out, rc, f.FileInfo().Size())
			if err != nil {
				return err
			}
			rc.Close()
			out.Close()

			mtime := f.FileInfo().ModTime()
			err = os.Chtimes(fullname, mtime, mtime)
			if err != nil {
				return err
			}
		}
	}
	return
}

// UnzipList Lists all the files in zip file
func UnzipList(path string) (list []string, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	for _, f := range r.File {
		size := f.FileInfo().Size()
		mtime := f.FileInfo().ModTime()
		metadata := fmt.Sprintf("%v\t\t %v\t\t %v", size, mtime, f.Name)
		list = append(list, metadata)
	}
	return
}
