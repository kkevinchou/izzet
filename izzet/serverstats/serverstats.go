package serverstats

type Stat struct {
	Name  string
	Value string
}

type ServerStats struct {
	Data []Stat
}
