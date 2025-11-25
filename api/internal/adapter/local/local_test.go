package local

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"timeship/internal/adapter"
)

func TestNew(t *testing.T) {
	t.Run("valid directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		a, err := New(tmpDir)
		if err != nil {
			t.Fatalf("New() failed: %v", err)
		}
		defer a.Close()

		if a.root == nil {
			t.Error("expected root to be set")
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := New("/nonexistent/path")
		if err == nil {
			t.Error("expected error for non-existent directory")
		}
	})

	t.Run("file instead of directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "file.txt")
		if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := New(tmpFile)
		if err == nil {
			t.Error("expected error when opening file as root")
		}
	})
}

func TestListContents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.md"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "nested.txt"), []byte("nested"), 0644)

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	t.Run("list root", func(t *testing.T) {
		nodes, err := a.ListContents(url.URL{Scheme: "local", Path: "/"})
		if err != nil {
			t.Fatalf("ListContents failed: %v", err)
		}

		if len(nodes) != 3 {
			t.Errorf("expected 3 nodes, got %d", len(nodes))
		}

		// Check that we have the expected files
		foundDir := false
		foundFile1 := false
		foundFile2 := false

		for _, node := range nodes {
			switch node.Basename {
			case "subdir":
				foundDir = true
				if node.Type != "dir" {
					t.Error("subdir should be type 'dir'")
				}
			case "file1.txt":
				foundFile1 = true
				if node.Type != "file" {
					t.Error("file1.txt should be type 'file'")
				}
				if node.Extension != "txt" {
					t.Errorf("file1.txt extension = %q, want 'txt'", node.Extension)
				}
				if node.Size != 8 {
					t.Errorf("file1.txt size = %d, want 8", node.Size)
				}
			case "file2.md":
				foundFile2 = true
				if node.Extension != "md" {
					t.Errorf("file2.md extension = %q, want 'md'", node.Extension)
				}
			}
		}

		if !foundDir || !foundFile1 || !foundFile2 {
			t.Error("missing expected files/directories")
		}
	})

	t.Run("list with local:// prefix", func(t *testing.T) {
		nodes, err := a.ListContents(url.URL{Scheme: "local", Path: "/"})
		if err != nil {
			t.Fatalf("ListContents failed: %v", err)
		}

		if len(nodes) != 3 {
			t.Errorf("expected 3 nodes, got %d", len(nodes))
		}
	})

	t.Run("list subdirectory", func(t *testing.T) {
		nodes, err := a.ListContents(url.URL{Scheme: "local", Path: "/subdir"})
		if err != nil {
			t.Fatalf("ListContents failed: %v", err)
		}

		if len(nodes) != 1 {
			t.Errorf("expected 1 node, got %d", len(nodes))
		}

		if len(nodes) > 0 {
			if nodes[0].Basename != "nested.txt" {
				t.Errorf("expected nested.txt, got %s", nodes[0].Basename)
			}
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		_, err := a.ListContents(url.URL{Scheme: "local", Path: "/nonexistent"})
		if err == nil {
			t.Error("expected error for non-existent directory")
		}
	})

	t.Run("list file instead of directory", func(t *testing.T) {
		_, err := a.ListContents(url.URL{Scheme: "local", Path: "/file1.txt"})
		if err == nil {
			t.Error("expected error when trying to list a file")
		}
	})

	t.Run("verify paths have local:// prefix when listing root", func(t *testing.T) {
		nodes, err := a.ListContents(url.URL{Scheme: "local", Path: "/"})
		if err != nil {
			t.Fatalf("ListContents failed: %v", err)
		}

		for _, node := range nodes {
			if !strings.HasPrefix(node.Path.String(), "local://") {
				t.Errorf("path %q should have 'local://' prefix", node.Path.String())
			}

			// Verify the path format matches expected: local://basename
			expectedPath := "local://" + node.Basename
			if node.Path.String() != expectedPath {
				t.Errorf("path = %q, want %q", node.Path.String(), expectedPath)
			}
		}
	})

	t.Run("verify paths have local:// prefix when listing subdirectory", func(t *testing.T) {
		nodes, err := a.ListContents(url.URL{Scheme: "local", Path: "/subdir"})
		if err != nil {
			t.Fatalf("ListContents failed: %v", err)
		}

		for _, node := range nodes {
			if !strings.HasPrefix(node.Path.String(), "local://") {
				t.Errorf("path %q should have 'local://' prefix", node.Path.String())
			}

			// Verify the path format matches expected: local://subdir/basename
			expectedPath := "local://subdir/" + node.Basename
			if node.Path.String() != expectedPath {
				t.Errorf("path = %q, want %q", node.Path.String(), expectedPath)
			}
		}
	})

	t.Run("verify paths have local:// prefix when input already has prefix", func(t *testing.T) {
		nodes, err := a.ListContents(url.URL{Scheme: "local", Path: "/subdir"})
		if err != nil {
			t.Fatalf("ListContents failed: %v", err)
		}

		for _, node := range nodes {
			if !strings.HasPrefix(node.Path.String(), "local://") {
				t.Errorf("path %q should have 'local://' prefix", node.Path.String())
			}

			// Verify the path format matches expected: local://subdir/basename
			expectedPath := "local://subdir/" + node.Basename
			if node.Path.String() != expectedPath {
				t.Errorf("path = %q, want %q", node.Path.String(), expectedPath)
			}
		}
	})

	t.Run("verify paths with nested directories", func(t *testing.T) {
		// Create a deeper structure
		nestedDir := filepath.Join(tmpDir, "public", "media")
		os.MkdirAll(nestedDir, 0755)
		os.WriteFile(filepath.Join(nestedDir, "image.jpg"), []byte("fake image"), 0644)

		nodes, err := a.ListContents(url.URL{Scheme: "local", Path: "/public/media"})
		if err != nil {
			t.Fatalf("ListContents failed: %v", err)
		}

		if len(nodes) != 1 {
			t.Errorf("expected 1 node, got %d", len(nodes))
		}

		if len(nodes) > 0 {
			expectedPath := "local://public/media/image.jpg"
			if nodes[0].Path.String() != expectedPath {
				t.Errorf("path = %q, want %q", nodes[0].Path.String(), expectedPath)
			}

			if nodes[0].Basename != "image.jpg" {
				t.Errorf("basename = %q, want 'image.jpg'", nodes[0].Basename)
			}
		}
	})
}

