package web

import (
	l4g "code.google.com/p/log4go"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"news"
	"strconv"
)

type WebManager struct {
	newsManager *news.NewsManager
	port        string
	attributes  map[string]interface{}
	currentPlay *news.NewsPlay
	stopPlayCh  chan int
}

func NewWebManager(newsManager *news.NewsManager, port string) *WebManager {
	webManager := &WebManager{newsManager,
		port,
		make(map[string]interface{}),
		nil,
		make(chan int, 1)}
	return webManager
}

func showMainPageHandler(w http.ResponseWriter, req *http.Request, webManager *WebManager) {
	t, err := template.ParseFiles("./pages/main.html")
	if err != nil {
		l4g.Error("template.ParseFiles error:%s", err.Error())
		return
	}
	execErr := t.Execute(w, nil)
	if execErr != nil {
		l4g.Error("template.Execute error:%s", execErr.Error())
	}
}

func getNewsHandler(w http.ResponseWriter, req *http.Request, webManager *WebManager) {
	refresh := req.FormValue("refresh")
	l4g.Debug("param refresh=%s", refresh)
	var voiceNewses []*news.VoiceNews
	newses, ok := webManager.attributes["voiceNewses"]
	if ok && refresh != "1" {
		l4g.Debug("Getting latest news from memory")
		voiceNewses = newses.([]*news.VoiceNews)
	} else {
		limitStr := req.FormValue("limit")
		l4g.Debug("param limit=%s", limitStr)
		newsLimit := 0
		if len(limitStr) == 0 {
			newsLimit = 10
		} else {
			newsLimit, _ = strconv.Atoi(limitStr)
		}
		l4g.Debug("newsLimit=%d", newsLimit)
		vNewses, getNewsErr := webManager.newsManager.GetVoiceNews(newsLimit, filterOverLenNews)
		if getNewsErr != nil {
			result := make(map[string]interface{})
			result["result"] = "error"
			result["errorMessage"] = getNewsErr.Error()
			writeJsonResponse(w, result)
			return
		}
		voiceNewses = vNewses
		webManager.attributes["voiceNewses"] = voiceNewses
	}

	result := make(map[string]interface{})
	result["result"] = "success"
	allnews := make([]map[string]string, 0)
	for _, news := range voiceNewses {
		oneNews := make(map[string]string)
		oneNews["id"] = news.Id
		oneNews["title"] = news.Title
		oneNews["updatedTime"] = news.UpdatedTime.Format("2006-01-02 15:04:05")
		allnews = append(allnews, oneNews)
	}
	result["news"] = allnews
	l4g.Debug(func() string {
		marshalString, _ := json.MarshalIndent(result, "", "    ")
		return fmt.Sprintf("response:\n%s", string(marshalString))
	})
	writeJsonResponse(w, result)
}

func writeJsonResponse(w http.ResponseWriter, result map[string]interface{}) {
	b, _ := json.Marshal(result)
	w.Write(b)
}

func filterOverLenNews(news *news.VoiceNews) bool {
	return len(news.Content) > 5000
}

func playNewsHandler(w http.ResponseWriter, req *http.Request, webManager *WebManager) {

	play := webManager.currentPlay
	result := make(map[string]interface{})
	if play != nil {
		l4g.Debug("Already start playing!")
		result["result"] = "success"
		writeJsonResponse(w, result)
		return
	}

	newses, ok := webManager.attributes["voiceNewses"]
	if !ok {
		result["result"] = "error"
		result["errorMessage"] = "news not exist!"
		writeJsonResponse(w, result)
		return
	}

	voiceNewses := newses.([]*news.VoiceNews)
	go playingNews(webManager, voiceNewses)

	result["result"] = "success"
	writeJsonResponse(w, result)
}

func playingNews(webManager *WebManager, voiceNewses []*news.VoiceNews) {
	l4g.Debug("Start playing")
	for _, vNews := range voiceNewses {
		select {
		case <-webManager.stopPlayCh:
			webManager.currentPlay = nil
			return
		default:
			play := news.CreateNewsPlay(vNews)
			webManager.currentPlay = play
			play.Play()
		}

	}
	webManager.currentPlay = nil

}

func stopPlayNewsHandler(w http.ResponseWriter, req *http.Request, webManager *WebManager) {
	webManager.stopPlayCh <- 1
	if webManager.currentPlay != nil {
		webManager.currentPlay.Stop()
	}
}

func playNextNewsHandler(w http.ResponseWriter, req *http.Request, webManager *WebManager) {
	if webManager.currentPlay != nil {
		webManager.currentPlay.Stop()
	}
}

func (self *WebManager) StartServer() {
	l4g.Info("start http server on port=[%s]", self.port)
	r := mux.NewRouter()
	r.HandleFunc("/news", func(w http.ResponseWriter, r *http.Request) { showMainPageHandler(w, r, self) })
	r.HandleFunc("/news/getnews", func(w http.ResponseWriter, r *http.Request) { getNewsHandler(w, r, self) })
	r.HandleFunc("/news/play", func(w http.ResponseWriter, r *http.Request) { playNewsHandler(w, r, self) })
	r.HandleFunc("/news/stop", func(w http.ResponseWriter, r *http.Request) { stopPlayNewsHandler(w, r, self) })
	r.HandleFunc("/news/next", func(w http.ResponseWriter, r *http.Request) { playNextNewsHandler(w, r, self) })
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("pages/static/"))))
	http.Handle("/", r)
	http.ListenAndServe(":"+self.port, nil)
}
