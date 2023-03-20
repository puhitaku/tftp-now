package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrustedRoot(t *testing.T) {
	actualTemp, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		t.Fatalf("failed to get the actual path of the temp dir: %s", err)
	}

	root := strings.TrimRight(actualTemp, "/")
	rootName := filepath.Base(root)
	defer os.RemoveAll(root)

	err = touch(filepath.Join(root, "foo"))
	if err != nil {
		t.Fatalf("failed to create foo: %s", err)
	}

	err = os.Symlink("/etc/passwd", filepath.Join(root, "symlink"))
	if err != nil {
		t.Fatalf("failed to create a symlink: %s", err)
	}

	var testCases = []struct {
		name      string
		dir       string
		trusted   bool
		validLink bool
	}{
		{
			name:    "trusted root itself",
			dir:     ".",
			trusted: true,
		},
		{
			name:    "parent of trusted root",
			dir:     "..",
			trusted: false,
		},
		{
			name:    "trusted root itself, but redundant",
			dir:     "../" + rootName,
			trusted: true,
		},
		{
			name:    "single slash should point the trusted root",
			dir:     "/",
			trusted: true,
		},
		{
			name:    "link to /etc/passwd",
			dir:     "symlink",
			trusted: false,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			_, err := evaluatePath(c.dir, root, true)
			if actual := err == nil; actual != c.trusted {
				t.Errorf("unexpected trust, expect: %t, actual: %t", c.trusted, err == nil)
			}
		})
	}
}

func touch(path string) error {
	f, err := os.Create(path)
	defer f.Close()
	return err
}
