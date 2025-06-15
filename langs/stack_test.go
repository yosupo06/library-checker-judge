package langs

import (
	"testing"
)

func TestStackLimitOptions(t *testing.T) {
	tests := []struct {
		name     string
		option   TaskInfoOption
		expected int
	}{
		{
			name:     "WithUnlimitedStackLimit",
			option:   WithUnlimitedStackLimit(),
			expected: -1,
		},
		{
			name:     "WithStackLimitBytes",
			option:   WithStackLimitBytes(8192000),
			expected: 8192000,
		},
		{
			name:     "WithStackLimitMB",
			option:   WithStackLimitMB(8),
			expected: 8 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &TaskInfo{}
			if err := tt.option(ti); err != nil {
				t.Fatalf("Option failed: %v", err)
			}
			if ti.StackLimitBytes != tt.expected {
				t.Errorf("Expected StackLimitBytes=%d, got %d", tt.expected, ti.StackLimitBytes)
			}
		})
	}
}

func TestStackLimitUnitsConsistency(t *testing.T) {
	// Test that different ways to specify the same stack size result in the same value
	testCases := []struct {
		name    string
		options []TaskInfoOption
	}{
		{
			name: "8MB specified in different units",
			options: []TaskInfoOption{
				WithStackLimitBytes(8 * 1024 * 1024),
				WithStackLimitMB(8),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var results []int
			for _, option := range tc.options {
				ti := &TaskInfo{}
				if err := option(ti); err != nil {
					t.Fatalf("Option failed: %v", err)
				}
				results = append(results, ti.StackLimitBytes)
			}

			// All should be equal
			for i := 1; i < len(results); i++ {
				if results[0] != results[i] {
					t.Errorf("Inconsistent stack limits: %v", results)
					break
				}
			}
		})
	}
}