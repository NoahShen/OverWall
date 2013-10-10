package web

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"news"
)

var newsManager *news.NewsManager

func initNewsManager() {
	opt := news.Option{"/home/noah/workspace/OverWall/news_speech_file/",
		1024 * 1024 * 1,
		500,
		4,
		"piassistant87@163.com",
		"15935787",
		false}
	newsManager = news.NewNewsManager(opt)
}

func showMainPageHandler(w http.ResponseWriter, req *http.Request) {
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

func getNewsHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("start getting news")
	voiceNewses, getNewsErr := newsManager.GetVoiceNews(10, filterOverLenNews)
	fmt.Println("getting news finished!")
	result := make(map[string]interface{})
	fmt.Println("getNewsErr:", getNewsErr)
	if getNewsErr != nil {
		result["result"] = "error"
		result["errorMessage"] = getNewsErr.Error()
	} else {
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
	}

	b, _ := json.Marshal(result)
	fmt.Println("response content:", string(b))
	w.Write(b)
}

func filterOverLenNews(news *news.VoiceNews) bool {
	return len(news.Content) > 5000
}

func StartGorilla() {
	initNewsManager()
	r := mux.NewRouter()
	r.HandleFunc("/news", showMainPageHandler)
	r.HandleFunc("/news/getnews", getNewsHandler)
	http.Handle("/", r)
	http.ListenAndServe(":8787", nil)
}
