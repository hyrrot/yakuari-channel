package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yakuari-channel/video-authorizer/internal/models"
)

func TestApplyPropertyOverrides(t *testing.T) {
	c := NewConverter()

	// Test VoiceItem property override
	voiceItem := &models.VoiceItem{
		Type:  "YukkuriMovieMaker.Project.VoiceItem",
		Serif: "Original text",
		BaseItem: models.BaseItem{
			Frame:  0,
			Length: 100,
		},
	}

	properties := map[string]interface{}{
		"Serif": "Overridden text",
	}

	err := c.applyPropertyOverrides(voiceItem, properties)
	require.NoError(t, err)

	assert.Equal(t, "Overridden text", voiceItem.Serif)
	assert.Equal(t, 0, voiceItem.Frame) // Should not change
	assert.Equal(t, 100, voiceItem.Length) // Should not change
}