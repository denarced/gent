// Package assfs contains filesystem assert functionality.
package assfs

import (
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// AssertFs contains filesystem operations with asserts.
type AssertFs struct {
	req *require.Assertions
	fs  *afero.Afero
}

// NewAssertFs is a ctor for AssertFs.
func NewAssertFs(req *require.Assertions, fs *afero.Afero) *AssertFs {
	return &AssertFs{req: req, fs: fs}
}

// WriteTextFile writes a text file and creates the directories.
func (v *AssertFs) WriteTextFile(filep, content, message string) {
	v.doWriteTextFile(filep, content, 0, message)
}

// WriteLargeTextFile creates directories and writes the content plus a megabyte.
func (v *AssertFs) WriteLargeTextFile(filep, content, message string) {
	v.doWriteTextFile(filep, content, 1024^2, message)
}

func (v *AssertFs) doWriteTextFile(filep, content string, n int, message string) {
	dirp := filepath.Dir(filep)
	v.MkdirAll(dirp, message)
	v.req.Nilf(
		v.fs.WriteFile(filep, []byte(content+strings.Repeat("0", n)), 0666),
		"write, filep: %s, message: %s",
		filep,
		message,
	)
}

// DirExists asserts that dirp exists.
func (v *AssertFs) DirExists(dirp, message string) {
	exists, err := v.fs.DirExists(dirp)
	v.req.Nilf(err, "dir exists, err, dirp: %s, message: %s", dirp, message)
	v.req.Truef(exists, "dir exists, dirp: %s, message: %s", dirp, message)
}

func (v *AssertFs) doExists(path, message string, shouldExist bool) {
	exists, err := v.fs.Exists(path)
	v.req.Nilf(err, "exists, path: %s, message: %s, error: %s", path, message, err)
	v.req.Equal(shouldExist, exists, "exists, path: %s, message: %s", path, message)
}

// Exists asserts that path exists.
func (v *AssertFs) Exists(path, message string) {
	v.doExists(path, message, true)
}

// NotExists assert that path doesn't exist.
func (v *AssertFs) NotExists(path, message string) {
	v.doExists(path, message, false)
}

// ReadLines reads lines of file.
func (v *AssertFs) ReadLines(filep, message string) []string {
	b, err := v.fs.ReadFile(filep)
	v.req.Nilf(err, "read lines, path: %s, message: %s", filep, message)
	if len(b) == 0 {
		return []string{}
	}
	return strings.Split(string(b), "\n")
}

// MkdirAll creates the dirp.
func (v *AssertFs) MkdirAll(dirp, message string) {
	err := v.fs.MkdirAll(dirp, 0700)
	v.req.Nilf(err, "mkdir, path: %s, message: %s, error: %s", dirp, message, err)
}

// Contains checks if the file contains content.
func (v *AssertFs) Contains(filep, content, message string) {
	actual := strings.Join(v.ReadLines(filep, message), "\n")
	v.req.Equalf(content, actual, "contains, path: %s, message: %s", filep, message)
}

// WriteBytes writes bytes to filep.
func (v *AssertFs) WriteBytes(filep string, b []byte) error {
	return v.fs.WriteFile(filep, b, 0600)
}
