package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/garyburd/redigo/redis"
	"strconv"
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

	grid, err := redis.String(conn.Do("GET", "game:board:"+id))
	g.Board = make([][]int8, size)
	for y := range g.Board {
		g.Board[y] = make([]int8, size)
		if len(grid) == size * size {
			for x := range g.Board[y] {
				g.Board[y][x] = int8(grid[y*size+x]) - 48
			}
		}
	}

	return g, nil
}

// Creates a game using a hash of the game parameters as an ID
func New(black string, white string, size int) *Game {
	sizestr := strconv.Itoa(size)
	turnstr := strconv.Itoa(1)
	hexid := hashGameParams(black + white + turnstr)
	g := &Game{Id: hexid, White: white, Black: black, Size: size, Turn: 1}
	conn.Do("HMSET", "game:"+g.Id, "black", g.Black, "white", g.White, "size", sizestr, "turn", turnstr)
	conn.Do("EXPIRE", "game:"+g.Id, 86400*7)
	return g
}

// Makes a move at a given point and saves the game
func (g *Game) Save(mx int, my int) bool {
	g.Turn += 1
	g.Board[my][mx] = int8(g.Turn % 2 + 1)

	var grid string
	for _, y := range g.Board {
		for _, x := range y {
			grid += strconv.Itoa(int(x))
		}
	}

	_, err := conn.Do("SET", "game:board:"+g.Id, grid)
	if err != nil {
		return false
	} else {
		conn.Do("HINCRBY", "game:"+g.Id, "turn", 1)
		return true
	}
}

// Returns the name of the current player
func (g *Game) Up() string {
	if g.Turn%2 == 0 {
		return g.Black
	} else {
		return g.White
	}
}
