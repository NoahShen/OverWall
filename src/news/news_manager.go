package news

import (
	l4g "code.google.com/p/log4go"
	"easyread"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"time"
	"tts"
	"utils"
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
	GenVoiceFile     bool
}

type NewsManager struct {
	ttsManager   *tts.TTSManager
	opt          Option
	easyReadSess *easyread.EasyreadSession
}

func NewNewsManager(opt Option) *NewsManager {
	ttsManager := tts.NewTTSManager(opt.SpeechFileDir, opt.SentenceLen, opt.MaxGenVoiceTask, opt.SpeechCache)
	newsManager := &NewsManager{ttsManager, opt, nil}
	return newsManager
}

func (self *NewsManager) GetVoiceNews(limit int, filter filterFunc) ([]*VoiceNews, error) {
	if self.easyReadSess == nil {
		l4g.Info("start creating easyread seesion, user=[%s] pwd=[%s]", self.opt.EasyreadUsername, self.opt.EasyreadPwd)
		session, createSeesErr := easyread.CreateEasyreadSession(self.opt.EasyreadUsername, self.opt.EasyreadPwd)
		if createSeesErr != nil {
			return []*VoiceNews{}, createSeesErr
		}
		self.easyReadSess = session
	}
	l4g.Info("get news subscriptions...")
	subs, getSubErr := self.easyReadSess.GetNewsSub()
	if getSubErr != nil {
		return []*VoiceNews{}, getSubErr
	}

	voiceNewses := make([]*VoiceNews, 0)
	for _, sub := range subs {
		newsType := sub.Type
		l4g.Debug("get news articles, sub_id=[%s], name=[%s], type=[%s]", sub.Id, sub.Name, sub.Type)
		articles, getArticlesErr := self.easyReadSess.GetNewsArticles(sub)
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
	if self.opt.GenVoiceFile {
		l4g.Debug("Start generating voice files.")
		for i, vNews := range voiceNewses {
			var speaker int
			if i%2 == 0 {
				speaker = tts.MALE
			} else {
				speaker = tts.FEMALE
			}
			go self.generateVoiceFile(vNews, speaker)
		}
	}
	return voiceNewses, nil
}

func (self *NewsManager) generateVoiceFile(vNews *VoiceNews, speaker int) {
	fileName := fmt.Sprintf("%s-{%s}.mp3", vNews.Title, vNews.Id)
	filePath := self.opt.SpeechFileDir + fileName
	if utils.Exists(filePath) {
		l4g.Debug("Voice file exist, news_id=[%s], title=[%s]", vNews.Id, vNews.Title)
		vNews.VoiceFile = filePath
		vNews.VoiceStatCh <- 1
		return
	}
	l4g.Debug("Start generating speech, news_id=[%s], title=[%s]", vNews.Id, vNews.Title)
	voiceFile, err := self.ttsManager.GenerateSpeechFiles(vNews.Content, fileName, speaker)
	if err != nil {
		l4g.Error("Generating speech error, news_id=[%s], title=[%s], error: %s", vNews.Id, vNews.Title, err.Error())
		vNews.VoiceStatCh <- -1
		return
	}
	l4g.Debug("Generating speech complete, news_id=[%s], title=[%s]", vNews.Id, vNews.Title)
	vNews.VoiceFile = voiceFile
	vNews.VoiceStatCh <- 1
}

type NewsPlay struct {
	cmd       *exec.Cmd
	VoiceNews *VoiceNews
}

func CreateNewsPlay(vNews *VoiceNews) *NewsPlay {
	play := &NewsPlay{}

	play.VoiceNews = vNews
	return play
}

func (self *NewsPlay) Play() error {
	l4g.Debug("Start playing speech file for news, news_id=[%s], title=[%s]", self.VoiceNews.Id, self.VoiceNews.Title)
	if len(self.VoiceNews.VoiceFile) == 0 {
		l4g.Debug("Waiting for generating voice file complete, news_id=[%s], title=[%s]", self.VoiceNews.Id, self.VoiceNews.Title)
		resp := <-self.VoiceNews.VoiceStatCh
		if resp != 1 {
			l4g.Error("Generating voice file error, news_id=[%s], title=[%s]", self.VoiceNews.Id, self.VoiceNews.Title)
			return errors.New("Generate voice file error!")
		}
	}
	self.cmd = exec.Command("mpg321", self.VoiceNews.VoiceFile)
	_, playErr := self.cmd.Output()
	if playErr != nil {
		return playErr
	}
	return nil
}

func (self *NewsPlay) Stop() error {
	l4g.Debug("Stop playing speech file for news, news_id=[%s], title=[%s]", self.VoiceNews.Id, self.VoiceNews.Title)
	if self.cmd != nil {
		err := self.cmd.Process.Kill()
		self.cmd = nil
		return err
	}
	return nil
}
