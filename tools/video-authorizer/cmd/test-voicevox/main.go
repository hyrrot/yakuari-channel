package main

import (
	"fmt"
	"os"

	"github.com/yakuari-channel/video-authorizer/internal/converter"
)

func main() {
	fmt.Println("VOICEVOX Connection Test")
	fmt.Println("========================")

	// Test default VOICEVOX endpoint
	client := converter.NewVoicevoxClient("")
	
	fmt.Println("Testing connection...")
	if client.IsAvailableWithLogging(true) {
		fmt.Println("✅ VOICEVOX API is available!")
		
		// Test actual API call
		fmt.Println("\nTesting audio duration calculation...")
		duration, err := client.GetAudioDuration("こんにちは", 1)
		if err != nil {
			fmt.Printf("❌ Audio duration test failed: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✅ Audio duration calculated: %.2f seconds\n", duration)
		frames := converter.ConvertDurationToFrames(duration, 30.0)
		fmt.Printf("✅ Converted to frames: %d frames (30 FPS)\n", frames)
		
	} else {
		fmt.Println("❌ VOICEVOX API is not available")
		fmt.Println("\nTroubleshooting:")
		fmt.Println("1. Is VOICEVOX running?")
		fmt.Println("2. Is it listening on http://localhost:50021?")
		fmt.Println("3. Are there any firewall restrictions?")
		os.Exit(1)
	}
	
	fmt.Println("\n🎉 All tests passed!")
}