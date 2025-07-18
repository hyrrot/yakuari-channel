package plugins

import (
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
)

func TestPluginRegistryCreation(t *testing.T) {
	// Red: This test will fail because PluginRegistry doesn't exist yet
	registry := NewPluginRegistry()
	
	if registry == nil {
		t.Error("Expected registry to be non-nil")
	}
}

func TestPluginRegistryRegisterItemPlugin(t *testing.T) {
	// Red: This test will fail because PluginRegistry doesn't exist yet
	registry := NewPluginRegistry()
	
	// Create a mock plugin
	mockPlugin := &MockItemPlugin{
		itemType: "test",
	}
	
	err := registry.RegisterItemPlugin(mockPlugin)
	if err != nil {
		t.Errorf("Failed to register plugin: %v", err)
	}
	
	// Try to get the plugin
	plugin, exists := registry.GetItemPlugin("test")
	if !exists {
		t.Error("Expected plugin to exist")
	}
	
	if plugin.GetType() != "test" {
		t.Errorf("Expected plugin type 'test', got %s", plugin.GetType())
	}
}

func TestPluginRegistryRegisterDuplicatePlugin(t *testing.T) {
	// Red: This test will fail because PluginRegistry doesn't exist yet
	registry := NewPluginRegistry()
	
	mockPlugin1 := &MockItemPlugin{itemType: "test"}
	mockPlugin2 := &MockItemPlugin{itemType: "test"}
	
	err := registry.RegisterItemPlugin(mockPlugin1)
	if err != nil {
		t.Errorf("Failed to register first plugin: %v", err)
	}
	
	err = registry.RegisterItemPlugin(mockPlugin2)
	if err == nil {
		t.Error("Expected error when registering duplicate plugin")
	}
}

// MockItemPlugin is a test implementation of ItemPlugin
type MockItemPlugin struct {
	itemType string
}

func (m *MockItemPlugin) GetType() string {
	return m.itemType
}

func (m *MockItemPlugin) ParseYAML(data interface{}) (models.Item, error) {
	return &MockItem{itemType: m.itemType}, nil
}

func (m *MockItemPlugin) ConvertToYMMP(item models.Item, basePath string) (*models.YMMPItem, error) {
	return &models.YMMPItem{
		Type:   "Mock.Type",
		Layer:  1,
		Frame:  0,
		Length: 60,
	}, nil
}

// MockItem is a test implementation of Item
type MockItem struct {
	itemType        string
	length          models.LengthSpec
	calculatedLength float64
	startTime       float64
}

func (m *MockItem) GetType() string {
	return m.itemType
}

func (m *MockItem) GetLength() models.LengthSpec {
	return m.length
}

func (m *MockItem) SetCalculatedLength(seconds float64) {
	m.calculatedLength = seconds
}

func (m *MockItem) GetStartTime() float64 {
	return m.startTime
}

func (m *MockItem) SetStartTime(seconds float64) {
	m.startTime = seconds
}