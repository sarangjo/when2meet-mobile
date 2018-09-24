package main

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/msoap/byline"
	"golang.org/x/net/html"
)

// AvailabilityResponse represents the response to the GetAvailability request
type AvailabilityResponse struct {
	avail          []availability
	startTimestamp uint64
	nOfTimeslots   uint64 // # of 15 minute intervals
}

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
func getAvailability(i instance) (ar AvailabilityResponse, err error) {
	url := fmt.Sprintf("https://when2meet.com/AvailabilityGrids.php?id=%d&code=%s", i.id, i.code)

	resp, err := http.Get(url)
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	var availabilityJs []string
	state := start

	// Tokenize
	doc, err := html.Parse(resp.Body)

	if err != nil {
		// error
		return ar, err
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
					ar.startTimestamp = temp
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
		return ar, err
	}

	// Custom parsing.
	// 1. Id and name
	stmts := strings.Split(availabilityJs[0], ";")
	ar.avail = make([]availability, len(stmts)/2)
	for i, stmt := range stmts {
		if strings.Index(stmt, "Names") >= 0 {
			ar.avail[i/2].name = strings.Trim(strings.Split(stmt, "=")[1], " '\"")
		} else if strings.Index(stmt, "IDs") >= 0 {
			id, err := strconv.Atoi(strings.Trim(strings.Split(stmt, "=")[1], " '\""))
			if err == nil {
				ar.avail[i/2].id = id
			} else {
				return ar, err
			}
		}
	}

	// 2. Hex parsing
	for hexIndex := 1; hexIndex < len(availabilityJs); hexIndex++ {
		line := availabilityJs[hexIndex]
		if strings.Index(line, "hexAvailability") >= 0 {
			// TODO
			ar.nOfTimeslots = 4 * uint64(len(strings.Replace(line, "// hexAvailability: 0", "", 1)))
			break
		}
	}

	return ar, nil
}

func addTimestamps(ins instance, ar AvailabilityResponse) {
	url := fmt.Sprintf("https://when2meet.com/?%d-%s", ins.id, ins.code)
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// TODO Setup
	// one byte per 2 timeslots
	for _, a := range ar.avail {
		a.freeSlots = make([]byte, int(math.Ceil(float64(ar.nOfTimeslots/8))))
	}

	// Process line by line and if a line contains the required string, get it
	lr := byline.NewReader(resp.Body)
	lr.GrepByRegexp(regexp.MustCompile("AvailableAtSlot\\[[0-9]+\\]\\.push\\([0-9]+\\);"))
	lr.EachString(func(line string) {
		var slotIndex int
		var userID uint64

		// Extract the two numbers we're looking for
		sec := strings.Split(line, "AvailableAtSlot[")
		if len(sec) > 2 {
			// TODO oops? more than one per line?
		}
		part := sec[1]
		sec = strings.Split(part, "].push(")
		slotIndex, err = strconv.Atoi(sec[0])
		if err != nil {
			// process error
			fmt.Println(err)
			return
		}

		part = sec[1]
		sec = strings.Split(part, ");")
		userID, err = strconv.ParseUint(sec[0], 10, 64)
		if err != nil {
			// process error
			fmt.Println(err)
			return
		}

		// TODO Calculate offset from start, set the hex
		fmt.Println(slotIndex, userID)
	})

	result, err := lr.ReadAllSliceString()
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}
	fmt.Println("Lines with availability found: ", len(result))

	// TODO
}

func main() {
	// day := instance{6785667, "7INRG"}
	date := instance{6939716, "nrhEh"}

	ar, err := getAvailability(date)
	if err != nil {
		fmt.Println(err)
	}
	addTimestamps(date, ar)
	fmt.Println(ar)
}
