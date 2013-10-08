package tts

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestGernateSpeechFiles(t *testing.T) {
	b, readFileErr := ioutil.ReadFile("/home/noah/workspace/OverWall/news_speech_file//news_content_testfile.txt")
	if readFileErr != nil {
		t.Fatal(readFileErr)
	}
	fileNames, err := GenerateSpeechFiles(string(b), "news.mp3", FEMALE)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fileNames)
}
