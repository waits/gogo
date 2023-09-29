package model

import (
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/waits/gogo/model/game"
)

var games = map[string]Game{
	"1": {Type: game.Online, Size: 9, Turn: 1, Ko: -1, Handicap: 0, Black: "Aaron", White: "Job"},
	"2": {Type: game.Online, Size: 9, Turn: 1, Ko: -1, Handicap: 0, Black: "Frank"},
	"3": {Type: game.Online, Size: 13, Turn: 1, Ko: -1, Handicap: 0, Black: "Blake", White: "Tracy"},
}

// Flushes the database and sets up test data
func init() {
	pool := InitPool(1)
	conn := pool.Get()
	defer conn.Close()

	conn.Send("FLUSHDB")
	for k, g := range games {
		args := redis.Args{}.Add("game:" + k).AddFlat(g)
		conn.Send("HMSET", args...)
		conn.Send("LPUSH", "games", k)
	}
	conn.Flush()
}

func TestNew(t *testing.T) {
	_, err := New(game.Online, "Andy", "white", 13, 0)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestLoad(t *testing.T) {
	_, err := Load("1")
	if err != nil {
		t.Error(err.Error())
	}
}

func TestRecent(t *testing.T) {
	games := Recent()
	if c := len(games); c != 3 {
		t.Errorf("method returned wrong number of games: got %v want %v", c, 3)
	}
}

func TestMove(t *testing.T) {
	g, _ := Load("1")
	err := g.Move("black", Point{3, 4})
	if err != nil {
		t.Error(err.Error())
	}
	err = g.Move("white", Point{1, 1})
	if err == nil {
		t.Error("allowed wrong player to move")
	}
	err = g.Move("black", Point{3, 4})
	if err == nil {
		t.Error("allowed move on occupied point")
	}
}

func TestPass(t *testing.T) {
	g, _ := Load("3")
	err := g.Pass("black")
	if err != nil {
		t.Error(err.Error())
	}
}
