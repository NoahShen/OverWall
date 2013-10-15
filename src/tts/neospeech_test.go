package tts

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func _TestGernateSpeechFiles(t *testing.T) {
	b, readFileErr := ioutil.ReadFile("/home/noah/workspace/OverWall/news_speech_file/news_content_testfile.txt")
	if readFileErr != nil {
		t.Fatal(readFileErr)
	}
	limitCh := make(chan int, 4)
	opt := option{"/home/noah/workspace/OverWall/news_speech_file", 500, limitCh, FEMALE}
	fileNames, err := GenerateSpeechFiles(string(b), "news.mp3", opt)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(fileNames)
}
