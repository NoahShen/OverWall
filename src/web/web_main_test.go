package web

import (
	l4g "code.google.com/p/log4go"
	"news"
	"testing"
)

func TestWebManager(t *testing.T) {
	l4g.LoadConfiguration("../../config/logConfig.xml")

	opt := news.Option{"/home/noah/workspace/OverWall/news_speech_file/",
		1024 * 1024 * 1, // 1MB
		500,
		4,
		"piassistant87@163.com",
		"15935787",
		true}
	newsManager := news.NewNewsManager(opt)
	webManager := NewWebManager(newsManager, "8787")
	webManager.StartServer()
}
