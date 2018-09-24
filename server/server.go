package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/msoap/byline"
	"golang.org/x/net/html"
)

type availability struct {
	id        int
	name      string
	freeSlots []byte
}

const (
	start               = iota
	tokenizerReadScript = iota
)

const (
	_                          = iota
	parserFoundScript          = iota
	parserLookingForSlotsStart = iota
	parserFoundSlotsStart      = iota
)

type instance struct {
	id   uint
	code string
}

// Returns availability, start timestamp, and number of 15-minute slots
func getAvailability(i instance) (avail []availability, startTimestamp uint64, nOfTimeslots uint64) {
	url := fmt.Sprintf("https://when2meet.com/AvailabilityGrids.php?id=%d&code=%s", i.id, i.code)

	resp, err := http.Get(url)
	if err != nil {
		// handle error
		fmt.Println(err)
		return nil, 0, 0
	}
	defer resp.Body.Close()

	var availabilityJs []string
	state := start

	// Tokenize
	shouldTokenize := false
	if shouldTokenize {
		z := html.NewTokenizer(resp.Body)

		for {
			tagType := z.Next()
			if tagType == html.ErrorToken {
				// error
				fmt.Println("tokenizing error")
				return nil, 0, 0
			} else if tagType == html.StartTagToken {
				tagName, _ := z.TagName()
				if string(tagName) != "script" && availabilityJs == nil {
					state = tokenizerReadScript
				}
			} else if tagType == html.TextToken {
				if state == tokenizerReadScript {
					content := string(z.Text())
					availabilityJs = strings.Split(content, "\n")
				}
			}
		}
	} else {
		doc, err := html.Parse(resp.Body)

		if err != nil {
			// error
			fmt.Println(err)
			return nil, 0, 0
		}
		// Returns whether to continue parsing the tree
		var f func(*html.Node) (bool, error)
		f = func(n *html.Node) (bool, error) {
			switch state {
			case start:
				if n.Type == html.ElementNode && n.Data == "script" {
					state = parserFoundScript
				}
				break
			case parserFoundScript:
				if n.Type == html.TextNode {
					content := string(n.Data)
					availabilityJs = strings.Split(content, "\n")
					state = parserLookingForSlotsStart
				}
				break
			case parserLookingForSlotsStart:
				if n.Type == html.ElementNode && n.Data == "div" {
					for _, attr := range n.Attr {
						if attr.Key == "id" && attr.Val == "YouGridSlots" {
							state = parserFoundSlotsStart
						}
					}
				}
				break
			case parserFoundSlotsStart:
				if n.Type == html.ElementNode && n.Data == "div" {
					found := false
					var temp uint64
					for _, attr := range n.Attr {
						if attr.Key == "id" && strings.Index(attr.Val, "YouTime") >= 0 {
							found = true
						} else if attr.Key == "data-time" {
							temp, err = strconv.ParseUint(attr.Val, 10, 64)
							if err != nil {
								// error
								return false, err
							}
						}
					}

					if found {
						startTimestamp = temp
						return false, nil
					}
				}
				break
			}

			// Recurse!
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				var cont bool
				cont, err = f(c)
				if !cont || err != nil {
					return cont, err
				}
			}

			return true, nil
		}
		_, err = f(doc)
		if err != nil {
			fmt.Println(err)
			return nil, 0, 0
		}
	}

	// Custom parsing.
	// 1. Id and name
	stmts := strings.Split(availabilityJs[0], ";")
	avail = make([]availability, len(stmts)/2)
	for i, stmt := range stmts {
		if strings.Index(stmt, "Names") >= 0 {
			avail[i/2].name = strings.Trim(strings.Split(stmt, "=")[1], " '\"")
		} else if strings.Index(stmt, "IDs") >= 0 {
			id, err := strconv.Atoi(strings.Trim(strings.Split(stmt, "=")[1], " '\""))
			if err == nil {
				avail[i/2].id = id
			} else {
				fmt.Println(err)
				return nil, 0, 0
			}
		}
	}

	// 2. Hex parsing
	for hexIndex := 1; hexIndex < len(availabilityJs); hexIndex++ {
		line := availabilityJs[hexIndex]
		if strings.Index(line, "hexAvailability") >= 0 {
			// TODO
			nOfTimeslots = uint64(len(strings.Replace(line, "// hexAvailability: 0", "", 1)))
			break
		}
	}

	return avail, startTimestamp, nOfTimeslots
}

func addTimestamps(i instance, avail []availability) {
	url := fmt.Sprintf("https://when2meet.com/?%d-%s", i.id, i.code)
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// Process line by line and if a line contains the required string, get it
	lr := byline.NewReader(resp.Body)
	lr.GrepByRegexp(regexp.MustCompile("AvailableAtSlot\\[[0-9]+\\]\\.push\\([0-9]+\\);"))
	result, err := lr.ReadAllSliceString()
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}
	fmt.Println("Lines with availability found: ", len(result))
}

func main() {
	// day := instance{6785667, "7INRG"}
	date := instance{6939716, "nrhEh"}

	avail, start, n := getAvailability(date)
	addTimestamps(date, avail)
	fmt.Println(avail, start, n)
}
