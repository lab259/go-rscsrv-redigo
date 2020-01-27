package redigosrv

import (
	"errors"
)

func recovery(done chan<- error) {
	if err := recover(); err != nil {
		switch n := err.(type) {
		case string:
			done <- errors.New(n)
		default:
			done <- errors.New("Type not mapped")
		}
	}
}
