package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
)

const StaleGameExpiration = 60 * 60 * 24 * 2

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
		return nil, errors.New("Load game: game not found")
	}
	size, _ := strconv.Atoi(resp["size"])
	turn, _ := strconv.Atoi(resp["turn"])
	g := &Game{Id: id, Black: resp["black"], White: resp["white"], Size: size, Turn: turn}

	grid, err := redis.String(conn.Do("GET", "game:board:"+id))
	g.Board = make([][]int8, size)
	for y := range g.Board {
		g.Board[y] = make([]int8, size)
		if len(grid) == size*size {
			for x := range g.Board[y] {
				g.Board[y][x] = int8(grid[y*size+x]) - 48
			}
		}
	}

	return g, nil
}

// Creates a game using a hash of the game parameters as an ID
func New(black string, white string, size int) (*Game, error) {
	if size > 19 || size < 9 || size%2 == 0 {
		return nil, errors.New("New game: invalid board size")
	} else if len(black) == 0 || len(white) == 0 {
		return nil, errors.New("New game: missing player name(s)")
	} else if len(black) > 35 || len(white) > 35 {
		return nil, errors.New("New game: player name(s) are too long")
	}
	sizestr := strconv.Itoa(size)
	turnstr := strconv.Itoa(1)
	hexid := hashGameParams(black + white + turnstr)
	g := &Game{Id: hexid, White: white, Black: black, Size: size, Turn: 1}
	conn.Do("HMSET", "game:"+g.Id, "black", g.Black, "white", g.White, "size", sizestr, "turn", turnstr)
	conn.Do("EXPIRE", "game:"+g.Id, StaleGameExpiration)
	return g, nil
}

// Makes a move at a given coordinate and saves the game
func (g *Game) Move(mx int, my int) error {
	if g.Board[my][mx] != 0 {
		return errors.New("Illegal move: point already occupied")
	}

	color := int8(2 - g.Turn%2)
	g.Board[my][mx] = color
	g.Turn += 1

	point := Point{mx, my}
	adjacentOpponents := piecesAdjacentTo(point, 3-color, g.Board)
	for _, opp := range adjacentOpponents {
		pieces := deadPiecesAdjacentTo(opp, g.Board, make([]Point, 0, 180))
		if pieces != nil {
			clearPoints(pieces, g.Board)
		}
	}


	var grid string
	for _, y := range g.Board {
		for _, x := range y {
			grid += strconv.Itoa(int(x))
		}
	}

	conn.Do("SET", "game:board:"+g.Id, grid, "EX", StaleGameExpiration)
	conn.Do("HINCRBY", "game:"+g.Id, "turn", 1)
	conn.Do("EXPIRE", "game:"+g.Id, StaleGameExpiration)

	return nil
}

// Returns the name of the current player
func (g *Game) Up() string {
	if g.Turn%2 == 1 {
		return g.Black
	} else {
		return g.White
	}
}
