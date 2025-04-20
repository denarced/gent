// Package gent contains miscellaneous tools that the author keeps rewriting in various projects.
package gent

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

var (
	nonSafeFilenamePattern = regexp.MustCompile(`[^0-9a-zA-Z-._]`)
)

// Pair is a pair of values.
type Pair[T any, U any] struct {
	First  T
	Second U
}

// NewPair create an initialized [gent.Pair].
func NewPair[T any, U any](first T, second U) Pair[T, U] {
	return Pair[T, U]{First: first, Second: second}
}

// Set is a naive map backed set.
type Set[T comparable] struct {
	m map[T]bool
}

// NewSet creates a new [gent.Set].
func NewSet[T comparable](items ...T) *Set[T] {
	set := &Set[T]{m: map[T]bool{}}
	for _, each := range items {
		set.Add(each)
	}
	return set
}

// Add item to the set, return true if it was added.
// Otherwise it already existed and wasn't added.
func (v *Set[T]) Add(item T) (added bool) {
	_, existed := v.m[item]
	if existed {
		return
	}
	added = true
	v.m[item] = false
	return
}

// Clear the set, remove all items.
func (v *Set[T]) Clear() {
	v.m = map[T]bool{}
}

// Equal returns true when the sets contain the exact same items.
func (v *Set[T]) Equal(s *Set[T]) bool {
	if len(v.m) != s.Len() {
		return false
	}
	for each := range v.m {
		if !s.Has(each) {
			return false
		}
	}
	return true
}

// Contains checks if item exists in the set.
// Alias for [gent.Set.Has].
func (v *Set[T]) Contains(item T) bool {
	return v.Has(item)
}

// Has checks if item exists in the set.
func (v *Set[T]) Has(item T) bool {
	_, ok := v.m[item]
	return ok
}

// ForEach iterates all items in the set, calls f for each item, stops if stop is called.
// Use [gent.ForEachAll] if there's no need to stop iteration.
func (v *Set[T]) ForEach(f func(each T, stop func())) {
	breaker := false
	for each := range v.m {
		f(each, func() {
			breaker = true
		})
		if breaker {
			break
		}
	}
}

// ForEachAll iterates all items in the set and calls f for each item.
// Use [gent.ForEach] if you need to stop iteration.
func (v *Set[T]) ForEachAll(f func(each T)) {
	for key := range v.m {
		f(key)
	}
}

// Len returns the number of items in the set.
func (v *Set[T]) Len() int {
	return len(v.m)
}

// Count returns the number of items in the set.
// Alias for [gent.Set.Len].
func (v *Set[T]) Count() int {
	return v.Len()
}

// Remove removes an item in the set, returns true if it was.
// I.e. if it existed.
func (v *Set[T]) Remove(item T) (existed bool) {
	_, existed = v.m[item]
	delete(v.m, item)
	return
}

// ToSlice returns a slice with all set items.
// Set itself doesn't change.
func (v *Set[T]) ToSlice() []T {
	keys := []T{}
	for each := range v.m {
		keys = append(keys, each)
	}
	return keys
}

// ReadLines read all lines in file filep.
// Empty lines are included.
// Returned lines do not contain newlines at the end.
func ReadLines(filep string) (lines []string, err error) {
	var f *os.File
	if f, err = os.Open(filep); err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return
}

// Tri returns one of the two values based on the condition.
// I.e. this is a ternary "operator".
func Tri[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}

// Map a slice into another slice of the same size.
func Map[T any, U any](s []T, f func(T) U) []U {
	mapped := make([]U, len(s))
	for i, v := range s {
		mapped[i] = f(v)
	}
	return mapped
}

