package web

import (
	l4g "code.google.com/p/log4go"
	"encoding/json"
	"fmt"
	"news"
	"os"
	"testing"
)

func TestWebManager(t *testing.T) {
	l4g.LoadConfiguration("../../config/logConfig.xml")

	configFile, readConfigFileErr := os.Open("../../config/overwall.config")
	if readConfigFileErr != nil {
		t.Fatal(readConfigFileErr)
	}
	decoder := json.NewDecoder(configFile)
	config := &Config{}
	decoder.Decode(&config)

	fmt.Printf("config info:\n%+v\n", config)

	opt := news.Option{}
	opt.SpeechFileDir = config.SpeechFileDir
	opt.SpeechCache = config.CacheSize
	opt.SentenceLen = config.OneSentenceLen
	opt.MaxGenVoiceTask = config.MaxGenVoiceTask
	opt.EasyreadUsername = config.EasyreadUsername
	opt.EasyreadPwd = config.EasyreadPwd
	opt.GenVoiceFile = true

	fmt.Printf("opt info:\n%+v\n", opt)

	newsManager := news.NewNewsManager(opt)
	webManager := NewWebManager(newsManager, "8787", "/home/noah/workspace/OverWall/config/chimes1.mp3")
	webManager.StartServer()
}
