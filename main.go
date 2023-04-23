package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gocolly/colly"
)

type Player struct {
	name string

	batorder   int
	runs       int
	balls      int
	four       int
	six        int
	strikeRate float32

	overs   int
	maidens int
	oruns   int
	wickets int
	economy float32
}

type Team struct {
	name    string
	short   string
	players map[int]Player // index by batting order
	state   bool
}

type Match struct {
	t1        Team
	t2        Team
	score     string // Short Summary
	state     string // finished, 2nd innings, 1st innings, toss, upcoming
	tossState bool   // 0 -> First t1 bat, t2 bowl, First 1 -> t1 bowl, t2 bat }
}

func regexRemove(str string, pattern string) string {
	r := regexp.MustCompile(pattern)
	s := r.ReplaceAllString(str, "")

	return s
}

func removeWhitespace(str string) string {
	return regexRemove(regexRemove(str, "^\\s"), "\\s$")
}

func findString(pattern string, str string) string {
	r := regexp.MustCompile(pattern)
	return r.FindString(str)
}

func parseBatters(c **colly.Collector, players *map[int]Player, innings string) {
	i := 1
	(*c).OnHTML("#innings_"+innings+" div.cb-ltst-wgt-hdr:nth-of-type(1) .cb-scrd-itms",
		func(e *colly.HTMLElement) {
			switch e.ChildText(".cb-col:nth-child(1)") {
			case "Extras", "Total", "Did not Bat", "Yet to Bat":
				e.ForEach("a.cb-text-link:not(.cb-col)", func(x int, e *colly.HTMLElement) {
					var p Player
					p.name = removeWhitespace(e.Text)
					p.batorder = i
					i++
					id, _ := strconv.Atoi(findString("\\d+", e.Attr("href")))
					(*players)[id] = p
				})
				return
			}

			var p Player
			id, _ := strconv.Atoi(findString("\\d+", e.ChildAttr("a.cb-text-link", "href")))

			p.batorder = i
			i++
			p.name = removeWhitespace(e.ChildText("a.cb-text-link"))
			p.runs, _ = strconv.Atoi(e.ChildText(":nth-child(3)"))
			p.balls, _ = strconv.Atoi(e.ChildText(":nth-child(4)"))
			p.four, _ = strconv.Atoi(e.ChildText(":nth-child(5)"))
			p.six, _ = strconv.Atoi(e.ChildText(":nth-child(6)"))
			p.strikeRate = (float32(p.runs) / float32(p.balls)) * 100.0

			(*players)[id] = p
		})
}

func parseBowlers(c **colly.Collector, players *map[int]Player, innings string) {
	(*c).OnHTML("#innings_"+innings+" div.cb-ltst-wgt-hdr:nth-of-type(4) .cb-scrd-itms", func(e *colly.HTMLElement) {
		id, _ := strconv.Atoi(findString("\\d+", e.ChildAttr("a.cb-text-link", "href")))
		p := (*players)[id]

		p.name = removeWhitespace(e.ChildText("a.cb-text-link"))
		p.overs, _ = strconv.Atoi(e.ChildText(":nth-child(2)"))
		p.maidens, _ = strconv.Atoi(e.ChildText(":nth-child(3)"))
		p.oruns, _ = strconv.Atoi(e.ChildText(":nth-child(4)"))
		p.wickets, _ = strconv.Atoi(e.ChildText(":nth-child(5)"))
		p.economy = float32(p.oruns) / float32(p.overs)

		(*players)[id] = p
	})
}

func main() {
	var m Match
	m.t1.players = make(map[int]Player)
	m.t2.players = make(map[int]Player)
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Got HTML")
	})

	parseBatters(&c, &m.t1.players, "1")
	parseBowlers(&c, &m.t1.players, "2")
	parseBatters(&c, &m.t2.players, "2")
	parseBowlers(&c, &m.t2.players, "1")

	c.Visit("https://www.cricbuzz.com/live-cricket-scorecard/66292/rr-vs-lsg-26th-match-indian-premier-league-2023")
	fmt.Println(m.t1.players)
	fmt.Println(m.t2.players)
}
