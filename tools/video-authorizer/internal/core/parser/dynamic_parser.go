package parser

import (
	"fmt"
	"io"
	
	"gopkg.in/yaml.v3"
	"github.com/user/ymmp-compiler/internal/models"
	"github.com/user/ymmp-compiler/internal/plugins"
)

// ParseSYMMPWithPlugins parses a SYMMP file with dynamic item parsing using plugins
func ParseSYMMPWithPlugins(reader io.Reader, registry *plugins.PluginRegistry) (*models.Episode, error) {
	// First, parse the YAML into a generic structure
	var rawData map[string]interface{}
	
	decoder := yaml.NewDecoder(reader)
	err := decoder.Decode(&rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %v", err)
	}
	
	// Create episode structure
	episode := &models.Episode{}
	
	// Parse basic fields
	if version, ok := rawData["symmp_format_version"]; ok {
		if versionStr, ok := version.(string); ok {
			episode.SYMMPFormatVersion = versionStr
		} else if versionFloat, ok := version.(float64); ok {
			episode.SYMMPFormatVersion = fmt.Sprintf("%.1f", versionFloat)
		}
	}
	
	if projectData, ok := rawData["project"]; ok {
		if projectMap, ok := projectData.(map[string]interface{}); ok {
			episode.Project = parseProjectConfig(projectMap)
		}
	}
	
	if defaults, ok := rawData["defaults"]; ok {
		if defaultsMap, ok := defaults.(map[string]interface{}); ok {
			episode.Defaults = defaultsMap
		}
	}
	
	// Parse sequences with dynamic items
	if sequences, ok := rawData["sequences"]; ok {
		if sequenceList, ok := sequences.([]interface{}); ok {
			parsedSequences, err := parseSequences(sequenceList, registry)
			if err != nil {
				return nil, fmt.Errorf("failed to parse sequences: %v", err)
			}
			episode.Sequences = parsedSequences
		}
	}
	
	return episode, nil
}

func parseProjectConfig(projectData map[string]interface{}) models.ProjectConfig {
	config := models.ProjectConfig{}
	
	if name, ok := projectData["name"]; ok {
		if nameStr, ok := name.(string); ok {
			config.Name = nameStr
		}
	}
	
	if outputDir, ok := projectData["output_dir"]; ok {
		if outputDirStr, ok := outputDir.(string); ok {
			config.OutputDir = outputDirStr
		}
	}
	
	return config
}

func parseSequences(sequenceList []interface{}, registry *plugins.PluginRegistry) ([]models.Sequence, error) {
	var sequences []models.Sequence
	
	for _, seqData := range sequenceList {
		seqMap, ok := seqData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("sequence must be a map")
		}
		
		sequence := models.Sequence{}
		
		if id, ok := seqMap["id"]; ok {
			if idStr, ok := id.(string); ok {
				sequence.ID = idStr
			}
		}
		
		if scenes, ok := seqMap["scenes"]; ok {
			if sceneList, ok := scenes.([]interface{}); ok {
				parsedScenes, err := parseScenes(sceneList, registry)
				if err != nil {
					return nil, fmt.Errorf("failed to parse scenes in sequence %s: %v", sequence.ID, err)
				}
				sequence.Scenes = parsedScenes
			}
		}
		
		sequences = append(sequences, sequence)
	}
	
	return sequences, nil
}

func parseScenes(sceneList []interface{}, registry *plugins.PluginRegistry) ([]models.Scene, error) {
	var scenes []models.Scene
	
	for _, sceneData := range sceneList {
		sceneMap, ok := sceneData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("scene must be a map")
		}
		
		scene := models.Scene{}
		
		if id, ok := sceneMap["id"]; ok {
			if idStr, ok := id.(string); ok {
				scene.ID = idStr
			}
		}
		
		if shots, ok := sceneMap["shots"]; ok {
			if shotList, ok := shots.([]interface{}); ok {
				parsedShots, err := parseShots(shotList, registry)
				if err != nil {
					return nil, fmt.Errorf("failed to parse shots in scene %s: %v", scene.ID, err)
				}
				scene.Shots = parsedShots
			}
		}
		
		scenes = append(scenes, scene)
	}
	
	return scenes, nil
}

func parseShots(shotList []interface{}, registry *plugins.PluginRegistry) ([]models.Shot, error) {
	var shots []models.Shot
	
	for _, shotData := range shotList {
		shotMap, ok := shotData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("shot must be a map")
		}
		
		shot := models.Shot{}
		
		// Parse items dynamically
		items, err := parseItems(shotMap, registry)
		if err != nil {
			return nil, fmt.Errorf("failed to parse items in shot: %v", err)
		}
		shot.Items = items
		
		shots = append(shots, shot)
	}
	
	return shots, nil
}

func parseItems(shotMap map[string]interface{}, registry *plugins.PluginRegistry) ([]models.Item, error) {
	var items []models.Item
	
	// Iterate through each key-value pair in the shot
	for itemType, itemData := range shotMap {
		// Try to get the plugin for this item type
		plugin, exists := registry.GetItemPlugin(itemType)
		if !exists {
			return nil, fmt.Errorf("unknown item type: %s", itemType)
		}
		
		// Parse the item using the plugin
		item, err := plugin.ParseYAML(itemData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s item: %v", itemType, err)
		}
		
		items = append(items, item)
	}
	
	return items, nil
}