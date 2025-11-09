package local

import (
	"testing"
	"time"
)

func TestParseTimestampFromName(t *testing.T) {
	tests := []struct {
		name         string
		snapshotName string
		patterns     []DateTimePattern
		wantUnix     int64
		wantParsed   bool
	}{
		{
			name:         "auto-weekly with default patterns",
			snapshotName: "auto-weekly-2025-11-09_00-00",
			patterns:     DefaultDateTimePatterns(),
			wantUnix:     time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC).Unix(),
			wantParsed:   true,
		},
		{
			name:         "auto-hourly with default patterns",
			snapshotName: "auto-hourly-2025-11-09_13-30",
			patterns:     DefaultDateTimePatterns(),
			wantUnix:     time.Date(2025, 11, 9, 13, 30, 0, 0, time.UTC).Unix(),
			wantParsed:   true,
		},
		{
			name:         "auto-daily with default patterns",
			snapshotName: "auto-daily-2025-11-09_00-00",
			patterns:     DefaultDateTimePatterns(),
			wantUnix:     time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC).Unix(),
			wantParsed:   true,
		},
		{
			name:         "with seconds",
			snapshotName: "backup-2025-11-09_14-30-45",
			patterns:     DefaultDateTimePatterns(),
			wantUnix:     time.Date(2025, 11, 9, 14, 30, 45, 0, time.UTC).Unix(),
			wantParsed:   true,
		},
		{
			name:         "compact format",
			snapshotName: "snapshot_20251109_143045",
			patterns:     DefaultDateTimePatterns(),
			wantUnix:     time.Date(2025, 11, 9, 14, 30, 45, 0, time.UTC).Unix(),
			wantParsed:   true,
		},
		{
			name:         "date only",
			snapshotName: "daily-2025-11-09",
			patterns:     DefaultDateTimePatterns(),
			wantUnix:     time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC).Unix(),
			wantParsed:   true,
		},
		{
			name:         "no matching pattern",
			snapshotName: "random-snapshot-name",
			patterns:     DefaultDateTimePatterns(),
			wantUnix:     0,
			wantParsed:   false,
		},
		{
			name:         "custom pattern",
			snapshotName: "snap_20251109",
			patterns: []DateTimePattern{
				{
					Regex:  `snap_(\d{8})`,
					Layout: "20060102",
				},
			},
			wantUnix:   time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC).Unix(),
			wantParsed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zfs := NewZFSWithConfig("/tmp", ZFSConfig{
				DateTimePatterns: tt.patterns,
			})

			gotUnix, gotParsed := zfs.parseTimestampFromName(tt.snapshotName)

			if gotParsed != tt.wantParsed {
				t.Errorf("parseTimestampFromName() parsed = %v, want %v", gotParsed, tt.wantParsed)
			}

			if gotParsed && gotUnix != tt.wantUnix {
				t.Errorf("parseTimestampFromName() unix = %v, want %v", gotUnix, tt.wantUnix)
				t.Errorf("  got time:  %v", time.Unix(gotUnix, 0).UTC())
				t.Errorf("  want time: %v", time.Unix(tt.wantUnix, 0).UTC())
			}
		})
	}
}

func TestDefaultDateTimePatterns(t *testing.T) {
	patterns := DefaultDateTimePatterns()

	if len(patterns) == 0 {
		t.Error("DefaultDateTimePatterns() returned empty slice")
	}

	for i, pattern := range patterns {
		if pattern.Regex == "" {
			t.Errorf("pattern %d has empty Regex", i)
		}
		if pattern.Layout == "" {
			t.Errorf("pattern %d has empty Layout", i)
		}
	}
}

func TestNewZFS(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		zfs := NewZFS("/tmp")
		if zfs == nil {
			t.Fatal("NewZFS returned nil")
		}
		if zfs.rootDir != "/tmp" {
			t.Errorf("rootDir = %q, want %q", zfs.rootDir, "/tmp")
		}
		if len(zfs.dateTimePatterns) == 0 {
			t.Error("dateTimePatterns is empty, expected default patterns")
		}
	})

	t.Run("custom config", func(t *testing.T) {
		customPatterns := []DateTimePattern{
			{
				Regex:  `custom_(\d{8})`,
				Layout: "20060102",
			},
		}
		zfs := NewZFSWithConfig("/tmp", ZFSConfig{
			DateTimePatterns: customPatterns,
		})
		if zfs == nil {
			t.Fatal("NewZFSWithConfig returned nil")
		}
		if len(zfs.dateTimePatterns) != 1 {
			t.Errorf("dateTimePatterns length = %d, want 1", len(zfs.dateTimePatterns))
		}
		if zfs.dateTimePatterns[0].compiled == nil {
			t.Error("pattern regex was not compiled")
		}
	})

	t.Run("empty config uses defaults", func(t *testing.T) {
		zfs := NewZFSWithConfig("/tmp", ZFSConfig{})
		if zfs == nil {
			t.Fatal("NewZFSWithConfig returned nil")
		}
		if len(zfs.dateTimePatterns) == 0 {
			t.Error("empty config should use default patterns")
		}
	})
}
