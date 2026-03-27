package testutil

import "github.com/openelf-labs/apple-tools/core"

// MockRegistry implements core.Registry for testing.
type MockRegistry struct {
	Tools []core.Tool
}

func (m *MockRegistry) Add(t core.Tool) {
	m.Tools = append(m.Tools, t)
}

func (m *MockRegistry) FindTool(name string) *core.Tool {
	for i := range m.Tools {
		if m.Tools[i].Name == name {
			return &m.Tools[i]
		}
	}
	return nil
}

func (m *MockRegistry) ToolNames() []string {
	names := make([]string, len(m.Tools))
	for i, t := range m.Tools {
		names[i] = t.Name
	}
	return names
}
