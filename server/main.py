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
    PARSER_FOUND_SCRIPT = 2
    PARSER_LOOKING_FOR_SLOTS_START = 3
    PARSER_FOUND_SLOTS_START = 4


def get_availability(inst: Instance) -> AvailabilityResponse:
    """
    Returns availability, start timestamp, and number of 15-minute slots
    """
    print("Getting availability grids...")

    url = f"https://when2meet.com/AvailabilityGrids.php?id={inst.id}&code={inst.code}"
    r = requests.get(url)

    class MyHtmlParser(HTMLParser):
        """
        1. The top-level <script> tag in the response contains information about availability.
        (parserFoundScript)
        2. The div with id "YouGridSlots" contains a div for each 15-minute timeslot, which in turn
        have id's starting with "YouTime" followed by the Unix timestamp. From this we can compute
        all of the timestamps that are in this when2meet.
        """

        def __init__(self):
            super().__init__()
            self.state = ParserState.START
            self.availabilityJs: List[str] = []
            self.availabilityTimestamps: Dict[int, bool] = collections.OrderedDict()

        def error(self, message):
            print(f"Error: {message}")

        def handle_starttag(self, tag, attrs: List[Tuple[str, str]]):
            if self.state == ParserState.START and tag == "script":
                self.state = ParserState.PARSER_FOUND_SCRIPT
            elif self.state == ParserState.PARSER_LOOKING_FOR_SLOTS_START and tag == "div" and (
                    "id", "YouGridSlots") in attrs:
                self.state = ParserState.PARSER_FOUND_SLOTS_START
            elif self.state == ParserState.PARSER_FOUND_SLOTS_START and tag == "div":
                valid_timeslot = False
                timestamp = 0
                for attr in attrs:
                    if attr[0] == "id" and attr[1].find("YouTime") >= 0:
                        valid_timeslot = True
                    elif attr[0] == "data-time":
                        timestamp = int(attr[1])

                if valid_timeslot:
                    self.availabilityTimestamps[timestamp] = True

        def handle_data(self, data: str):
            if self.state == ParserState.PARSER_FOUND_SCRIPT:
                self.availabilityJs = data.split("\n")  # strings.Split(content, "\n")
                self.state = ParserState.PARSER_LOOKING_FOR_SLOTS_START

    parser = MyHtmlParser()
    parser.feed(r.text)

    # Transfer set into instance array
    inst.timeslots = parser.availabilityTimestamps.keys()

    ar = AvailabilityResponse()

    # Custom parsing through the availabilityJs content in the <script> we parsed out
    # 1. Id and name
    stmts = parser.availabilityJs[0].split(";")
    for i, stmt in enumerate(stmts):
        # strip out space, single quotes, and double quotes
        if stmt.find("Names") >= 0:
            ar.avail.append(Availability(0, stmt.split("=")[1].strip(" '\"")))
        elif stmt.find("IDs") >= 0:
            ar.avail[math.floor(i / 2)].id = int(stmt.split("=")[1].strip(" '\""))

    # 2. Hex parsing
    hex_index = 1
    while hex_index < len(parser.availabilityJs):
        line = parser.availabilityJs[hex_index]
        if line.find("hexAvailability") >= 0:
            # TODO
            n_of_timeslots = 4 * len(line.replace("# hexAvailability: 0", "", 1))
            print(f"hex gives us {n_of_timeslots} timeslots, html gives us {len(inst.timeslots)}")
            break

        hex_index += 1

    return ar


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

kinspireWhen2Meet = Instance(6939716, "nrhEh")


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
    inst = kinspireWhen2Meet  # instance{uint(id), code, make([]uint64, 512)}

    ar = get_availability(inst)

    print("Full timestamps:")
    print(inst.timeslots)

    # addTimestamps(date, ar)
    print(ar)


if __name__ == '__main__':
    main()
