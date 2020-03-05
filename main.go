package main

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

const pixels = 300
const width = 20
const defaultTTL = 100 // inactivity for this many game loops kills the session

type fruit struct {
	position              int
	placementWaitCycle    int
	maxPlacementWaitCycle int
	consumed              bool
	variant               byte
	points                int
}

type session struct {
	snek             []int
	currentDirection string
	token            string
	randomizer       *rand.Rand
	ttl              int
	fruits           []fruit
	totalPoints      int
}

type state struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

type gameState struct {
	Board  string `json:"board"`
	Points int    `json:"points"`
}

type server struct {
	session *session
}

type gameData struct {
	PlayerToken string `json:"playerToken"`
}

func setCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
func main() {
	log.Println("SNAKES ON A MOTHERFUCKING PLATE GETTING IT ON")

	s := server{}
	mux := http.NewServeMux()

	mux.HandleFunc("/gamestate", func(w http.ResponseWriter, r *http.Request) {
		setCors(&w)
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("method not supported"))
			return
		}
		board, err := s.getBoard()
		if err != nil {
			// error means no board means game over
			log.Printf("err %v", err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
		}

		points := s.session.totalPoints
		gameState := gameState{
			Board:  string(board),
			Points: points,
		}

		if err := json.NewEncoder(w).Encode(&gameState); err != nil {
			log.Println(err)
		}
	})

	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		setCors(&w)
		currentState := state{}
		if s.session != nil {
			currentState.Status = "playing"
		}
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(&currentState); err != nil {
			log.Println(err)
		}
	})
	mux.HandleFunc("/play", func(w http.ResponseWriter, r *http.Request) {
		setCors(&w)
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("method not supported"))
			return
		}
		if s.session != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("game already running"))
			return
		}

		s.session = newSession()
		w.Header().Add("Content-Type", "application/json")
		res := gameData{PlayerToken: s.session.token}
		if err := json.NewEncoder(w).Encode(&res); err != nil {
			log.Println(err)
		}
		log.Println("new game started")
	})
	mux.HandleFunc("/screen", func(w http.ResponseWriter, r *http.Request) {
		setCors(&w)
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("method not supported"))
			return
		}
		b, err := s.getBoard()
		if err != nil {
			// error means no board means game over
			log.Printf("err %v", err)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
		}

		_, err = w.Write(b)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/action", func(w http.ResponseWriter, r *http.Request) {
		setCors(&w)
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("method not supported"))
			return
		}

		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse form data: " + err.Error()))
			return
		}
		playerToken := r.Form.Get("playerToken")
		keyPressed := r.Form.Get("keyPressed")
		if playerToken != s.session.token {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("token '" + playerToken + "' does not match current player token"))

			return
		}
		key := strings.ToUpper(keyPressed)
		b, err := s.input(key)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error())) // TODO return a fancy screen here
			return
		}
		_, err = w.Write(b)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	go gameLoop(&s)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Println(err)
	}
}

func newSession() *session {
	randomizer := makeRandomizer()
	snek := []int{112, 111, 110}

	token := uuid.NewV4().String()
	fruits := []fruit{
		fruit{
			position:              placeFruit(snek, randomizer),
			maxPlacementWaitCycle: 10,
			points:                10,
			variant:               '2',
		},
		fruit{
			position:              placeFruit(snek, randomizer),
			maxPlacementWaitCycle: 15,
			points:                20,
			variant:               '3',
		},
	}

	sess := session{
		snek:             snek,
		currentDirection: "R",
		token:            token,
		randomizer:       randomizer,
		ttl:              defaultTTL,
		fruits:           fruits,
	}
	log.Printf("created new session %#v", sess)
	return &sess
}

func (s *server) getBoard() ([]byte, error) {
	if s.session == nil {
		return nil, errors.New("game over")
	}

	b := boardAsBytes(s.session.snek, s.session.fruits)
	return b, nil
}

func (s *server) input(cmd string) ([]byte, error) {
	if err := s.updateBoard(cmd); err != nil {
		log.Println(err)
		return nil, err
	}

	b := boardAsBytes(s.session.snek, s.session.fruits)
	return b, nil
}

func (s *server) updateBoard(cmd string) error {
	if cmd == "X" {
		s.session = newSession()
		return nil
	}

	log.Printf("updating board with command %s", cmd)
	switch cmd {
	case "U":
		if s.session.currentDirection == "U" || s.session.currentDirection == "D" {
			return nil // cannot move up or down now
		}
	case "D":
		if s.session.currentDirection == "U" || s.session.currentDirection == "D" {
			return nil // cannot move up or down now
		}
	case "L":
		if s.session.currentDirection == "L" || s.session.currentDirection == "R" {
			return nil // cannot move left or right now
		}
	case "R":
		if s.session.currentDirection == "L" || s.session.currentDirection == "R" {
			return nil // cannot move left or right now
		}
	default:
		return errors.New("unknown command " + cmd)
	}
	s.session.currentDirection = cmd
	s.session.ttl = defaultTTL
	return nil
}

