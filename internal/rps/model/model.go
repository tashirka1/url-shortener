package model

type JoinRow struct {
	Payload   string
	MetaKey   string
	MetaValue string
	ID        int64
	Ts        int64
	Duration  int64
}
