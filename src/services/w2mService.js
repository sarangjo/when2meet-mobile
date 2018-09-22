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

    let html;

    let request = new Request("http://when2meet.com/AvailabilityGrids.php?id=6939716&lel=nrhEh");
    let myInit = {
      method: 'POST',
      mode: 'cors'
    };

    return fetch(request, myInit).then(function(response) {
      console.log(response);
    });

    // Return availability of the form:
    // [
    //     "id": USER_ID,
    //     "name": USER_NAME,
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
