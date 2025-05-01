// Package snap contains snapshot test functionality.
package snap

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	nonSafeFilenamePattern = regexp.MustCompile(`[^0-9a-zA-Z-._]`)
)

// A SnapshotSuite is a suite of snapshot tests with a shared directory for the snapshot files.
// It is made of [snap.Snapshot]s.
type SnapshotSuite struct {
	rootDir string
}

// NewSnapshotSuite creates a [snap.SnapshotSuite] with a root directory.
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
// Name is [snap.Snapshot.Name] and with [snap.SnapshotSuite.rootDir],
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

// Run the snapshot process according to parameters set in [snap.SnapshotSuite.NewSnapshot].
// Error is returned when something unexpected fails, not when the test itself fails.
// Determining whether any given test fails
// is left for "equal" function defined in [snap.SnapshotSuite.NewSnapshot].
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
