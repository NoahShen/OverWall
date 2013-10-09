package news

import (
	"easyread"
	"fmt"
	"os/exec"
	"sort"
	"time"
	"tts"
)

type VoiceNews struct {
	Id          string
	Title       string
	Type        string
	UpdatedTime time.Time
	Content     string
	VoiceFile   string
	VoiceStatCh chan int
}

type byUpdateTime []*VoiceNews

func (s byUpdateTime) Len() int {
	return len(s)
}

func (s byUpdateTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byUpdateTime) Less(i, j int) bool {
	return s[i].UpdatedTime.After(s[j].UpdatedTime)
}

type filterFunc func(news *VoiceNews) bool

type Option struct {
	SpeechFileDir    string
	SpeechCache      int64
	SentenceLen      int
	MaxGenVoiceTask  int
	EasyreadUsername string
	EasyreadPwd      string
}

type NewsManager struct {
	ttsManager *tts.TTSManager
	opt        Option
}

func NewNewsManager(opt Option) *NewsManager {
	ttsManager := tts.NewTTSManager(opt.SpeechFileDir, opt.SentenceLen, opt.MaxGenVoiceTask, opt.SpeechCache)
	newsManager := &NewsManager{ttsManager, opt}
	return newsManager
}

func (self *NewsManager) GetVoiceNews(limit int, filter filterFunc) ([]*VoiceNews, error) {
	session, createSeesErr := easyread.CreateEasyreadSession(self.opt.EasyreadUsername, self.opt.EasyreadPwd)
	if createSeesErr != nil {
		return []*VoiceNews{}, createSeesErr
	}
	subs, getSubErr := session.GetNewsSub()
	if getSubErr != nil {
		return []*VoiceNews{}, getSubErr
	}

	voiceNewses := make([]*VoiceNews, 0)
	for _, sub := range subs {
		newsType := sub.Type
		articles, getArticlesErr := session.GetNewsArticles(sub)
		if getArticlesErr != nil {
			return []*VoiceNews{}, getArticlesErr
		}

		for _, article := range articles {
			voiceNews := &VoiceNews{}
			voiceNews.Id = article.Id
			voiceNews.Type = newsType
			voiceNews.Title = article.Title
			voiceNews.UpdatedTime = article.UpdatedDate
			voiceNews.Content = article.Content
			voiceNews.VoiceStatCh = make(chan int, 1)
			if filter != nil && filter(voiceNews) {
				continue
			}
			voiceNewses = append(voiceNewses, voiceNews)
		}
	}
	sort.Sort(byUpdateTime(voiceNewses))
	if limit > 0 && limit < len(voiceNewses) {
		voiceNewses = voiceNewses[0:limit]
	}
	for i, vNews := range voiceNewses {
		var speaker int
		if i%2 == 0 {
			speaker = tts.MALE
		} else {
			speaker = tts.FEMALE
		}
		go self.generateVoiceFile(vNews, speaker)
	}
	return voiceNewses, nil
}

func (self *NewsManager) generateVoiceFile(vNews *VoiceNews, speaker int) {
	fileName := fmt.Sprintf("%s-{%s}.mp3", vNews.Title, vNews.Id)
	voiceFile, err := self.ttsManager.GenerateSpeechFiles(vNews.Content, fileName, speaker)
	if err != nil {
		vNews.VoiceStatCh <- -1
		return
	}
	vNews.VoiceFile = voiceFile
	vNews.VoiceStatCh <- 1
}

type NewsPlay struct {
	cmd       *exec.Cmd
	VoiceNews *VoiceNews
}

func CreateNewsPlay(vNews *VoiceNews) *NewsPlay {
	play := &NewsPlay{}
	play.cmd = exec.Command("mpg321", vNews.VoiceFile)
	play.VoiceNews = vNews
	return play
}

func (self *NewsPlay) Play() error {
	b, playErr := self.cmd.Output()
	if playErr != nil {
		return playErr
	}
	fmt.Println(string(b))
	return nil
}

func (self *NewsPlay) Stop() error {
	return self.cmd.Process.Kill()
}
