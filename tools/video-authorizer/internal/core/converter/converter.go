package converter

import (
	"fmt"
	"github.com/user/ymmp-compiler/internal/models"
	"github.com/user/ymmp-compiler/internal/plugins"
	"github.com/user/ymmp-compiler/internal/plugins/items"
)

// Convert converts a SYMMP episode to YMMP format
func Convert(episode *models.Episode, basePath string) (*models.YMMPProject, error) {
	project := &models.YMMPProject{
		Timeline: models.Timeline{
			Items: []models.YMMPItem{},
		},
	}
	
	// For now, just return an empty project
	// TODO: Implement actual conversion logic
	
	return project, nil
}

// ConvertToYMMP converts a SYMMP episode to YMMP format
func ConvertToYMMP(episode *models.Episode, basePath string) (*models.YMMPProject, error) {
	project := &models.YMMPProject{
		Timeline: models.Timeline{
			Items: []models.YMMPItem{},
		},
	}
	
	// Create plugin registry
	registry := plugins.NewPluginRegistry()
	registry.RegisterItemPlugin(items.NewImagePlugin())
	registry.RegisterItemPlugin(items.NewVoicePlugin())
	registry.RegisterItemPlugin(items.NewAudioPlugin())
	registry.RegisterItemPlugin(items.NewVideoPlugin())
	registry.RegisterItemPlugin(items.NewTachiePlugin())
	
	// Convert all items to YMMP format
	for _, sequence := range episode.Sequences {
		for _, scene := range sequence.Scenes {
			for _, shot := range scene.Shots {
				for _, item := range shot.Items {
					// Get the appropriate plugin for this item type
					plugin, exists := registry.GetItemPlugin(item.GetType())
					if !exists {
						return nil, fmt.Errorf("no plugin found for item type: %s", item.GetType())
					}
					
					// Convert the item to YMMP format
					ymmItem, err := plugin.ConvertToYMMP(item, basePath)
					if err != nil {
						return nil, fmt.Errorf("failed to convert %s item: %v", item.GetType(), err)
					}
					
					// Add the item to the timeline
					project.Timeline.Items = append(project.Timeline.Items, *ymmItem)
				}
			}
		}
	}
	
	return project, nil
}