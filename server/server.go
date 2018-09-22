package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type freeSlot struct {
	startTimestamp uint
	numIntervals   uint
}

type availability struct {
	id        int
	name      string
	freeSlots []freeSlot
}

func getAvailabilityJs() []string {
	url := "http://when2meet.com/AvailabilityGrids.php?id=6939716&lel=nrhEh"

	resp, err := http.Get(url)
	if err != nil {
		// handle error
	}

	z := html.NewTokenizer(resp.Body)
	defer resp.Body.Close()

	for {
		tagType := z.Next()
		if tagType == html.ErrorToken {
			// error
			break
		} else if tagType == html.StartTagToken {
			tagName, _ := z.TagName()
			if string(tagName) != "script" {
				// Should not happen! Error.
				fmt.Println("ERROR!!!!! Did not start with script")
				break
			}
		} else if tagType == html.TextToken {
			content := string(z.Text())
			return strings.Split(content, "\n")
		}
	}
	return nil
}

func getAvailability() []availability {
	content := getAvailabilityJs()
	if content == nil {
		// error
		fmt.Println("Error")
		return nil
	}

	// Custom parsing.
    // 1. Id and name
	stmts := strings.Split(content[0], ";")
	avail := make([]availability, len(stmts)/2)
	for i, stmt := range stmts {
		if strings.Index(stmt, "Names") >= 0 {
			avail[i/2].name = strings.Trim(strings.Split(stmt, "=")[1], " '\"")
		} else if strings.Index(stmt, "IDs") >= 0 {
			id, err := strconv.Atoi(strings.Trim(strings.Split(stmt, "=")[1], " '\""))
			if err == nil {
				avail[i/2].id = id
			} else {
				fmt.Println("ERROR", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("l u l w u t", i)
		}
	}

    // 2. Hex parsing
    i := 1
    for i < len(content) {

    }

    for _, a := range avail {
        fmt.Println("id", a.id, "name", a.name)
    }

	return avail
}

func main() {
	getAvailability()
}
