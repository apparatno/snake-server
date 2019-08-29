package main

import (
	"errors"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const pixels = 225

type pixel struct {
	direction string
	fruit     bool
}

type session struct {
	board            []pixel
	tip              int
	length           int
	currentDirection string
	fruit            int
	eatedAFruit      bool
}

type server struct {
	session *session
}

func main() {
	log.Println("SNAKES ON A MOTHERFUCKING PLATE GETTING IT ON")
	s := server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/screen", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("method not supported"))
			return
		}
		s.getBoard(w, r)
	})
	mux.HandleFunc("/action", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("method not supported"))
			return
		}
		s.input(w, r)
	})

	go gameLoop(&s)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Println(err)
	}
}

func newSession() *session {
	board := make([]pixel, 225)
	board[110] = pixel{direction: "R"}
	board[111] = pixel{direction: "R"}
	board[112] = pixel{direction: "R"}
	f := placeFruit(board)

	sess := session{
		board:            board,
		tip:              112,
		length:           3,
		currentDirection: "R",
		fruit:            f,
	}
	return &sess
}

func (s *server) getBoard(w http.ResponseWriter, r *http.Request) {
	if s.session == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no game")) // TODO return a fancy screen here
		return
	}

	b := boardAsBytes(s.session.board, s.session.fruit)
	log.Printf("board %#v", b)

	_, err := w.Write(b)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *server) input(w http.ResponseWriter, r *http.Request) {
	by, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to read body"))
		return
	}

	cmd := string(by)
	if err = s.updateBoard(cmd); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("failed to update: " + err.Error()))
		if err != nil {
			log.Println(err)
		}
		return
	}

	b := boardAsBytes(s.session.board, s.session.fruit)

	w.Header().Add("Content-Type", "text")
	if _, err = w.Write(b); err != nil {
		log.Println(err)
	}
}

func (s *server) updateBoard(cmd string) error {
	if cmd == "X" {
		s.session = newSession()
		return nil
	}

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
	return nil
}

func boardAsBytes(board []pixel, fruit int) []byte {
	b := make([]byte, pixels)
	for i := range board {
		b[i] = byte('0')
		if board[i].direction != "" {
			b[i] = byte('1')
		}
	}
	b[fruit] = byte('2')

	return b
}

func placeFruit(b []pixel) int {
	var i int
	for {
		i = rand.Intn(pixels)
		if b[i].direction == "" {
			return i
		}
	}
}

func gameLoop(s *server) {
	t := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case <-t.C:
			if s.session == nil {
				continue
			}

			var ateFruit bool

			board := s.session.board
			buf := make([]pixel, pixels)
			for i, p := range s.session.board {
				switch p.direction {
				case "U":
					x := (i - 15) % pixels
					buf[x] = p
					if board[x].fruit {
						ateFruit = true
					}
					if board[i+15].direction == "" && ateFruit {
						buf[i] = p
					}
				case "D":
					x := (i + 15) % pixels
					buf[x] = p
					if board[x].fruit {
						ateFruit = true
					}
					if board[i-15].direction == "" && ateFruit {
						buf[i] = p
					}
				case "L":
					x := (i - 1) % pixels
					buf[x] = p
					if board[x].fruit {
						ateFruit = true
					}
					if board[i+1].direction == "" && ateFruit {
						buf[i] = p
					}
				case "R":
					x := (i + 1) % pixels
					buf[x] = p
					if board[x].fruit {
						ateFruit = true
					}
					if board[i-1].direction == "" && ateFruit {
						buf[i] = p
					}
				}
			}
			s.session.board = buf
		}
	}
}
