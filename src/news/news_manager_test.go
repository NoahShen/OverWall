package news

import (
	"fmt"
	"testing"
	"time"
)

func TestGetVoiceNews(t *testing.T) {
	opt := Option{"/home/noah/workspace/OverWall/news_speech_file/",
		1024 * 1024 * 1,
		500,
		4,
		"piassistant87@163.com",
		"15935787",
		false}
	newsManager := NewNewsManager(opt)
	voiceNewses, getNewsErr := newsManager.GetVoiceNews(30, filterOverLenNews)
	if getNewsErr != nil {
		t.Fatal(getNewsErr)
	}
	//for _, news := range voiceNewses {
	//	news.Play()
	//}
	fmt.Printf("news count:%d\n", len(voiceNewses))
}

func filterOverLenNews(news *VoiceNews) bool {
	return len(news.Content) > 5000
}

func _TestPlayVoiceNews(t *testing.T) {
	voiceNews := &VoiceNews{}
	voiceNews.VoiceFile = "/home/noah/workspace/OverWall/news_speech_file/上海夫妻俩侍奉8位老人30年 除自家父母还有邻居孤老-{8203d388622246548c416b05f309a8a1_1}.mp3"
	go func() {
		err := voiceNews.Play()
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(10e9)

	voiceNews.Stop()
}
