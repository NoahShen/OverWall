package easyread

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"utils"
)

const (
	LOGIN_URL                 = "https://easyread.163.com/sns/login/login.atom"
	GET_SUBSUMMARY_URL        = "http://easyread.163.com/user/subsummary.atom?rand=%d"
	GET_SUBSUMMARY_SOURCE_URL = "http://easyread.163.com/news/source/index.atom?id=%s"
	GET_ARTICLE_URL           = "http://cdn.easyread.163.com/news/article.atom?uuid=%s"
)

type subSummary struct {
	XMLName xml.Name          `xml:"usrsubsummary"`
	Entries []subSummaryEntry `xml:"entry"`
}

type subSummaryEntry struct {
	XMLName xml.Name    `xml:"entry"`
	Id      string      `xml:"id"`
	Name    string      `xml:"title"`
	Status  entryStatus `xml:"entry_status"`
}

type entryStatus struct {
	XMLName xml.Name `xml:"entry_status"`
	Type    string   `xml:"type,attr"`
	Style   string   `xml:"style,attr"`
}

type newsFeed struct {
	XMLName     xml.Name    `xml:"feed"`
	Id          string      `xml:"id"`
	Title       string      `xml:"title"`
	UpdatedDate string      `xml:"updated"`
	Entries     []newsEntry `xml:"entry"`
}

type newsEntry struct {
	XMLName      xml.Name     `xml:"entry"`
	Id           string       `xml:"id"`
	Title        string       `xml:"title"`
	Author       string       `xml:"author>name"`
	UpdatedDate  string       `xml:"updated"`
	EntryContent entryContent `xml:"content"`
}

type entryContent struct {
	XMLName     xml.Name `xml:"content"`
	Content     string   `xml:",chardata"`
	ContentType string   `xml:"type,attr"`
}

type articleFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Id      string       `xml:"id"`
	Title   string       `xml:"title"`
	Entry   articleEntry `xml:"entry"`
}

type articleEntry struct {
	XMLName     xml.Name       `xml:"entry"`
	Id          string         `xml:"id"`
	Title       string         `xml:"title"`
	UpdatedDate string         `xml:"updated"`
	Content     articleContent `xml:"content"`
}

type articleContent struct {
	XMLName     xml.Name `xml:"content"`
	Content     string   `xml:",chardata"`
	ContentType string   `xml:"type,attr"`
}

type NewsSub struct {
	Id   string
	Name string
	Type string
}

type NewsArticle struct {
	Id              string
	Title           string
	UpdatedDate     string
	ContentType     string
	Content         string
	OriginalContent string
}

type EasyreadSession struct {
	UserId   string
	Username string
	cookies  []*http.Cookie
}

func CreateEasyreadSession(username, password string) (*EasyreadSession, error) {
	easyreadSession := &EasyreadSession{}
	loginInfo := make(map[string]interface{})
	loginInfo["accountType"] = 0
	loginInfo["auth"] = utils.MD5Encode(password)
	loginInfo["username"] = username
	loginInfoJson, _ := json.Marshal(loginInfo)
	b := strings.NewReader(string(loginInfoJson))
	req := easyreadSession.createHttpRequest("POST", LOGIN_URL, "application/json", b)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	content, respErr := easyreadSession.getResponseContent(resp)
	if respErr != nil {
		return nil, respErr
	}
	loginResult := make(map[string]interface{})
	unmarshalErr := json.Unmarshal(content, &loginResult)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	resCode := loginResult["resCode"].(float64)
	if resCode != 0 {
		return nil, errors.New("resCode is not zero! username or password is invalid!")
	}
	userInfo := loginResult["userInfo"].(map[string]interface{})
	easyreadSession.Username = userInfo["urs"].(string)
	easyreadSession.UserId = userInfo["userId"].(string)

	cookies := resp.Cookies()
	easyreadSession.cookies = cookies

	return easyreadSession, nil
}

func (self *EasyreadSession) GetNewsSub() ([]NewsSub, error) {
	subs := make([]NewsSub, 0)
	subSummary, getSummaryErr := self.getSubSummary()
	if getSummaryErr != nil {
		return subs, getSummaryErr
	}
	for _, entry := range subSummary.Entries {
		id := entry.Id
		name := entry.Name
		subType := entry.Status.Type
		newsSub := NewsSub{id, name, subType}
		subs = append(subs, newsSub)
	}
	return subs, nil
}

