package tts

import (
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

var GenerateSpeechTimeout time.Duration = 60 * time.Second

const (
	allowErrTimes = 3
	NEOSPEECH_URL = "http://208.109.168.116/GetAudio1.ashx?speaker=%d&content=%s"
)
const (
	MALE   = 203
	FEMALE = 202
)

type option struct {
	SpeechFileDir string
	SentenceLen   int
	LimitCh       chan int
	Speaker       int
}

func GenerateSpeechFiles(sentence, outputFileName string, option option) (string, error) {
	sentences := SplitSentence(sentence, option.SentenceLen)
	files := make([]string, 0)
	var genErr error
	for i := 0; i < allowErrTimes; i++ {
		files, genErr = doGenerateSpeechFiles(sentences, option)
		if genErr == nil {
			return mergeSpeechFiles(files, option.SpeechFileDir, outputFileName)
		}
		fmt.Println("error occour:", genErr, " try again!")
	}
	return "", genErr
}

func doGenerateSpeechFiles(sentences []string, option option) ([]string, error) {
	sLen := len(sentences)
	files := make([]string, 0)
	filePrefix := utils.RandomString(16)
	reply := make(chan int, sLen)
	timeout := false
	hasError := false
	for i, sentence := range sentences {
		option.LimitCh <- 1
		fileName := fmt.Sprintf("%s%s-%d.mp3", option.SpeechFileDir, filePrefix, i)
		go getSpeech(sentence, option.Speaker, fileName, reply, option.LimitCh)
		files = append(files, fileName)
	}
L:
	for i := 0; i < sLen; i++ {
		select {
		case <-time.After(time.Duration(GenerateSpeechTimeout)):
			timeout = true
			break L
		case r := <-reply:
			if r == -1 {
				hasError = true
				break L
			}
		}
	}
	if timeout {
		return []string{}, errors.New("Generate speech timeout!")
	} else if hasError {
		return []string{}, errors.New("Generate speech error!")
	}
	return files, nil
}

func mergeSpeechFiles(files []string, speechFileDir, outputFileName string) (string, error) {
	outputFilePath := speechFileDir + outputFileName

	if len(files) == 1 {
		renameErr := os.Rename(files[0], outputFilePath)
		return outputFilePath, renameErr
	}

	fileNames := ""
	for _, name := range files {
		fileNames += name + "|"
		defer os.Remove(name)
	}

	fileArgs := fmt.Sprintf("concat:%s", fileNames)
	//ffmpeg -i "concat:file1.mp3|file2.mp3" -acodec copy output.mp3
	//fmt.Println("ffmpeg", "-i", fileArgs, "-acodec", "copy", outputFilePath)
	cmd := exec.Command("ffmpeg", "-i", fileArgs, "-acodec", "copy", outputFilePath)
	_, mergeErr := cmd.Output()
	if mergeErr != nil {
		fmt.Println("merge error!", mergeErr)
		return "", mergeErr
	}
	return outputFilePath, nil
}

func getSpeech(sentence string, speaker int, fileName string, reply chan<- int, limitCh <-chan int) error {
	defer func() { <-limitCh }()
	speechUrl := fmt.Sprintf(NEOSPEECH_URL, speaker, url.QueryEscape(sentence))
	req, _ := http.NewRequest("GET", speechUrl, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1547.66 Safari/537.36")
	resp, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		fmt.Println("sentence:", sentence)
		fmt.Println("httpErr:", httpErr)
		reply <- -1
		return httpErr
	}
	defer resp.Body.Close()

	out, createFileErr := os.Create(fileName)
	if createFileErr != nil {
		fmt.Println("sentence:", sentence)
		fmt.Println("createFileErr:", createFileErr)
		reply <- -1
		return createFileErr
	}
	defer out.Close()

	_, copyErr := io.Copy(out, resp.Body)
	if copyErr != nil {
		fmt.Println("sentence:", sentence)
		fmt.Println("copyErr:", copyErr)
		reply <- -1
		return copyErr
	}
	reply <- 1
	return nil
}
