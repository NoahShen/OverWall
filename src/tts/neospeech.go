package tts

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"
	"utils"
)

var SpeechFileDir = "/home/noah/workspace/OverWall/news_speech_file/"

var SentenceLen = 500

var GenerateSpeechTimeout time.Duration = 30 * time.Second

var TaskLimit = 4

const (
	NEOSPEECH_URL = "http://208.109.168.116/GetAudio1.ashx?speaker=%d&content=%s"

	MERGE_MP3_ARG = `"concat:%s`
)
const (
	MALE   = 203
	FEMALE = 202
)

func GenerateSpeechFiles(sentence string, speaker int) (string, error) {
	sentences := SplitSentence(sentence, SentenceLen)
	sLen := len(sentences)
	files := make([]string, 0)
	filePrefix := utils.RandomString(7)
	reply := make(chan int, sLen)
	timeout := false
	hasError := false
	limitCh := make(chan int, TaskLimit)
	for i, sentence := range sentences {
		limitCh <- 1
		fileName := fmt.Sprintf("%s%s-%d.mp3", SpeechFileDir, filePrefix, i)
		go getSpeech(sentence, speaker, fileName, reply, limitCh)
		files = append(files, fileName)
	}
	for i := 0; i < sLen; i++ {
		select {
		case <-time.After(time.Duration(GenerateSpeechTimeout)):
			timeout = true
			break
		case r := <-reply:
			if r == -1 {
				hasError = true
				break
			}
		}
	}
	if timeout {
		return "", errors.New("Generate speech timeout!")
	} else if hasError {
		return "", errors.New("Generate speech error!")
	}
	return mergeSpeechFiles(files)
}

func mergeSpeechFiles(files []string) (string, error) {
	fileNames := ""
	for _, name := range files {
		fileNames += name + "|"
		defer os.Remove(name)
	}
	outputFileName := SpeechFileDir + utils.RandomString(10) + ".mp3"
	fileArgs := fmt.Sprintf("concat:%s", fileNames)
	//ffmpeg -i "concat:file1.mp3|file2.mp3" -acodec copy output.mp3
	cmd := exec.Command("ffmpeg", "-i", fileArgs, "-acodec", "copy", outputFileName)
	var out bytes.Buffer
	cmd.Stdout = &out
	mergeErr := cmd.Run()
	if mergeErr != nil {
		return "", mergeErr
	}
	for _, name := range files {
		fileNames += name + "|"
	}
	return outputFileName, nil
}

func getSpeech(sentence string, speaker int, fileName string, reply chan<- int, limitCh <-chan int) error {
	defer func() { <-limitCh }()
	speechUrl := fmt.Sprintf(NEOSPEECH_URL, speaker, url.QueryEscape(sentence))
	req, _ := http.NewRequest("GET", speechUrl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1547.66 Safari/537.36")
	resp, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		reply <- -1
		return httpErr
	}
	defer resp.Body.Close()

	out, createFileErr := os.Create(fileName)
	if createFileErr != nil {
		reply <- -1
		return createFileErr
	}
	defer out.Close()

	_, copyErr := io.Copy(out, resp.Body)
	if copyErr != nil {
		reply <- -1
		return copyErr
	}
	reply <- 1
	return nil
}
