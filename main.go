package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Player struct {
	name string

	batRuns    int
	balls      int
	four       int
	six        int
	strikeRate int

	overs   int
	maidens int
	oRuns   int
	wickets int
	economy int
}

type Team struct {
	name    string
	short   string
	players []Player
	state   bool
}

type Match struct {
	t1        Team
	t2        Team
	score     string // Short Summary
	state     string // finished, 2nd innings, 1st innings, toss, upcoming
	tossState bool   // 0 -> First t1 bat, t2 bowl, First 1 -> t1 bowl, t2 bat
}

const cricURL = "https://www.cricbuzz.com"

func httpReq(url string) io.ReadCloser {
	resp, err := http.Get(url)
	if err != nil {
		log.Panic(err)
	}

	return resp.Body
}

func parseResp(resp io.ReadCloser) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(resp)
	if err != nil {
		log.Panic(err)
	}

	return doc
}

func getMatch() string {
	matchLink := cricURL

	resp := httpReq(cricURL + "/live-scores")
	doc := parseResp(resp)
	fmt.Println("Parsing live scores")

	doc.Find("[ng-show~=\"active_match_type\"] a[href*=india].cb-lv-scrs-well-live").Each(func(i int, s *goquery.Selection) {
		path, _ := s.Attr("href")
		matchLink += path
	})

	return matchLink
}

func initMatch(url string) Match {
	var m Match

	resp := httpReq(url)
	doc := parseResp(resp)

	doc.Find(".cb-team-lft-item h1").Each(func(i int, s *goquery.Selection) {
		r := regexp.MustCompile(",.*")
		str := strings.Split(r.ReplaceAllString(s.Text(), ""), " vs ")

		m.t1.name = str[0]
		m.t2.name = str[1]
	})

	doc.Find("div.cb-min-bat-rw span.text-bold.cb-font-20").Each(func(i int, s *goquery.Selection) {
		m.score = s.Text()
	})

	resp = httpReq(strings.Replace(url, "scores", "scorecard", -1))
	doc = parseResp(resp)

	return m
}

func main() {
	url := getMatch()
	m := initMatch(url)
	fmt.Println(m.score + "\n" + m.t1.name + " vs " + m.t2.name)
}
