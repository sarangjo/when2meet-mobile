# Notes

Important endpoints:

URL: `https://www.when2meet.com/?6939716-nrhEh`

- `AvailabilityGrids.php`
    - method: `post`
    - parameters:
        - `id=6939716`
        - `code=nrhEh`
        - `participantTimeZone=" + participantTimeZone` (using moment.js)
    - response:
        - a bunch of HTML that shows availability
- `ProcessLogin.php`
    - method: `post`
    - parameters:
        - `id=6939716"`
        - `name`
        - `password`
    - response:
        - person's ID
- `SaveTimes.php`
    - method: "post",
    - parameters:
        - person="+UserID
        - event=6939716"
        - slots="+TimesToToggle.join(",")
        - availability="+binaryAvailability
        - ChangeToAvailable="+ChangeToAvailable,

## Things to do

1. Parse response HTML for `AvailabilityGrids.php`
2. Set up parameters TimesToToggle, binaryAvailability, and ChangeToAvailable for `SaveTimes.php`
