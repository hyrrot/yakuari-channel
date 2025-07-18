package plugins

import (
	"fmt"
	
	"github.com/user/ymmp-compiler/internal/models"
)

// ItemPlugin defines the interface for item type plugins
type ItemPlugin interface {
	GetType() string
	ParseYAML(data interface{}) (models.Item, error)
	ConvertToYMMP(item models.Item, basePath string) (*models.YMMPItem, error)
}

// VoiceService defines the interface for voice synthesis services
type VoiceService interface {
	GetName() string
	CalculateLength(text string, params map[string]interface{}) (float64, error)
	Configure(config map[string]interface{}) error
}

// LengthProvider defines the interface for length calculation
type LengthProvider interface {
	CanCalculate(item models.Item) bool
	CalculateLength(item models.Item, basePath string) (float64, error)
}

// PluginRegistry manages all plugins
type PluginRegistry struct {
	itemPlugins     map[string]ItemPlugin
	voiceServices   map[string]VoiceService
	lengthProviders []LengthProvider
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		itemPlugins:     make(map[string]ItemPlugin),
		voiceServices:   make(map[string]VoiceService),
		lengthProviders: make([]LengthProvider, 0),
	}
}

// RegisterItemPlugin registers a new item plugin
func (r *PluginRegistry) RegisterItemPlugin(plugin ItemPlugin) error {
	itemType := plugin.GetType()
	
	if _, exists := r.itemPlugins[itemType]; exists {
		return fmt.Errorf("plugin for type '%s' already registered", itemType)
	}
	
	r.itemPlugins[itemType] = plugin
	return nil
}

// GetItemPlugin retrieves an item plugin by type
func (r *PluginRegistry) GetItemPlugin(itemType string) (ItemPlugin, bool) {
	plugin, exists := r.itemPlugins[itemType]
	return plugin, exists
}

// RegisterVoiceService registers a new voice service
func (r *PluginRegistry) RegisterVoiceService(service VoiceService) error {
	serviceName := service.GetName()
	
	if _, exists := r.voiceServices[serviceName]; exists {
		return fmt.Errorf("voice service '%s' already registered", serviceName)
	}
	
	r.voiceServices[serviceName] = service
	return nil
}

// GetVoiceService retrieves a voice service by name
func (r *PluginRegistry) GetVoiceService(serviceName string) (VoiceService, bool) {
	service, exists := r.voiceServices[serviceName]
	return service, exists
}

// RegisterLengthProvider registers a new length provider
func (r *PluginRegistry) RegisterLengthProvider(provider LengthProvider) {
	r.lengthProviders = append(r.lengthProviders, provider)
}

// GetLengthProviders returns all registered length providers
func (r *PluginRegistry) GetLengthProviders() []LengthProvider {
	return r.lengthProviders
}