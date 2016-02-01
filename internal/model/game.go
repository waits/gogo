package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/garyburd/redigo/redis"
	"log"
	"strconv"
	"time"
)

const StaleGameExpiration = 60 * 60 * 24 * 2
var colors = map[string]int8{"black": 1, "white": 2}

type Game struct {
	Id       string
	Black    string
	White    string
	Size     int
	Handicap int8
	Turn     int
	BlackScr int
	WhiteScr int
	Board    [][]int8 `redis:"-"`
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
	conn := pool.Get()
	defer conn.Close()
	resp, _ := conn.Do("HGETALL", "game:"+id)
	attrs := resp.([]interface{})
	if len(attrs) == 0 {
		return nil, errors.New("load game: game not found")
	}
	g := &Game{}
	redis.ScanStruct(attrs, g)

	grid, _ := redis.String(conn.Do("GET", "game:board:"+id))
	g.Board = make([][]int8, g.Size)
	for y := range g.Board {
		g.Board[y] = make([]int8, g.Size)
		if len(grid) == g.Size*g.Size {
			for x := range g.Board[y] {
				g.Board[y][x] = int8(grid[y*g.Size+x]) - 48
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
	turnstr := strconv.Itoa(1)
	hexid := hashGameParams(black + white + turnstr)
	g := &Game{Id: hexid, White: white, Black: black, Size: size, Turn: 1}
	args := redis.Args{}.Add("game:" + g.Id).AddFlat(g)
	conn := pool.Get()
	conn.Send("HMSET", args...)
	defer conn.Close()
	conn.Do("EXPIRE", "game:"+g.Id, StaleGameExpiration)
	return g, nil
}

// Subscribes to game updates
func Subscribe(id string, callback func(*Game)) {
	conn := redis.PubSubConn{pool.Get()}
	conn.Subscribe("game:" + id)
	for {
		switch reply := conn.Receive().(type) {
		case redis.Message:
			g, err := Load(id)
			if err != nil {
				log.Panicln(err)
			} else {
				callback(g)
			}
		case redis.Subscription:
			log.Printf("Subscribing to channel %s [%d active]\n", reply.Channel, reply.Count)
		case error:
			log.Fatalln(reply)
		}
	}
}

// Makes a move at a given coordinate and saves the game
func (g *Game) Move(c string, mx int, my int) error {
	player := colors[c]
	if player != int8(2 - g.Turn%2) {
		return errors.New("Illegal move: not your turn")
	}
	if g.Board[my][mx] != 0 {
		return errors.New("Illegal move: point already occupied")
	}

	g.Board[my][mx] = player

	point := Point{mx, my}
	captured, err := removeDeadPiecesAround(point, g.Board)
	if err != nil {
		return err
	}

	var grid string
	for _, y := range g.Board {
		for _, x := range y {
			grid += strconv.Itoa(int(x))
		}
	}

	conn := pool.Get()
	defer conn.Close()
	conn.Send("SET", "game:board:"+g.Id, grid, "EX", StaleGameExpiration)
	conn.Send("HINCRBY", "game:"+g.Id, "Turn", 1)
	conn.Send("HINCRBY", "game:"+g.Id, g.Up()+"Scr", captured)
	conn.Send("PUBLISH", "game:"+g.Id, "move")
	conn.Do("EXPIRE", "game:"+g.Id, StaleGameExpiration)

	return nil
}

// Returns the color to move next
func (g *Game) Up() string {
	if g.Turn%2 == 1 {
		return "Black"
	} else {
		return "White"
	}
}