func boardAsBytes(snek []int, fruits []fruit) []byte {
	b := make([]byte, pixels)
	for i := range b {
		b[i] = byte('0')
	}
	for _, s := range snek {
		b[s] = byte('1')
	}

	for _, fruit := range fruits {
		if !fruit.consumed {
			b[fruit.position] = fruit.variant
		}
	}

	return b
}

func makeRandomizer() *rand.Rand {
	s := rand.NewSource(time.Now().Unix())
	return rand.New(s)
}

func placeFruit(snek []int, r *rand.Rand) int {
	log.Println("placing a fruit")
	var i int
	for {
		i = r.Intn(pixels)
		var inUse bool
		for _, n := range snek {
			if n == i {
				inUse = true
				break
			}
		}
		if !inUse {
			log.Printf("placed fruit at %d", i)
			return i
		}
	}
}

func consumedFruit(snekHeadPosition int, fruits []fruit) (bool, int) {
	for i := range fruits {
		if snekHeadPosition == fruits[i].position {
			return true, i
		}
	}
	return false, -1
}

func gameLoop(s *server) {
	t := time.NewTicker(time.Millisecond * 500)
	for {
		select {
		case <-t.C:
			if s.session == nil {
				continue
			}

			snake, err := moveMotherfuckingSnake(s.session.snek, s.session.currentDirection)
			if err != nil {
				// error means snake collided with something - game over
				s.session = nil
				continue
			}

			consumed, fruitIdx := consumedFruit(snake[0], s.session.fruits)
			if consumed {
				log.Printf("snake comsumed fruit at %d", snake[0])
				fruit := s.session.fruits[fruitIdx]

				fruit.consumed = true
				fruit.placementWaitCycle = s.session.randomizer.Intn(fruit.maxPlacementWaitCycle) + 1
				fruit.position = placeFruit(snake, s.session.randomizer)

				s.session.fruits[fruitIdx] = fruit
				s.session.totalPoints += fruit.points
				snake = append(snake, s.session.snek[len(s.session.snek)-1])
			}

			for i, fruit := range s.session.fruits {
				if fruit.consumed {
					s.session.fruits[i].placementWaitCycle--
					if fruit.placementWaitCycle == 0 {
						s.session.fruits[i].consumed = false
					}
				}
			}

			s.session.snek = snake

			s.session.ttl--
			if s.session.ttl == 0 {
				log.Println("zombie game - killing it")
				s.session = nil // killing the session for inactivity
			}
		}
	}
}

func moveMotherfuckingSnake(snake []int, direction string) ([]int, error) {
	nextPixel := calculateNextPixel(snake, direction)

	snek := make([]int, 1, len(snake))
	snek[0] = nextPixel
	snek = append(snek, snake[:len(snake)-1]...)

	if collides(snek) {
		log.Printf("snake collided %#v", snek)
		return nil, errors.New("game over")
	}

	return snek, nil
}

func collides(snake []int) bool {
	for _, s := range snake[1:] {
		if s == snake[0] {
			log.Printf("%d collided with %d (%#v)", s, snake[0], snake)
			return true
		}
	}
	return false
}

func calculateNextPixel(snake []int, direction string) int {
	currentPixel := snake[0]
	var nextPixel int

	switch direction {
	case "U":
		nextPixel = currentPixel - width
		if nextPixel < 0 {
			nextPixel += pixels
		}
	case "D":
		nextPixel = currentPixel + width
		if nextPixel > pixels {
			nextPixel -= pixels
		}
	case "L":
		nextPixel = currentPixel - 1
		if currentPixel%width == 0 {
			nextPixel = nextPixel + width
		}
	case "R":
		nextPixel = currentPixel + 1
		if nextPixel%width == 0 {
			nextPixel = nextPixel - width
		}
	}
	return nextPixel
}

func getDefaultBoard(ballPos int) []byte {
	b := byte('0')
	bo := make([]byte, pixels)
	for i := range bo {
		bo[i] = b
	}
	b = byte('1')

	bo[ballPos] = b
	for j := ballPos - 1; j <= ballPos+1; j++ {
		for k := -1; k <= 1; k++ {
			x := k*width + j
			log.Printf("%d x width + %d = %d", k, j, x)
			bo[x] = b
		}
	}

	return bo
}
