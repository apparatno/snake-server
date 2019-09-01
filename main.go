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
const defaultTTL = 20 // inactivity for this many game loops kills the session

type session struct {
	snek             []int
	currentDirection string
	fruit            int
	token            string
	randomizer       *rand.Rand
	ttl              int
}

type State struct {
	Status string `json:"status"`
	Token  string `json:"token"`
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
	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		setCors(&w)
		currentState := State{}
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

	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Println(err)
	}
}

func newSession() *session {
	randomizer := makeRandomizer()
	snek := []int{112, 111, 110}
	f := placeFruit(snek, randomizer)

	token := uuid.NewV4().String()

	sess := session{
		snek:             snek,
		currentDirection: "R",
		fruit:            f,
		token:            token,
		randomizer:       randomizer,
		ttl:              defaultTTL,
	}
	log.Printf("created new session %#v", sess)
	return &sess
}

func (s *server) getBoard() ([]byte, error) {
	if s.session == nil {
		return nil, errors.New("game over")
	}

	b := boardAsBytes(s.session.snek, s.session.fruit)
	return b, nil
}

func (s *server) input(cmd string) ([]byte, error) {
	if err := s.updateBoard(cmd); err != nil {
		log.Println(err)
		return nil, err
	}

	b := boardAsBytes(s.session.snek, s.session.fruit)

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

func boardAsBytes(snek []int, fruit int) []byte {
	b := make([]byte, pixels)
	for i := range b {
		b[i] = byte('0')
	}
	for _, s := range snek {
		b[s] = byte('1')
	}
	if fruit >= 0 {
		b[fruit] = byte('2')
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

func gameLoop(s *server) {
	t := time.NewTicker(time.Millisecond * 500)
	var waitCyclesToPlaceFruit int
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

			// Keep the last snake pixel if it ate a fruit
			if s.session.fruit == snake[0] {
				log.Printf("snake ate fruit at %d", snake[0])
				snake = append(snake, s.session.snek[len(s.session.snek)-1])
				s.session.fruit = -1 // delete the fruit
				waitCyclesToPlaceFruit = s.session.randomizer.Intn(10) + 1
				log.Printf("wait for %d cycles to place new fruit", waitCyclesToPlaceFruit)
			}

			s.session.snek = snake

			// count down to place new fruit
			if s.session.fruit == -1 {
				log.Printf("there is no fruit")
				waitCyclesToPlaceFruit--
				if waitCyclesToPlaceFruit == 0 {
					s.session.fruit = placeFruit(snake, s.session.randomizer)
					log.Printf("dropped fruit at %d", s.session.fruit)
				}
			}

			s.session.ttl--
			if s.session.ttl == 0 {
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
		return nil, errors.New("game over")
	}

	return snek, nil
}

func collides(snake []int) bool {
	for _, s := range snake[1:] {
		if s == snake[0] {
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
