package tts

import (
	"os"
	"path/filepath"
	"sort"
)

type TTSManager struct {
	speechFileDir string
	speechCache   int64
	sentenceLen   int
	limitCh       chan int
}

func NewTTSManager(speechFileDir string, sentenceLen, taskLimit int, speechCache int64) *TTSManager {
	ttsManager := &TTSManager{
		speechFileDir,
		speechCache,
		sentenceLen,
		make(chan int, taskLimit),
	}
	return ttsManager
}

func (self *TTSManager) GenerateSpeechFiles(text, outputFileName string, speaker int) (string, error) {
	opt := option{self.speechFileDir, self.sentenceLen, self.limitCh, speaker}
	return GenerateSpeechFiles(text, outputFileName, opt)
}

type byModTime []os.FileInfo

func (s byModTime) Len() int {
	return len(s)
}

func (s byModTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byModTime) Less(i, j int) bool {
	return s[i].ModTime().Before(s[j].ModTime())
}

func (self *TTSManager) ClearSpeechFiles() []os.FileInfo {
	speechFiles := make([]os.FileInfo, 0)
	var totalSize int64 = 0
	err := filepath.Walk(self.speechFileDir, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			totalSize += f.Size()
			speechFiles = append(speechFiles, f)
		}
		return nil
	})
	if err != nil {
		return []os.FileInfo{}
	}

	// start deleting files
	if totalSize > self.speechCache {
		// sort by time
		sort.Sort(byModTime(speechFiles))
		var deletedSize int64 = 0
		removeIndex := 0
		for i, sFile := range speechFiles {
			deletedSize += sFile.Size()
			if totalSize-deletedSize <= self.speechCache {
				removeIndex = i
				break
			}
		}
		deletedFiles := speechFiles[0 : removeIndex+1]
		for _, sFile := range deletedFiles {
			os.Remove(filepath.Join(self.speechFileDir, sFile.Name()))
		}
		return deletedFiles
	}
	return []os.FileInfo{}
}
