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
}

func NewWebManager(newsManager *news.NewsManager, port string) *WebManager {
	webManager := &WebManager{newsManager, port, make(map[string]interface{})}
	return webManager
}

func showMainPageHandler(w http.ResponseWriter, req *http.Request, webManager *WebManager) {
	t, err := template.ParseFiles("./pages/main.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	execErr := t.Execute(w, nil)
	if execErr != nil {
		fmt.Println(execErr)
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
			b, _ := json.Marshal(result)
			w.Write(b)
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
	b, _ := json.Marshal(result)
	l4g.Debug(func() string {
		marshalString, _ := json.MarshalIndent(result, "", "    ")
		return fmt.Sprintf("response:\n%s", string(marshalString))
	})
	w.Write(b)
}

func filterOverLenNews(news *news.VoiceNews) bool {
	return len(news.Content) > 5000
}

func (self *WebManager) StartServer() {
	l4g.Info("start http server on port=[%s]", self.port)
	r := mux.NewRouter()
	r.HandleFunc("/news", func(w http.ResponseWriter, r *http.Request) { showMainPageHandler(w, r, self) })
	r.HandleFunc("/news/getnews", func(w http.ResponseWriter, r *http.Request) { getNewsHandler(w, r, self) })
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("pages/static/"))))
	http.Handle("/", r)
	http.ListenAndServe(":"+self.port, nil)
}
