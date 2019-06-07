import math
from html.parser import HTMLParser
import collections
from enum import Enum
from typing import List, Tuple, Dict
import requests

TEST = True


class Instance:
    """
    Represents a single When2Meet "instance". Identified by an id and a code.
    """

    def __init__(self, _id, code):
        self.id = _id
        self.code = code
        self.timeslots = []  # In-order 15-minute intervals
        self.mode = Mode.WEEKDAY


class AvailabilityResponse:
    """
    Represents the response to the GetAvailability request
    """

    def __init__(self):
        self.avail: List[Availability] = []


class Availability:
    """
    # Per-user availability
    """

    def __init__(self, _id, name):
        self.id = _id
        self.name = name
        self.freeSlots = []


class ParserState(Enum):
    START = 1
    FOUND_SCRIPT = 2
    LOOKING_FOR_GRID = 3
    FOUND_GRID = 4
    LOOKING_FOR_GRID_DATE = 5
    LOOKING_FOR_SLOTS_START = 10
    FOUND_SLOTS_START = 11


class Mode(Enum):
    WEEKDAY = 0
    CALENDAR_DATE = 1


class AvailabilityParser(HTMLParser):
    """
    1. The top-level <script> tag in the response contains information about availability.
    (parserFoundScript)
    2. YouGrid > last child/4th child has the actual grid > all divs except the last are the headers for each column
    3. The last div from above is the div with id "YouGridSlots" contains a div for each 15-minute timeslot, which
    in turn have id's starting with "YouTime" followed by the Unix timestamp. From this we can compute all of the
    timestamps that are in this when2meet.
    """

    def __init__(self):
        super().__init__()
        self._state = ParserState.START
        self._grid_depth = 0
        self._grid_child_idx = 0

        self.availabilityJs: List[str] = []
        self.availabilityTimestamps: Dict[int, bool] = collections.OrderedDict()
        self.mode = Mode.WEEKDAY

    def error(self, message):
        print(f"Error: {message}")

    def handle_starttag(self, tag: str, attrs: List[Tuple[str, str]]):
        if self._state == ParserState.START and tag == "script":
            self._state = ParserState.FOUND_SCRIPT
        elif self._state == ParserState.LOOKING_FOR_GRID and tag == "div" and ("id", "YouGrid") in attrs:
            self._state = ParserState.FOUND_GRID
        elif self._state == ParserState.FOUND_GRID:
            # Look for the 4th child, that's where we determine whether it's weekday or calendar date mode
            if self._grid_child_idx == 3 and self._grid_depth == 0:
                self._state = ParserState.LOOKING_FOR_GRID_DATE
            elif self._grid_depth == 0:
                self._grid_child_idx += 1

            # Annoying hack, but br is the only tag that's self-closing without actually being self-closing
            if tag != "br":
                self._grid_depth += 1
        elif self._state == ParserState.LOOKING_FOR_GRID_DATE:
            if tag == "br":
                # At this point we didn't find a data element, which means we are in weekday mode
                self.mode = Mode.WEEKDAY
                self._state = ParserState.LOOKING_FOR_SLOTS_START
        elif self._state == ParserState.LOOKING_FOR_SLOTS_START and tag == "div" and (
                "id", "YouGridSlots") in attrs:
            self._state = ParserState.FOUND_SLOTS_START
        elif self._state == ParserState.FOUND_SLOTS_START and tag == "div":
            valid_timeslot = False
            timestamp = 0
            for attr in attrs:
                if attr[0] == "id" and "YouTime" in attr[1]:
                    valid_timeslot = True
                elif attr[0] == "data-time":
                    timestamp = int(attr[1])

            if valid_timeslot:
                self.availabilityTimestamps[timestamp] = True

    def handle_data(self, data: str):
        if self._state == ParserState.FOUND_SCRIPT:
            self.availabilityJs = data.split("\n")  # strings.Split(content, "\n")
            self._state = ParserState.LOOKING_FOR_GRID
        if self._state == ParserState.LOOKING_FOR_GRID_DATE:
            if len(data.strip()) > 0:
                print("Found date! Huzzah.", data)
                self.mode = Mode.CALENDAR_DATE
                self._state = ParserState.LOOKING_FOR_SLOTS_START

        # print(data)

    def handle_endtag(self, tag):
        if self._state == ParserState.FOUND_GRID:
            self._grid_depth -= 1


