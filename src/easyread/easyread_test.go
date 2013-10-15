package easyread

import (
	"fmt"
	"testing"
)

func _TestEasyLogin(t *testing.T) {
	_, err := CreateEasyreadSession("piassistant87@163.com", "15935787")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetSubSummary(t *testing.T) {
	session, err := CreateEasyreadSession("piassistant87@163.com", "15935787")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("userId:", session.UserId)
	fmt.Println("username:", session.Username)
	subs, getSubErr := session.GetNewsSub()
	if getSubErr != nil {
		t.Fatal(getSubErr)
	}

	for _, sub := range subs {
		fmt.Printf("sub:%+v\n", sub)
		if sub.Type == "news" {
			articles, getArticlesErr := session.GetNewsArticles(sub)
			if getArticlesErr != nil {
				t.Fatal(getArticlesErr)
			}
			for _, article := range articles {
				fmt.Printf("article:\n%+v\n", article.Content)
			}
		}
	}
}
