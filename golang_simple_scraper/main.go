package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return
}

func crawl(url string, ch chan string, chFinished chan bool) {
	res, err := http.Get(url)

	defer func() {
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("Error: failed to crawl", url)
		return
	}

	b := res.Body

	defer b.Close()

	z := html.NewTokenizer(b)
	for {
		tt := z.Next()
		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}
			ok, url := getHref(t)
			if !ok {
				continue
			}

			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {
	foundUrls := make(map[string]bool)
	// we can alse declare map like this:
	// foundUrls := map[string]bool{};
	seedUrls := os.Args[1:]

	chUrls := make(chan string)
	chFinished := make(chan bool)

	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	fmt.Print("\nFound", len(foundUrls), "Unique urls:\n")
	for url, _ := range foundUrls {
		fmt.Println("-" + url)
	}
	close(chUrls)
}