def get_availability(inst: Instance) -> AvailabilityResponse:
    """
    Returns availability, start timestamp, and number of 15-minute slots
    """
    print("Getting availability grids...")

    url = f"https://when2meet.com/AvailabilityGrids.php?id={inst.id}&code={inst.code}"
    r = requests.get(url)

    parser = AvailabilityParser()
    parser.feed(r.text)

    # Get info from the parser
    # 1. availability timestamps
    inst.timeslots = parser.availabilityTimestamps.keys()

    ar = AvailabilityResponse()

    # 2. availabilityJs content (Id and name)
    stmts = parser.availabilityJs[0].split(";")
    for i, stmt in enumerate(stmts):
        # strip out space, single quotes, and double quotes
        if "PeopleNames" in stmt:
            ar.avail.append(Availability(0, stmt.split("=")[1].strip(" '\"")))
        elif "PeopleIDs" in stmt:
            ar.avail[math.floor(i / 2)].id = int(stmt.split("=")[1].strip(" '\""))

    # 3. mode
    inst.mode = parser.mode

    return ar


def login(inst: Instance, name: str, password="") -> int:
    print("Logging in...")

    url = f"https://when2meet.com/ProcessLogin.php?id={inst.id}&name={name}&password={password}"
    r = requests.get(url)

    if "Wrong" in r.text:
        raise RuntimeError(r.text)

    return int(r.text)


# TODO not sure what this is even doing
"""
func addTimestamps(ins instance, ar availabilityResponse) {
url := fmt.Sprintf("https:#when2meet.com/?%d-%s", ins.id, ins.code)
resp, err := http.Get(url)
if err != nil {
# handle error
print(err)
return
}
defer resp.Body.Close()

# TODO Setup
# one byte per 2 timeslots
for _, a := range ar.avail {
a.freeSlots = make([]byte, int(math.Ceil(float64(ar.nOfTimeslots/8))))
}

# Process line by line and if a line contains the required string, get it
lr := byline.NewReader(resp.Body)
lr.GrepByRegexp(regexp.MustCompile("AvailableAtSlot\\[[0-9]+\\]\\.push\\([0-9]+\\);"))
lr.EachString(func(line string) {
var slotIndex int
var userID uint64

# Extract the two numbers we're looking for
sec := strings.Split(line, "AvailableAtSlot[")
if len(sec) > 2 {
# TODO oops? more than one per line?
}
part := sec[1]
sec = strings.Split(part, "].push(")
slotIndex, err = strconv.Atoi(sec[0])
if err != nil {
# process error
print(err)
return
}

part = sec[1]
sec = strings.Split(part, ");")
userID, err = strconv.ParseUint(sec[0], 10, 64)
if err != nil {
# process error
print(err)
return
}

# TODO Calculate offset from start, set the hex
print(slotIndex, userID)
})

result, err := lr.ReadAllSliceString()
if err != nil {
# handle error
print(err)
return
}
print("Lines with availability found: ", len(result))

# TODO
}
"""

testCalendarDates = Instance(6939716, "nrhEh")
testWeekdays = Instance(7885916, "dZacy")


def main():
    """
    reader := bufio.NewReader(os.Stdin)

    # ID
    fmt.Print("id: ")
    idString, err := reader.ReadString('\n')
    if err != nil {
    print("Invalid ID");
    os.Exit(1)
    }
    if TEST {
    idString = "6939716"
    }
    id, err := strconv.ParseUint(idString, 10, 64)
    if err != nil {
    print("ID needs to be a valid number")
    os.Exit(1)
    }

    # Code
    fmt.Print("code: ")
    code, err := reader.ReadString('\n')
    if err != nil {
    print("Invalid code");
    os.Exit(1)
    }
    if TEST {
    code = "nrhEh"
    }
    """

    # Set up instance for this run
    inst = testWeekdays

    if False:
        ar = get_availability(inst)

        print("Full timestamps:")
        print(inst.timeslots)
        print("Instance mode:", inst.mode)

    user_id = login(inst, name="Sarang")

    # addTimestamps(date, ar)


if __name__ == '__main__':
    main()
