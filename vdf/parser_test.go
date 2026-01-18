package vdf

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "simple key-value",
			input: `"root"
{
	"key"		"value"
}`,
			wantErr: false,
		},
		{
			name: "nested structure",
			input: `"root"
{
	"parent"
	{
		"child"		"value"
	}
}`,
			wantErr: false,
		},
		{
			name: "multiple values",
			input: `"root"
{
	"key1"		"value1"
	"key2"		"value2"
	"key3"		"value3"
}`,
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: false, // Parser handles empty input gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(strings.NewReader(tt.input))
			root, err := parser.Parse()

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && root == nil {
				t.Error("Parse() returned nil root node")
			}
		})
	}
}

func TestFindNode(t *testing.T) {
	input := `"root"
{
	"level1"
	{
		"level2"
		{
			"target"		"found"
		}
		"other"		"value"
	}
}`

	parser := NewParser(strings.NewReader(input))
	root, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantKey string
		wantVal string
		wantNil bool
	}{
		{
			name:    "find nested node",
			path:    "root/level1/level2/target",
			wantKey: "target",
			wantVal: "found",
			wantNil: false,
		},
		{
			name:    "find intermediate node",
			path:    "root/level1/other",
			wantKey: "other",
			wantVal: "value",
			wantNil: false,
		},
		{
			name:    "non-existent path",
			path:    "root/level1/nonexistent",
			wantNil: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := FindNode(root, tt.path)

			if tt.wantNil {
				if node != nil {
					t.Errorf("FindNode() = %v, want nil", node)
				}
				return
			}

			if node == nil {
				t.Error("FindNode() returned nil, want node")
				return
			}

			if node.Key != tt.wantKey {
				t.Errorf("FindNode() key = %v, want %v", node.Key, tt.wantKey)
			}

			if node.Value != tt.wantVal {
				t.Errorf("FindNode() value = %v, want %v", node.Value, tt.wantVal)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
	input := `"root"
{
	"apps"
	{
		"123"
		{
			"LaunchOptions"		"old value"
		}
	}
}`

	parser := NewParser(strings.NewReader(input))
	root, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Test setting an existing value
	err = SetValue(root, "root/apps/123/LaunchOptions", "new value")
	if err != nil {
		t.Errorf("SetValue() error = %v", err)
	}

	node := FindNode(root, "root/apps/123/LaunchOptions")
	if node == nil {
		t.Fatal("FindNode() returned nil after SetValue()")
	}

	if node.Value != "new value" {
		t.Errorf("SetValue() value = %v, want %v", node.Value, "new value")
	}

	// Test setting a non-existent path (should create the path)
	err = SetValue(root, "root/apps/999/NewOption", "created value")
	if err != nil {
		t.Errorf("SetValue() error = %v, want nil", err)
	}

	newNode := FindNode(root, "root/apps/999/NewOption")
	if newNode == nil {
		t.Error("SetValue() should create non-existent path")
	} else if newNode.Value != "created value" {
		t.Errorf("SetValue() created node value = %v, want %v", newNode.Value, "created value")
	}
}

func TestWrite(t *testing.T) {
	input := `"root"
{
	"key1"		"value1"
	"nested"
	{
		"key2"		"value2"
	}
}`

	parser := NewParser(strings.NewReader(input))
	root, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	var output strings.Builder
	err = Write(&output, root, 0)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	result := output.String()
	if result == "" {
		t.Error("Write() produced empty output")
	}

	// Check that it contains the expected keys
	if !strings.Contains(result, "root") {
		t.Error("Write() output missing 'root'")
	}
	if !strings.Contains(result, "key1") {
		t.Error("Write() output missing 'key1'")
	}
	if !strings.Contains(result, "value1") {
		t.Error("Write() output missing 'value1'")
	}
}

func TestRoundTrip(t *testing.T) {
	input := `"UserLocalConfigStore"
{
	"Software"
	{
		"Valve"
		{
			"Steam"
			{
				"apps"
				{
					"570"
					{
						"LaunchOptions"		"gamemoderun %command%"
					}
				}
			}
		}
	}
}`

	// Parse
	parser := NewParser(strings.NewReader(input))
	root, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Modify
	err = SetValue(root, "UserLocalConfigStore/Software/Valve/Steam/apps/570/LaunchOptions", "modified value")
	if err != nil {
		t.Fatalf("SetValue() failed: %v", err)
	}

	// Write back
	var output strings.Builder
	err = Write(&output, root, 0)
	if err != nil {
		t.Fatalf("Write() failed: %v", err)
	}

	// Parse again
	parser2 := NewParser(strings.NewReader(output.String()))
	root2, err := parser2.Parse()
	if err != nil {
		t.Fatalf("Second Parse() failed: %v", err)
	}

	// Verify the modification persisted
	node := FindNode(root2, "UserLocalConfigStore/Software/Valve/Steam/apps/570/LaunchOptions")
	if node == nil {
		t.Fatal("FindNode() returned nil after round-trip")
	}

	if node.Value != "modified value" {
		t.Errorf("Round-trip value = %v, want %v", node.Value, "modified value")
	}
}
