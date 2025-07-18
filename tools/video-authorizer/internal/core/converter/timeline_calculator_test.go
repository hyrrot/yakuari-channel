package converter

import (
	"testing"
	
	"github.com/user/ymmp-compiler/internal/models"
	"github.com/user/ymmp-compiler/internal/plugins/items"
)

func TestCalculateTimelineSimpleNumeric(t *testing.T) {
	// Red: This test will fail because CalculateTimeline doesn't exist yet
	episode := &models.Episode{
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID: "scene1",
						Shots: []models.Shot{
							{
								Items: []models.Item{
									&items.ImageItem{
										FilePath: "./test.png",
										X:        0,
										Y:        0,
										Z:        0,
										Zoom:     1.0,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Simulate parsed length
	imageItem := episode.Sequences[0].Scenes[0].Shots[0].Items[0].(*items.ImageItem)
	imageItem.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 5.0})
	
	err := CalculateTimeline(episode)
	if err != nil {
		t.Errorf("CalculateTimeline failed: %v", err)
	}
	
	// Check calculated start time
	if imageItem.GetStartTime() != 0.0 {
		t.Errorf("Expected start time 0.0, got %f", imageItem.GetStartTime())
	}
}

func TestCalculateTimelineSequentialItems(t *testing.T) {
	// Red: This test will fail because CalculateTimeline doesn't exist yet
	episode := &models.Episode{
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID: "scene1",
						Shots: []models.Shot{
							{
								Items: []models.Item{
									&items.ImageItem{
										FilePath: "./image1.png",
									},
								},
							},
						},
					},
					{
						ID: "scene2",
						Shots: []models.Shot{
							{
								Items: []models.Item{
									&items.ImageItem{
										FilePath: "./image2.png",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Set lengths
	image1 := episode.Sequences[0].Scenes[0].Shots[0].Items[0].(*items.ImageItem)
	image1.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 3.0})
	
	image2 := episode.Sequences[0].Scenes[1].Shots[0].Items[0].(*items.ImageItem)
	image2.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 2.0})
	
	err := CalculateTimeline(episode)
	if err != nil {
		t.Errorf("CalculateTimeline failed: %v", err)
	}
	
	// Check calculated start times
	if image1.GetStartTime() != 0.0 {
		t.Errorf("Expected first image start time 0.0, got %f", image1.GetStartTime())
	}
	
	if image2.GetStartTime() != 3.0 {
		t.Errorf("Expected second image start time 3.0, got %f", image2.GetStartTime())
	}
}

func TestCalculateTimelineSimultaneousItems(t *testing.T) {
	// Red: This test will fail because CalculateTimeline doesn't exist yet
	episode := &models.Episode{
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID: "scene1",
						Shots: []models.Shot{
							{
								Items: []models.Item{
									&items.ImageItem{
										FilePath: "./image.png",
									},
									&items.VoiceItem{
										Line: "テスト音声",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Set lengths
	imageItem := episode.Sequences[0].Scenes[0].Shots[0].Items[0].(*items.ImageItem)
	imageItem.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 5.0})
	
	voiceItem := episode.Sequences[0].Scenes[0].Shots[0].Items[1].(*items.VoiceItem)
	voiceItem.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 3.0})
	
	err := CalculateTimeline(episode)
	if err != nil {
		t.Errorf("CalculateTimeline failed: %v", err)
	}
	
	// Both items should start at the same time (they're in the same shot)
	if imageItem.GetStartTime() != 0.0 {
		t.Errorf("Expected image start time 0.0, got %f", imageItem.GetStartTime())
	}
	
	if voiceItem.GetStartTime() != 0.0 {
		t.Errorf("Expected voice start time 0.0, got %f", voiceItem.GetStartTime())
	}
}

func TestCalculateTimelineUntilShotEnd(t *testing.T) {
	// Red: This test will fail because CalculateTimeline doesn't exist yet
	episode := &models.Episode{
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID: "scene1",
						Shots: []models.Shot{
							{
								Items: []models.Item{
									&items.ImageItem{
										FilePath: "./image.png",
									},
									&items.VoiceItem{
										Line: "テスト音声",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Set lengths - image should extend until shot end
	imageItem := episode.Sequences[0].Scenes[0].Shots[0].Items[0].(*items.ImageItem)
	imageItem.SetLength(models.LengthSpec{Type: models.LengthTypeUntilShotEnd})
	
	voiceItem := episode.Sequences[0].Scenes[0].Shots[0].Items[1].(*items.VoiceItem)
	voiceItem.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 4.0})
	
	err := CalculateTimeline(episode)
	if err != nil {
		t.Errorf("CalculateTimeline failed: %v", err)
	}
	
	// Image should be calculated to match the longest item (voice = 4.0 seconds)
	if imageItem.GetLength().Value != 4.0 {
		t.Errorf("Expected image length 4.0, got %f", imageItem.GetLength().Value)
	}
}

func TestCalculateTimelineUntilSceneEnd(t *testing.T) {
	// Red: This test will fail because CalculateTimeline doesn't exist yet
	episode := &models.Episode{
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID: "scene1",
						Shots: []models.Shot{
							{
								Items: []models.Item{
									&items.ImageItem{
										FilePath: "./background.png",
									},
								},
							},
							{
								Items: []models.Item{
									&items.VoiceItem{
										Line: "最初のセリフ",
									},
								},
							},
							{
								Items: []models.Item{
									&items.VoiceItem{
										Line: "2番目のセリフ",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Set lengths
	imageItem := episode.Sequences[0].Scenes[0].Shots[0].Items[0].(*items.ImageItem)
	imageItem.SetLength(models.LengthSpec{Type: models.LengthTypeUntilSceneEnd})
	
	voice1 := episode.Sequences[0].Scenes[0].Shots[1].Items[0].(*items.VoiceItem)
	voice1.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 2.0})
	
	voice2 := episode.Sequences[0].Scenes[0].Shots[2].Items[0].(*items.VoiceItem)
	voice2.SetLength(models.LengthSpec{Type: models.LengthTypeNumeric, Value: 3.0})
	
	err := CalculateTimeline(episode)
	if err != nil {
		t.Errorf("CalculateTimeline failed: %v", err)
	}
	
	// Image should extend until scene end (2.0 + 3.0 = 5.0 seconds)
	if imageItem.GetLength().Value != 5.0 {
		t.Errorf("Expected image length 5.0, got %f", imageItem.GetLength().Value)
	}
}

func TestCalculateTimelineAutoLength(t *testing.T) {
	// Red: This test will fail because CalculateTimeline doesn't exist yet
	episode := &models.Episode{
		Sequences: []models.Sequence{
			{
				ID: "seq1",
				Scenes: []models.Scene{
					{
						ID: "scene1",
						Shots: []models.Shot{
							{
								Items: []models.Item{
									&items.VoiceItem{
										Line: "自動計算テスト",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	
	// Set auto length
	voiceItem := episode.Sequences[0].Scenes[0].Shots[0].Items[0].(*items.VoiceItem)
	voiceItem.SetLength(models.LengthSpec{Type: models.LengthTypeAuto})
	
	// Mock calculated length (this would normally come from VOICEVOX)
	voiceItem.SetCalculatedLength(2.5)
	
	err := CalculateTimeline(episode)
	if err != nil {
		t.Errorf("CalculateTimeline failed: %v", err)
	}
	
	// Voice item should use calculated length
	if voiceItem.GetLength().Value != 2.5 {
		t.Errorf("Expected voice length 2.5, got %f", voiceItem.GetLength().Value)
	}
}