package news

import (
	"fmt"
	"testing"
	"time"
)

func _TestGetVoiceNews(t *testing.T) {
	voiceNewses, getNewsErr := GetVoiceNews(10, filterOverLenNews)
	if getNewsErr != nil {
		t.Fatal(getNewsErr)
	}
	for _, news := range voiceNewses {
		fmt.Printf("title: %s %s\nupdate_time: %s\ncontent len: %d\n", news.Title, news.Type, news.UpdatedTime, len(news.Content))
		resp := <-news.VoiceStatCh
		if resp == 1 {
			fmt.Printf("voiceFile: %s\n=======\n", news.VoiceFile)
		} else {
			fmt.Printf("generate voice file error!\n=======\n")
		}
	}
	fmt.Printf("news count:%d\n", len(voiceNewses))
}

func filterOverLenNews(news *VoiceNews) bool {
	return len(news.Content) > 5000
}

func TestPlayVoiceNews(t *testing.T) {
	voiceNews := &VoiceNews{}
	voiceNews.VoiceFile = "/home/noah/workspace/OverWall/news_speech_file/杭州萧山一越野车冲入河中致3人溺亡-{4cb77afd5e7d4fd1995b48991652b779_1}.mp3"
	play := CreateNewsPlay(voiceNews)
	go play.Play()

	time.Sleep(10e9)

	play.Stop()
}
