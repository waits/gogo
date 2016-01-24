package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
// 	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"time"
)

type Game struct {
	Id       string
	Black    string
	White    string
	Size     int
	Handicap int
	Turn     int
	Board    [][]int8
	Captured [2]int16
}

func hashGameParams(params string) string {
	time := time.Now().Unix()
	uniq := []byte(strconv.FormatInt(time, 10) + params)
	checksum := sha256.Sum224(uniq)
	hexid := hex.EncodeToString(checksum[:8])
	return hexid
}

// Creates a game using a hash of the game parameters as an ID
func New(black string, white string, size int) *Game {
	sizestr := strconv.Itoa(size)
	turnstr := strconv.Itoa(1)
	hexid := hashGameParams(black + white + turnstr)
	g := &Game{Id: hexid, White: white, Black: black, Size: size, Turn: 1}
	conn.Do("HMSET", "game:"+g.Id, "black", g.Black, "white", g.White, "size", sizestr, "turn", turnstr)
	conn.Do("EXPIRE", "game:"+g.Id, 86400*7)
	conn.Do("SETEX", "game:board:"+g.Id, 86400*7, strings.Repeat("0", size * size))
	return g
}

// Loads a game for a provided id
func Load(id string) (*Game, error) {
	resp, err := redis.StringMap(conn.Do("HGETALL", "game:"+id))
	if err != nil {
		return nil, err
	} else if len(resp) == 0 {
		return nil, errors.New("game.Load: game not found")
	}
	size, _ := strconv.Atoi(resp["size"])
	turn, _ := strconv.Atoi(resp["turn"])
	g := &Game{Id: id, Black: resp["black"], White: resp["white"], Size: size, Turn: turn}

	bstr, err := redis.String(conn.Do("GET", "game:board:"+id))
	if len(bstr) != size * size {
		return nil, errors.New("game.Load: game board is corrupt")
	}
	g.Board = make([][]int8, size)
	for x := range g.Board {
		g.Board[x] = make([]int8, size)
		for y := range g.Board[x] {
			g.Board[x][y] = int8(bstr[x*y]) - 48
		}
	}

	return g, nil
}

// Returns the name of the current player
func (g *Game) Up() string {
	if g.Turn%2 == 0 {
		return g.Black
	} else {
		return g.White
	}
}
