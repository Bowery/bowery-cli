// Copyright 2013-2014 Bowery, Inc.
// Package tar implements routines to read directories into tar streams.
package tar

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Bowery/cli/db"
)

// Tar reads the given dir and writes contents to a tar stream.
func Tar(dir string) (io.Reader, error) {
	buf := &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(buf)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	ignores, err := db.GetIgnores(dir)
	if err != nil {
		return nil, err
	}

	walker := func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Paths should be relative to the dir.
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}

		// Skip ignore paths and any descendents.
		for _, ignore := range ignores {
			if ignore == p {
				if info.IsDir() {
					return filepath.SkipDir
				}

				return nil
			}
		}

		// Ignore directories.
		if info.IsDir() {
			return nil
		}

		// Generate header from info.
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = path.Join(strings.Split(rel, string(filepath.Separator))...)

		// Detect if the path is a symlink.
		isLink := false
		if header.Typeflag == tar.TypeSymlink {
			isLink = true
		}

		// Get the correct link target.
		if isLink {
			target, err := os.Readlink(p)
			if err != nil {
				return err
			}

			header.Linkname = path.Join(strings.Split(target, string(filepath.Separator))...)
		}

		// Write header.
		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		// Copy the contents to tar.
		if !isLink {
			file, err := os.Open(p)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = filepath.Walk(dir, walker)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

// Untar a compressed source reader to a destination directory.
func Untar(source io.Reader, destDir string) error {
	// Uncompress source.
	gzipReader, err := gzip.NewReader(source)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

	// Untar each path in the header.
	for {
		header, err := tarReader.Next()
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF || header == nil {
			break
		}

		err = writeTarHeader(header, tarReader, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeTarHeader writes the contents of a given tar header.
func writeTarHeader(header *tar.Header, reader *tar.Reader, root string) error {
	var err error
	path := filepath.Join(root, filepath.Join(strings.Split(header.Name, "/")...))
	if header.Linkname != "" {
		header.Linkname = filepath.Join(strings.Split(header.Linkname, "/")...)
	}

	switch header.Typeflag {
	// Regular file
	case tar.TypeReg, tar.TypeRegA:
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm|os.ModeDir)
		if err != nil {
			return err
		}

		dest, err := os.Create(path)
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = io.Copy(dest, reader)
		if err != nil {
			return err
		}

		return dest.Chmod(os.FileMode(header.Mode))
	// Hard link
	case tar.TypeLink:
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm|os.ModeDir)
		if err != nil {
			return err
		}

		err = os.Link(header.Linkname, path)
		if os.IsExist(err) {
			err = nil
		}

		return err
	// Soft link
	case tar.TypeSymlink:
		err = os.MkdirAll(filepath.Dir(path), os.ModePerm|os.ModeDir)
		if err != nil {
			return err
		}

		err = os.Symlink(header.Linkname, path)
		if os.IsExist(err) {
			err = nil
		}

		return err
	// Directory
	case tar.TypeDir:
		err = os.MkdirAll(path, os.ModePerm|os.ModeDir)
		if err != nil {
			return err
		}

		return os.Chmod(path, os.FileMode(header.Mode))
	}

	return nil
}
