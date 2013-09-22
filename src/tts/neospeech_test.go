package tts

import (
	"testing"
)

func TestSentence(t *testing.T) {
	err := GetSpeech("你好，謝謝你測試我的聲音，請輸入你要念的文字。", FEMALE, "test")
	if err != nil {
		t.Fatal(err)
	}
}
