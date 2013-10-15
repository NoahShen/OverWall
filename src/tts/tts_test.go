package tts

import (
	"fmt"
	"testing"
)

func TestTTSClearCache(t *testing.T) {
	speechDir := "/home/noah/workspace/OverWall/news_speech_file/"
	var cacheSize int64 = 1024 * 1024 * 1 // 1MB
	ttsManager := NewTTSManager(speechDir, 500, 4, cacheSize)
	deletedFiles := ttsManager.ClearSpeechFiles()
	fmt.Printf("delete files: %v\n", deletedFiles)
	for _, f := range deletedFiles {
		fmt.Printf("files: %v\n", f.Name())
	}
}
