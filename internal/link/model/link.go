package model

import (
	"errors"
	"time"
)

const MaxURLLength = 2048

var ErrLinkAlreadyExists = errors.New("link already exists")

type Link struct {
	Id        int64
	Code      string
	Url       string
	Clicks    int
	CreatedAt time.Time
}
