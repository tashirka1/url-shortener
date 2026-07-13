package model

import (
	"errors"
	"time"
)

const MaxURLLength = 2048

var ErrLinkAlreadyExists = errors.New("link already exists")
var ErrAliasTaken = errors.New("alias already taken")

type Link struct {
	CreatedAt time.Time
	Code      string
	Url       string
	Id        int64
	Clicks    int
}
