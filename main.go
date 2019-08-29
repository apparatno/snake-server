package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

type session struct {
	board []byte
	tip int
	length int
	currentDirection string
}

func main() {
	board := make([]byte, 215)
	board[110] = 1
	board[111] = 1
	board[112] = 1
	sess := session{
		board: board,
		tip: 112,
		length: 3,
		currentDirection: "R",
	}
	
	/*
	TODO
	- listen for http
	- create new session when X is received
	- run game loop for session (update board based on direction)
	 */
}

func (b *session) getBoard(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(b.board)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (b *session) input(w http.ResponseWriter, r *http.Request) {
	by, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to read body"))
		return
	}

	cmd := string(by)
	if err = b.updateBoard(cmd); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		_, err = w.Write([]byte("failed to update: " + err.Error()))
		if err != nil {
			log.Println(err)
		}
		return
	}

	if _ ,err = w.Write(b); err != nil {
		log.Println(err)
	}
}

func (s *session) updateBoard(cmd string) error {
	switch(cmd) {
	case "U":
		if s.currentDirection == "U" || s.currentDirection == "D" {
			return nil // cannot move up or down now
		}
	case "D":
		if s.currentDirection == "U" || s.currentDirection == "D" {
			return nil // cannot move up or down now
		}
	case "L":
		if s.currentDirection == "L" || s.currentDirection == "R" {
			return nil // cannot move left or right now
		}
	case "R":
		if s.currentDirection == "L" || s.currentDirection == "R" {
			return nil // cannot move left or right now
		}
	case "X":
		s.board = make([]byte, 215)
		return nil // TODO
	default:
		return errors.New("unknown command " + cmd)
	}
	s.currentDirection = cmd
	return nil
}

