package tts

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

var SpeechFileDir = ""

const (
	NEOSPEECH_URL = "http://208.109.168.116/GetAudio1.ashx?speaker=%d&content=%s"
)
const (
	MALE   = 203
	FEMALE = 202
)

func GetSpeech(sentence string, speaker int, fileName string) error {
	url := fmt.Sprintf(NEOSPEECH_URL, speaker, sentence)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/29.0.1547.66 Safari/537.36")
	resp, httpErr := http.DefaultClient.Do(req)
	if httpErr != nil {
		return httpErr
	}
	defer resp.Body.Close()

	out, createFileErr := os.Create(fileName + ".mp3")
	if createFileErr != nil {
		return createFileErr
	}
	defer out.Close()

	_, copyErr := io.Copy(out, resp.Body)
	return copyErr
}
