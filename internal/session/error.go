package session

import "errors"

var (
	ErrNotLoggedIn = errors.New("no Advent of Code session found, log in at " + LoginURL + " in your browser and try again")
)
