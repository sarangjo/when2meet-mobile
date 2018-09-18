// @flow
export default {
    getAvailability,
    login,
    saveTimes,
};

// TODO add timezone
function getAvailability(id, code) {
    // POST(https://when2meet.com/AvailabilityGrids.php)
    // Parse response HTML

    // Return availability of the form:
    // [
    //     "id": USER_ID,
    //     "availability":
    //     [
    //         {
    //             "start-timestamp": START_TIMESTAMP,
    //             "number-of-15min-intervals": NUM_INVERVALS
    //         },
    //         ...
    //     ]
    // ]
}

function login(id, name, password) {
    // TODO
}

function saveTimes(userId, id, userAvailability) {
    // TODO
}
