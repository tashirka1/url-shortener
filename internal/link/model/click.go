package model

type LinkClick struct {
	Referrer  string
	UserAgent string
	ClickedAt string
	Id        int64
	LinkId    int64
}

type DailyClick struct {
	Day    string
	Clicks int
}

type ReferrerStat struct {
	Referrer string
	Clicks   int
}