func (self *EasyreadSession) GetNewsArticles(newsSub NewsSub) ([]NewsArticle, error) {
	articles := make([]NewsArticle, 0)
	newsFeed, newsSourceErr := self.getNewsSource(newsSub.Id)
	if newsSourceErr != nil {
		return articles, newsSourceErr
	}

	for _, newsEntry := range newsFeed.Entries {
		articleFeed, getArticleErr := self.getArticle(newsEntry.Id)
		if getArticleErr != nil {
			return articles, getArticleErr
		}
		id := articleFeed.Entry.Id
		title := articleFeed.Entry.Title
		updatedDate := articleFeed.Entry.UpdatedDate

		contentType := articleFeed.Entry.Content.ContentType
		originalContent := articleFeed.Entry.Content.Content
		var content string
		if contentType == "xhtml" || contentType == "html" {
			var parseErr error
			content, parseErr = self.parseContent(newsSub.Type, originalContent)
			if parseErr != nil {
				return articles, parseErr
			}
		} else {
			content = originalContent
		}

		newsArticle := NewsArticle{id, title, updatedDate, contentType, content, originalContent}
		articles = append(articles, newsArticle)
	}
	return articles, nil
}

func (self *EasyreadSession) getArticle(atricleId string) (articleFeed, error) {
	var articleFeed = articleFeed{}
	url := fmt.Sprintf(GET_ARTICLE_URL, atricleId)
	req := self.createHttpRequest("GET", url, "", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return articleFeed, err
	}
	content, respErr := self.getResponseContent(resp)
	if respErr != nil {
		return articleFeed, respErr
	}
	unmarshalErr := xml.Unmarshal(content, &articleFeed)
	if unmarshalErr != nil {
		return articleFeed, unmarshalErr
	}
	return articleFeed, nil
}

func (self *EasyreadSession) getNewsSource(summaryEntryId string) (newsFeed, error) {
	var newsFeed = newsFeed{}
	url := fmt.Sprintf(GET_SUBSUMMARY_SOURCE_URL, summaryEntryId)
	req := self.createHttpRequest("GET", url, "", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return newsFeed, err
	}
	content, respErr := self.getResponseContent(resp)
	if respErr != nil {
		return newsFeed, respErr
	}
	unmarshalErr := xml.Unmarshal(content, &newsFeed)
	if unmarshalErr != nil {
		return newsFeed, unmarshalErr
	}
	return newsFeed, nil
}

func (self *EasyreadSession) getSubSummary() (subSummary, error) {
	var subSummar = subSummary{}
	url := fmt.Sprintf(GET_SUBSUMMARY_URL, time.Now().UTC().Unix())
	req := self.createHttpRequest("GET", url, "", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return subSummar, err
	}
	content, respErr := self.getResponseContent(resp)
	if respErr != nil {
		return subSummar, respErr
	}
	unmarshalErr := xml.Unmarshal(content, &subSummar)
	if unmarshalErr != nil {
		return subSummar, unmarshalErr
	}
	return subSummar, nil
}

func (self *EasyreadSession) createHttpRequest(method, url, contentType string, body io.Reader) *http.Request {
	req, _ := http.NewRequest(method, url, body)
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("User-Agent", "Pris/3.0.0")
	req.Header.Set("X-User-Agent", "PRIS/3.0.0 (768/1184; android; 4.3; zh-CN; android) 1.2.8")
	req.Header.Set("X-Pid", "(000000000000000;d41d8cd98f00b204e9800998ecf8427e;)")
	for _, cookie := range self.cookies {
		req.AddCookie(cookie)
	}
	return req
}

func (self *EasyreadSession) getResponseContent(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (self *EasyreadSession) parseContent(subType, htmlContent string) (string, error) {
	doc, newDocErr := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if newDocErr != nil {
		return "", newDocErr
	}
	var content string
	if subType == "news" {
		contentObj := doc.Find("div.fs-content")
		content = contentObj.Text()
	} else if subType == "mblog" {
		contentObj := doc.Find(".fs-ori-content")
		if contentObj.Size() == 0 {
			contentObj = doc.Find(".fs-content")
		}
		content = contentObj.Text()
	}
	return strings.Replace(content, " ", "", -1), nil
}
