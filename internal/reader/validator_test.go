package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() string
		cleanup func(string)
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid file exists",
			setup: func() string {
				tmpFile, err := os.CreateTemp("", "test-*.log")
				if err != nil {
					t.Fatalf("failed to create temp file: %v", err)
				}
				tmpFile.Close()
				return tmpFile.Name()
			},
			cleanup: func(path string) {
				os.Remove(path)
			},
			wantErr: false,
		},
		{
			name: "file does not exist",
			setup: func() string {
				return "/tmp/nonexistent-file-12345.log"
			},
			cleanup: func(path string) {},
			wantErr: true,
			errMsg:  "arquivo não encontrado",
		},
		{
			name: "path is a directory",
			setup: func() string {
				tmpDir, err := os.MkdirTemp("", "test-dir-*")
				if err != nil {
					t.Fatalf("failed to create temp dir: %v", err)
				}
				return tmpDir
			},
			cleanup: func(path string) {
				os.RemoveAll(path)
			},
			wantErr: true,
			errMsg:  "caminho é um diretório, não um arquivo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			defer tt.cleanup(path)

			err := ValidateFile(path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFile() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateFile() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateFile() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestValidateFile_SymbolicLink(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create a symbolic link to the file
	linkPath := filepath.Join(os.TempDir(), "test-symlink.log")
	err = os.Symlink(tmpFile.Name(), linkPath)
	if err != nil {
		t.Skipf("skipping symlink test: %v", err)
	}
	defer os.Remove(linkPath)

	// Should validate successfully as it points to a valid file
	err = ValidateFile(linkPath)
	if err != nil {
		t.Errorf("ValidateFile() with symlink error = %v, want nil", err)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
