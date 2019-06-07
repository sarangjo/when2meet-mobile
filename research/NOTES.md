# Notes

Important endpoints:

URL: `https://www.when2meet.com/?6939716-nrhEh`

- `AvailabilityGrids.php`
    - method: `get`
    - GET parameters:
        - `id=6939716`
        - `code=nrhEh`
        - `participantTimeZone=" + participantTimeZone` (using moment.js)
    - response:
        - a bunch of HTML that shows availability
- `ProcessLogin.php`
    - method: `get`
    - GET parameters:
        - `id=6939716`
        - `name`
        - `password` (plaintext?)
    - response:
        - person's ID
- `SaveTimes.php`
    - method: "post",
    - parameters:
        - `person`=UserID
        - `event`=6939716
        - `slots=`+TimesToToggle.join(","): a comma-separated list of the actual time slots to save
        - availability="+binaryAvailability: binary representation of availability, starting with the very first quarter-hour and spanning to the end
        - ChangeToAvailable="+ChangeToAvailable,: whether the selection is being set to "available" or "busy"
    - response
        - empty response content
    - [ ] This request is not actually doing anything.
    
```
curl 'https://www.when2meet.com/SaveTimes.php' -H 'Origin: https://www.when2meet.com' -H 'Accept-Encoding: gzip, deflate, br' -H 'Accept-Language: en-US,en;q=0.9' -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36' -H 'Content-type: application/x-www-form-urlencoded; charset=UTF-8' -H 'Accept: text/javascript, text/html, application/xml, text/xml, */*' -H 'X-Prototype-Version: 1.6.0' -H 'X-Requested-With: XMLHttpRequest' -H 'Connection: keep-alive' -H 'Referer: https://www.when2meet.com/?7885916-dZacy' --data 'person=34306825&event=7885916&slots=279795600&availability=0000000000000000000000000000000000000000000000000000000000000000&ChangeToAvailable=false&_=' --compressed
```

## Things to do

2. Set up parameters TimesToToggle, binaryAvailability, and ChangeToAvailable for `SaveTimes.php`

What exactly are the values for "slots"? That's all that's left, I feel like that ties together the Availability HTML and what to send back for SaveTimes

**It's the start Unix timestamp for the time slot!**

```
1530333900 is June 29, 9:45pm-10:00pm

1530374400 is June 30, 9:00am-9:15am
...
1530418500 is June 30, 9:15-9:30
1530419400 is June 30, 9:30pm-9:45pm
1530420300 is June 30, 9:45pm-10:00pm
```

- [ ] Need to parse out AvailabilityGrids.php to get the start/stop timestamp too. fuhhhhhhhh
- [x] Keep in mind that when2meets are not always continuous periods of time - you could have disparate days of the week and/or dates in the year

### Get people's availability

AvailableAtSlot - array of arrays

Ah shit. The timeslots in a when2meet are not necessarily contiguous: the start **time** and end **time** are the same across the board, but the actual dates can be **random** dates or days of the week.

## Step by step

- [x] Get ID and code from user &#x2192; corresponds to w2m instance
- [ ] Use `AvailabilityGrids.php` to get:
    - [ ] the exact timestamps for the current instance
        - [ ] TODO: currently our HTML parser is missing out on ~84 or so timeslots on one of the sample responses
        - [ ] Distinguish between "day of week" and "calendar date" modes
    - [ ] Get the current availability information:
        - [ ] Users (name + id)
        - [ ] Free/busy timestamps (denotes start of 15 minute interval)
- [ ] Package and send this to the client to be displayed cleanly
- [ ] User sends up list of timestamps and action (set as free vs. busy)
- [ ] Use `SaveTimes.php` to save these times