func TestPathTraversalPrevention(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file outside the root
	outsideFile := filepath.Join(filepath.Dir(tmpDir), "outside.txt")
	os.WriteFile(outsideFile, []byte("should not be accessible"), 0644)
	defer os.Remove(outsideFile)

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	t.Run("prevent .. traversal", func(t *testing.T) {
		_, err := a.ListContents(url.URL{Scheme: "local", Path: "/../"})
		if err == nil {
			t.Error("expected error when trying to traverse outside root")
		}
	})

	t.Run("prevent ../../ traversal", func(t *testing.T) {
		_, err := a.FileExists(url.URL{Scheme: "local", Path: "/../../outside.txt"})
		if err == nil {
			t.Error("expected error when trying to access file outside root")
		}
	})

	t.Run("absolute-looking paths are safely relative to root", func(t *testing.T) {
		// Paths like "/etc/passwd" are interpreted as "etc/passwd" relative to the storage root
		// This is safe because os.OpenRoot prevents access outside the root directory
		exists, err := a.FileExists(url.URL{Scheme: "local", Path: "/etc/passwd"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Should return false (doesn't exist) but no error, because it's safely looking
		// for "etc/passwd" relative to tmpDir, not the system /etc/passwd
		if exists {
			t.Error("etc/passwd should not exist in temp directory")
		}
	})
}

func TestMimeType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("plain text"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test.html"), []byte("<html><body>test</body></html>"), 0644)
	// Note: http.DetectContentType doesn't reliably detect JSON from content alone
	// It needs special markers or will default to text/plain

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	tests := []struct {
		file     string
		expected string
	}{
		{"test.txt", "text/plain"},
		{"test.html", "text/html"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			mimeType, err := a.MimeType(url.URL{Scheme: "local", Path: "/" + tt.file})
			if err != nil {
				t.Fatalf("MimeType failed: %v", err)
			}

			// http.DetectContentType returns charset info, so we check prefix
			if mimeType[:len(tt.expected)] != tt.expected {
				t.Errorf("MimeType(%q) = %q, want prefix %q", tt.file, mimeType, tt.expected)
			}
		})
	}

	t.Run("non-existent file", func(t *testing.T) {
		_, err := a.MimeType(url.URL{Scheme: "local", Path: "/nonexistent.txt"})
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestFileSize(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("test content with known length")
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), content, 0644)

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	size, err := a.FileSize(url.URL{Scheme: "local", Path: "/test.txt"})
	if err != nil {
		t.Fatalf("FileSize failed: %v", err)
	}

	if size != int64(len(content)) {
		t.Errorf("FileSize = %d, want %d", size, len(content))
	}

	t.Run("non-existent file", func(t *testing.T) {
		_, err := a.FileSize(url.URL{Scheme: "local", Path: "/nonexistent.txt"})
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestReadStream(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("test file content")
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), content, 0644)

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	stream, err := a.ReadStream(url.URL{Scheme: "local", Path: "/test.txt"})
	if err != nil {
		t.Fatalf("ReadStream failed: %v", err)
	}
	defer stream.Close()

	readContent, err := io.ReadAll(stream)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("ReadStream content = %q, want %q", string(readContent), string(content))
	}

	t.Run("non-existent file", func(t *testing.T) {
		_, err := a.ReadStream(url.URL{Scheme: "local", Path: "/nonexistent.txt"})
		if err == nil {
			t.Error("expected error for non-existent file")
		}
	})
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("content"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "dir"), 0755)

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	t.Run("existing file", func(t *testing.T) {
		exists, err := a.FileExists(url.URL{Scheme: "local", Path: "/exists.txt"})
		if err != nil {
			t.Fatalf("FileExists failed: %v", err)
		}
		if !exists {
			t.Error("FileExists should return true for existing file")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		exists, err := a.FileExists(url.URL{Scheme: "local", Path: "/nonexistent.txt"})
		if err != nil {
			t.Fatalf("FileExists failed: %v", err)
		}
		if exists {
			t.Error("FileExists should return false for non-existent file")
		}
	})

	t.Run("directory", func(t *testing.T) {
		exists, err := a.FileExists(url.URL{Scheme: "local", Path: "/dir"})
		if err != nil {
			t.Fatalf("FileExists failed: %v", err)
		}
		if exists {
			t.Error("FileExists should return false for directory")
		}
	})
}

func TestDirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()

	os.Mkdir(filepath.Join(tmpDir, "dir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	t.Run("existing directory", func(t *testing.T) {
		exists, err := a.DirectoryExists(url.URL{Scheme: "local", Path: "/dir"})
		if err != nil {
			t.Fatalf("DirectoryExists failed: %v", err)
		}
		if !exists {
			t.Error("DirectoryExists should return true for existing directory")
		}
	})

	t.Run("non-existent directory", func(t *testing.T) {
		exists, err := a.DirectoryExists(url.URL{Scheme: "local", Path: "/nonexistent"})
		if err != nil {
			t.Fatalf("DirectoryExists failed: %v", err)
		}
		if exists {
			t.Error("DirectoryExists should return false for non-existent directory")
		}
	})

	t.Run("file", func(t *testing.T) {
		exists, err := a.DirectoryExists(url.URL{Scheme: "local", Path: "/file.txt"})
		if err != nil {
			t.Fatalf("DirectoryExists failed: %v", err)
		}
		if exists {
			t.Error("DirectoryExists should return false for file")
		}
	})
}

func TestEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a restricted directory to test permission errors
	restrictedDir := filepath.Join(tmpDir, "restricted")
	os.Mkdir(restrictedDir, 0755)

	// Create a subdirectory inside it
	os.Mkdir(filepath.Join(restrictedDir, "subdir"), 0755)

	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	// On Unix systems, we can create a permission denied error by removing permissions
	// Note: This test might not work on all systems (e.g., running as root)
	t.Run("stat error handling", func(t *testing.T) {
		// Change permissions to make the subdirectory inaccessible
		// First, we need to remove execute permission on the parent
		if err := os.Chmod(restrictedDir, 0000); err != nil {
			t.Skip("cannot change directory permissions")
		}
		defer os.Chmod(restrictedDir, 0755) // Restore permissions

		// Try to check if a subdirectory exists - this should give a permission error
		_, err := a.DirectoryExists(url.URL{Scheme: "local", Path: "/restricted/subdir"})
		// We expect an error, but not IsNotExist
		if err == nil {
			t.Skip("expected permission error but got none (might be running as root)")
		}
		if os.IsNotExist(err) {
			t.Error("expected permission error, got IsNotExist")
		}
	})
}

func TestImplementsInterfaces(t *testing.T) {
	tmpDir := t.TempDir()
	a, err := New(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	// Test that adapter implements the expected interfaces
	var _ adapter.Lister = a
	var _ adapter.Reader = a
	var _ adapter.Existence = a
}
