package adapter

import "testing"

func TestStripPrefix(t *testing.T) {
	tests := []struct {
		name        string
		vfPath      string
		adapterName string
		expected    string
	}{
		{
			name:        "strip local prefix from file path",
			vfPath:      "local://documents/file.txt",
			adapterName: "local",
			expected:    "documents/file.txt",
		},
		{
			name:        "strip local prefix from root returns dot",
			vfPath:      "local://",
			adapterName: "local",
			expected:    ".",
		},
		{
			name:        "no prefix to strip",
			vfPath:      "documents/file.txt",
			adapterName: "local",
			expected:    "documents/file.txt",
		},
		{
			name:        "strip s3 prefix",
			vfPath:      "s3://bucket/key.txt",
			adapterName: "s3",
			expected:    "bucket/key.txt",
		},
		{
			name:        "empty path returns dot",
			vfPath:      "",
			adapterName: "local",
			expected:    ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripPrefix(tt.vfPath, tt.adapterName)
			if result != tt.expected {
				t.Errorf("StripPrefix(%q, %q) = %q, want %q", tt.vfPath, tt.adapterName, result, tt.expected)
			}
		})
	}
}

func TestAddPrefix(t *testing.T) {
	tests := []struct {
		name        string
		fsPath      string
		adapterName string
		expected    string
	}{
		{
			name:        "add local prefix to file path",
			fsPath:      "documents/file.txt",
			adapterName: "local",
			expected:    "local://documents/file.txt",
		},
		{
			name:        "add local prefix to empty path",
			fsPath:      "",
			adapterName: "local",
			expected:    "local://",
		},
		{
			name:        "path already has prefix",
			fsPath:      "local://documents/file.txt",
			adapterName: "local",
			expected:    "local://documents/file.txt",
		},
		{
			name:        "add s3 prefix",
			fsPath:      "bucket/key.txt",
			adapterName: "s3",
			expected:    "s3://bucket/key.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddPrefix(tt.fsPath, tt.adapterName)
			if result != tt.expected {
				t.Errorf("AddPrefix(%q, %q) = %q, want %q", tt.fsPath, tt.adapterName, result, tt.expected)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name        string
		basePath    string
		component   string
		adapterName string
		expected    string
	}{
		{
			name:        "join with local prefix at root",
			basePath:    "local://",
			component:   "file.txt",
			adapterName: "local",
			expected:    "local://file.txt",
		},
		{
			name:        "join with local prefix in subdirectory",
			basePath:    "local://documents",
			component:   "file.txt",
			adapterName: "local",
			expected:    "local://documents/file.txt",
		},
		{
			name:        "join without prefix (adds it)",
			basePath:    "documents",
			component:   "file.txt",
			adapterName: "local",
			expected:    "local://documents/file.txt",
		},
		{
			name:        "join empty base path",
			basePath:    "",
			component:   "file.txt",
			adapterName: "local",
			expected:    "local://file.txt",
		},
		{
			name:        "join nested path",
			basePath:    "local://public/media",
			component:   "image.jpg",
			adapterName: "local",
			expected:    "local://public/media/image.jpg",
		},
		{
			name:        "join with s3 adapter",
			basePath:    "s3://bucket",
			component:   "key.txt",
			adapterName: "s3",
			expected:    "s3://bucket/key.txt",
		},
		{
			name:        "join with trailing slash",
			basePath:    "local://documents/",
			component:   "file.txt",
			adapterName: "local",
			expected:    "local://documents/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinPath(tt.basePath, tt.component, tt.adapterName)
			if result != tt.expected {
				t.Errorf("JoinPath(%q, %q, %q) = %q, want %q", tt.basePath, tt.component, tt.adapterName, result, tt.expected)
			}
		})
	}
}
