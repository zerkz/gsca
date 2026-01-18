package steam

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterGameIDs(t *testing.T) {
	allGameIDs := []string{"100", "200", "300", "400", "500"}
	allowList := []string{"100", "300"}
	denyList := []string{"200", "400"}

	tests := []struct {
		name      string
		allIDs    []string
		allowList []string
		denyList  []string
		want      []string
	}{
		{
			name:      "with allow list",
			allIDs:    allGameIDs,
			allowList: allowList,
			denyList:  nil,
			want:      []string{"100", "300"},
		},
		{
			name:      "with deny list",
			allIDs:    allGameIDs,
			allowList: nil,
			denyList:  denyList,
			want:      []string{"100", "300", "500"},
		},
		{
			name:      "with both lists (allow takes precedence)",
			allIDs:    allGameIDs,
			allowList: allowList,
			denyList:  denyList,
			want:      []string{"100", "300"},
		},
		{
			name:      "no filters",
			allIDs:    allGameIDs,
			allowList: nil,
			denyList:  nil,
			want:      allGameIDs,
		},
		{
			name:      "empty allow list (treated as no filter)",
			allIDs:    allGameIDs,
			allowList: []string{},
			denyList:  nil,
			want:      allGameIDs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterGameIDs(tt.allIDs, tt.allowList, tt.denyList)

			if len(got) != len(tt.want) {
				t.Errorf("FilterGameIDs() length = %v, want %v", len(got), len(tt.want))
				return
			}

			gotMap := make(map[string]bool)
			for _, id := range got {
				gotMap[id] = true
			}

			for _, id := range tt.want {
				if !gotMap[id] {
					t.Errorf("FilterGameIDs() missing ID %v", id)
				}
			}
		})
	}
}

func TestLoadFilterList(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-list.txt")

	content := `# This is a comment
Counter-Strike 2
570

# Another comment
Dota 2
730
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		file    string
		want    []string
		wantErr bool
	}{
		{
			name:    "valid file",
			file:    testFile,
			want:    []string{"Counter-Strike 2", "570", "Dota 2", "730"},
			wantErr: false,
		},
		{
			name:    "non-existent file",
			file:    filepath.Join(tmpDir, "nonexistent.txt"),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFilterList(tt.file)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFilterList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("LoadFilterList() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for i, item := range got {
				if item != tt.want[i] {
					t.Errorf("LoadFilterList()[%d] = %v, want %v", i, item, tt.want[i])
				}
			}
		})
	}
}

func TestResolveGameIDs(t *testing.T) {
	mapping := map[string]string{
		"counter-strike 2": "730",
		"dota 2":           "570",
		"730":              "730", // Direct ID mapping
		"570":              "570",
	}

	tests := []struct {
		name       string
		list       []string
		mapping    map[string]string
		wantIDs    []string
		wantMissed []string
	}{
		{
			name:       "numeric IDs only",
			list:       []string{"730", "570"},
			mapping:    mapping,
			wantIDs:    []string{"730", "570"},
			wantMissed: []string{},
		},
		{
			name:       "game names rejected",
			list:       []string{"Counter-Strike 2", "Dota 2"},
			mapping:    mapping,
			wantIDs:    []string{},
			wantMissed: []string{"Counter-Strike 2", "Dota 2"},
		},
		{
			name:       "mixed IDs and names",
			list:       []string{"730", "Counter-Strike 2"},
			mapping:    mapping,
			wantIDs:    []string{"730"},
			wantMissed: []string{"Counter-Strike 2"},
		},
		{
			name:       "invalid numeric ID",
			list:       []string{"730", "999999"},
			mapping:    mapping,
			wantIDs:    []string{"730", "999999"},
			wantMissed: []string{},
		},
		{
			name:       "non-alphanumeric rejected",
			list:       []string{"730", "test-game"},
			mapping:    mapping,
			wantIDs:    []string{"730"},
			wantMissed: []string{"test-game"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIDs, gotMissed := ResolveGameIDs(tt.list, tt.mapping)

			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("ResolveGameIDs() IDs length = %v, want %v", len(gotIDs), len(tt.wantIDs))
			}

			for i, id := range gotIDs {
				if i < len(tt.wantIDs) && id != tt.wantIDs[i] {
					t.Errorf("ResolveGameIDs() ID[%d] = %v, want %v", i, id, tt.wantIDs[i])
				}
			}

			if len(gotMissed) != len(tt.wantMissed) {
				t.Errorf("ResolveGameIDs() missed length = %v, want %v", len(gotMissed), len(tt.wantMissed))
			}

			for i, missed := range gotMissed {
				if i < len(tt.wantMissed) && missed != tt.wantMissed[i] {
					t.Errorf("ResolveGameIDs() missed[%d] = %v, want %v", i, missed, tt.wantMissed[i])
				}
			}
		})
	}
}

func TestGetLibraryFolders(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	steamappsDir := filepath.Join(tmpDir, "steamapps")
	err := os.MkdirAll(steamappsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create steamapps dir: %v", err)
	}

	// Create a test libraryfolders.vdf
	libraryContent := `"libraryfolders"
{
	"0"
	{
		"path"		"/home/user/.local/share/Steam"
	}
	"1"
	{
		"path"		"/mnt/games"
	}
}`

	libraryFile := filepath.Join(steamappsDir, "libraryfolders.vdf")
	err = os.WriteFile(libraryFile, []byte(libraryContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create libraryfolders.vdf: %v", err)
	}

	tests := []struct {
		name      string
		steamPath string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "valid library folders file",
			steamPath: tmpDir,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "missing library folders file",
			steamPath: t.TempDir(),
			wantCount: 1, // Should return default path
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLibraryFolders(tt.steamPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLibraryFolders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("GetLibraryFolders() count = %v, want %v", len(got), tt.wantCount)
			}
		})
	}
}