// Filter values in s with f.
// When f returns true, item is included in the response slice.
func Filter[T any](s []T, f func(T) bool) []T {
	var filtered []T
	for _, v := range s {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// OrPanic2 returns function that returns value if err is nil, else panics with message.
// Useful for cases where failure should result in panic
// and you don't want to deal with the returned error.
func OrPanic2[T any](value T, err error) func(message string) T {
	if err == nil {
		return func(_ string) T {
			return value
		}
	}
	return func(message string) T {
		panic(fmt.Sprintf("Message: %s. Error: %s.", message, err))
	}
}

// A SnapshotSuite is a suite of snapshot tests with a shared directory for the snapshot files.
// It is made of [gent.Snapshot]s.
type SnapshotSuite struct {
	rootDir string
}

// NewSnapshotSuite creates a [gent.SnapshotSuite] with a root directory.
// Usually it's under "testdata".
func NewSnapshotSuite(rootDir string) *SnapshotSuite {
	return &SnapshotSuite{rootDir: rootDir}
}

// VerifyFunc is used to assert that snapshot matches to the string that code produced.
// This is your standard "assertEqual" function in any unit test library.
type VerifyFunc func(expected, actual, message string)

// Snapshot represents a single test with a snapshot file.
type Snapshot struct {
	// Name of the test that's also the last part of the snapshot file's filepath.
	Name   string
	filep  string
	verify bool
	equal  VerifyFunc
}

// NewSnapshot creates a snapshot.
// Name is [gent.Snapshot.Name] and with [gent.SnapshotSuite.rootDir],
// becomes the full filepath of the snapshot file.
// When verify is false, snapshots are written, and tests won't fail.
// That's how you initialize or update snapshots.
// When verify is true and snapshot file doesn't exist or it's empty,
// content produced by the tested code is written.
// And finally, when verify is true and the snapshot file exists,
// equal function is used to assert equality.
func (v *SnapshotSuite) NewSnapshot(name string, verify bool, equal VerifyFunc) *Snapshot {
	return &Snapshot{
		Name:   name,
		filep:  v.deriveSnapshotFilep(name),
		verify: verify,
		equal:  equal,
	}
}

func (v *SnapshotSuite) deriveSnapshotFilep(name string) string {
	return filepath.Join(v.rootDir, name)
}

func (v *Snapshot) read() (string, error) {
	b, err := os.ReadFile(v.filep)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

func (v *Snapshot) write(content string) error {
	return os.WriteFile(v.filep, []byte(content), 0644)
}

// Run the snapshot process according to parameters set in [gent.SnapshotSuite.NewSnapshot].
// Error is returned when something unexpected fails, not when the test itself fails.
// Determining whether any given test fails
// is left for "equal" function defined in [gent.SnapshotSuite.NewSnapshot].
func (v *Snapshot) Run(view string) error {
	content, err := v.read()
	if err != nil {
		return err
	}
	if v.verify && content != "" {
		v.equal(content, view, v.Name)
		return nil
	}
	if view != content {
		return v.write(view)
	}
	return nil
}

// ToSafeFilename replaces all non-safe characters with underscore.
func ToSafeFilename(s string) string {
	return nonSafeFilenamePattern.ReplaceAllString(s, "_")
}

// RunBubbleTeaSnapshots runs snapshots for bubbletea TUIs.
func RunBubbleTeaSnapshots(
	snapshotSuite *SnapshotSuite,
	m tea.Model,
	verify bool,
	seriesID string,
	equal VerifyFunc,
) {
	runSnapshot := func(i int) {
		snapshot := snapshotSuite.NewSnapshot(
			fmt.Sprintf("%s_%03d", seriesID, i),
			verify,
			equal)
		if err := snapshot.Run(m.View()); err != nil {
			panic(err)
		}
	}
	messageGroups := readMessageGroups(snapshotSuite.rootDir, seriesID)
	// Quick test elsewhere showed that normal run does init, view, update, and view.
	cmd := m.Init()
	m.View()
	m = runUpdates(m, cmd)
	runSnapshot(0)

	for i, group := range messageGroups {
		for _, each := range group {
			m = runUpdates(m, createKey(each))
		}
		runSnapshot(i + 1)
	}
}

func runUpdates(m tea.Model, msg tea.Msg) tea.Model {
	var cmd tea.Cmd
	m, cmd = m.Update(msg)
	counter := 100
	for cmd != nil {
		m, cmd = m.Update(cmd())
		counter--
		if counter <= 0 {
			panic("counter == 0, eternal loop")
		}
	}
	return m
}

func readMessageGroups(snapshotRootDir, id string) [][]string {
	filep := filepath.Join(snapshotRootDir, fmt.Sprintf("%s.txt", id))
	b, err := os.ReadFile(filep)
	if err != nil {
		panic(err)
	}
	groups := [][]string{}
	for _, each := range bytes.Split(b, []byte{'\n'}) {
		line := string(bytes.TrimSpace(each))
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		groups = append(groups, strings.Split(line, ","))
	}
	return groups
}

func createKey(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

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
	bytes, err := v.fs.ReadFile(filep)
	v.req.Nilf(err, "read lines, path: %s, message: %s", filep, message)
	if len(bytes) == 0 {
		return []string{}
	}
	return strings.Split(string(bytes), "\n")
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
func (v *AssertFs) WriteBytes(filep string, bytes []byte) error {
	return v.fs.WriteFile(filep, bytes, 0600)
}
