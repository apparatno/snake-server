# Snakes on a plate

Backend for a snake game where the screen is an
ESP8266 controlled strip of Neopixels
and the controller is a web page.

See also
[the client](https://github.com/apparatno/snake-client)
and the screen-controlling
[microprocessor](https://github.com/apparatno/snake-microprocessor).

## API

**`GET /state`**

Checks whether there is an active game.
Returns a JSON document with the field
`status` set to `playing`
if there is a game.
All other values can be interpreted as no game is running.

**`GET /screen`**

Returns the current board as a string of characters
where each character represents a pixel.

The values can be parsed as:

* `0` pixel off
* `1` part of snake
* `2` a fruit

If the request returns status `HTTP/404`
it means the game is over.

**POST /play`**

Starts a new game.
The response returns a JSON document
which contains a `playerToken`.
This token must be included in all subsequent
requests when controlling the game.

If a game is already running
this request will return `HTTP/400`.

**`POST /action`**

Sends a command to move the snake.
The body of the request must be form-data
with these fields:

* `playerToken`: The token returned when a new game
  was started
* `keyPressed`: A character representing the button
  that was pressed. See below for possible values.

**Key presses**

* `U`: move up
* `D`: Move down
* `L`: Move left
* `R`: Move right
* `X`: (Re)start game

All other values will return an error.  
