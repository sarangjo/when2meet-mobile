package main

import (
	"net/http"
	"fmt"
	// "math"
	"strconv"
	"strings"
	"golang.org/x/net/html"
	"github.com/emirpasic/gods/sets/treeset"
	"github.com/emirpasic/gods/utils"
)

// TEST debug
const TEST = true

// Represents a single When2Meet "instance". Identified by an id and a code.
type instance struct {
	id   uint
	code string
	timeslots	[]uint64 // In-order 15-minute intervals
}

// Represents the response to the GetAvailability request
type availabilityResponse struct {
	avail          []availability
}

// Per-user availability
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

// Returns availability, start timestamp, and number of 15-minute slots
func getAvailability(inst *instance) (ar availabilityResponse, err error) {
	fmt.Println("Getting availability grids...")

	url := fmt.Sprintf("https://when2meet.com/AvailabilityGrids.php?id=%d&code=%s", inst.id, inst.code)

	resp, err := http.Get(url)
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// Variables to collect data
	var availabilityJs []string
	state := start
	availabilityTimestamps := treeset.NewWith(utils.UInt64Comparator)

	doc, err := html.Parse(resp.Body)
	if err != nil {
		// error
		return ar, err
	}
	// Returns whether to continue parsing the tree
	var f func(*html.Node) (bool, error)
	// 1. The top-level <script> tag in the response contains information about availability.
	// (parserFoundScript)
	// 2. The div with id "YouGridSlots" contains a div for each 15-minute timeslot, which in turn
	// have id's starting with "YouTime" followed by the Unix timestamp. From this we can compute
	// all of the timestamps that are in this when2meet.
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
				validTimeslot := false
				var timestamp uint64
				for _, attr := range n.Attr {
					if attr.Key == "id" && strings.Index(attr.Val, "YouTime") >= 0 {
						validTimeslot = true
					} else if attr.Key == "data-time" {
						timestamp, _ = strconv.ParseUint(attr.Val, 10, 64)
					}
				}

				if validTimeslot {
					availabilityTimestamps.Add(timestamp)
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

	// Transfer set into array
	for _, ts := range availabilityTimestamps.Values() {
		if tsUint, ok := ts.(uint64); ok {
			inst.timeslots = append(inst.timeslots, tsUint)
		}
	}

	// Custom parsing through the availabilityJs content in the <script> we parsed out
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
			// ar.nOfTimeslots = 4 * uint64(len(strings.Replace(line, "// hexAvailability: 0", "", 1)))
			break
		}
	}

	return ar, nil
}

// TODO not sure what this is even doing
/*
func addTimestamps(ins instance, ar availabilityResponse) {
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
*/

var kinspireWhen2Meet = instance{6939716, "nrhEh", make([]uint64, 0)}

func main() {
	/*
	reader := bufio.NewReader(os.Stdin)

	// ID
	fmt.Print("id: ")
	idString, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Invalid ID");
		os.Exit(1)
	}
	if TEST {
		idString = "6939716"
	}
	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		fmt.Println("ID needs to be a valid number")
		os.Exit(1)
	}

	// Code
	fmt.Print("code: ")
	code, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Invalid code");
		os.Exit(1)
	}
	if TEST {
		code = "nrhEh"
	}
	*/

	// Set up instance for this run
	inst := kinspireWhen2Meet // instance{uint(id), code, make([]uint64, 512)}

	ar, err := getAvailability(&inst)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Full timestamps:")
	fmt.Println(inst.timeslots)

	// addTimestamps(date, ar)
	fmt.Println(ar)
}
