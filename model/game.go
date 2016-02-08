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

const staleGameTTL = 60 * 60 * 24 * 2

var colors = [2]string{"Black", "White"}

// Game holds the parameters representing a game of Go
type Game struct {
	Key      string `redis:"-"`
	Black    string
	White    string
	Size     int
	Handicap int
	Turn     int
	BlackScr int
	WhiteScr int
	Ko       int
	Last     string
	Board    [][]int `redis:"-"`
}

// Load returns a game from the database for a provided key
func Load(key string) (*Game, error) {
	conn := pool.Get()
	defer conn.Close()
	resp, err := conn.Do("HGETALL", "game:"+key)
	if err != nil {
		return nil, errors.New("load game: could not connect to database")
	}
	attrs := resp.([]interface{})
	if len(attrs) == 0 {
		return nil, errors.New("load game: game not found")
	}
	g := &Game{Key: key}
	redis.ScanStruct(attrs, g)

	grid, _ := redis.String(conn.Do("GET", "game:board:"+key))
	g.Board = make([][]int, g.Size)
	for y := range g.Board {
		g.Board[y] = make([]int, g.Size)
		if len(grid) == g.Size*g.Size {
			for x := range g.Board[y] {
				g.Board[y][x] = int(grid[y*g.Size+x]) - 48
			}
		}
	}

	return g, nil
}

// New creates a game using a hash of the game parameters
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
	g := &Game{Key: hexid, White: white, Black: black, Size: size, Turn: 1}
	args := redis.Args{}.Add("game:" + g.Key).AddFlat(g)

	conn := pool.Get()
	defer conn.Close()

	conn.Send("HMSET", args...)
	_, err := conn.Do("EXPIRE", "game:"+g.Key, staleGameTTL)
	if err != nil {
		return nil, errors.New("new game: could not connect to database")
	}

	return g, nil
}

// Subscribe creates a Redis subscription for a key
func Subscribe(key string, callback func(*Game)) {
	conn := redis.PubSubConn{Conn: pool.Get()}
	conn.Subscribe("game:" + key)
	for {
		switch reply := conn.Receive().(type) {
		case redis.Message:
			g, err := Load(key)
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

// Move makes a move at a given coordinate and saves the game
func (g *Game) Move(color int, mx int, my int) error {
	if g.Last == "f" {
		return errors.New("Illegal move: game over")
	} else if color != 2-g.Turn%2 {
		return errors.New("Illegal move: not your turn")
	} else if g.Board[my][mx] != 0 {
		return errors.New("Illegal move: point already occupied")
	} else if g.Ko == mx*19+my {
		return errors.New("Illegal move: ko")
	}

	g.Board[my][mx] = color

	point := Point{mx, my}
	captured, err := point.CheckLife(g.Board)
	if err != nil {
		return err
	}
	if len(captured) == 1 {
		g.Ko = captured[0].X*19 + captured[0].Y
	} else {
		g.Ko = -1
	}

	var grid string
	for _, y := range g.Board {
		for _, x := range y {
			grid += strconv.Itoa(x)
		}
	}

	g.Last = strconv.Itoa(mx*19 + my)
	g.Save(len(captured), grid)

	return nil
}

// Pass increments the turn number without making a move
func (g *Game) Pass(color int) error {
	if g.Last == "f" {
		return errors.New("Illegal move: game over")
	} else if color != 2-g.Turn%2 {
		return errors.New("Illegal move: not your turn")
	}

	g.Ko = -2
	if g.Last == "p" {
		g.Last = "f"
	} else {
		g.Last = "p"
	}
	g.Save(0, "")

	return nil
}

// Save persists the game to the database
func (g *Game) Save(cap int, grid string) {
	conn := pool.Get()
	defer conn.Close()

	key := "game:" + g.Key
	if grid != "" {
		conn.Send("SET", "game:board:"+g.Key, grid)
	}
	if cap > 0 {
		conn.Send("HINCRBY", key, colors[1-g.Turn%2]+"Scr", cap)
	}
	conn.Send("HINCRBY", key, "Turn", 1)
	conn.Send("HSET", key, "Ko", g.Ko)
	conn.Send("HSET", key, "Last", g.Last)
	conn.Send("EXPIRE", key, staleGameTTL)
	conn.Send("EXPIRE", "game:board:"+g.Key, staleGameTTL)
	conn.Do("PUBLISH", key, "move")
}

// Returns the SHA-224 checksum of the game parameters truncated to 64 bits
func hashGameParams(params string) string {
	time := time.Now().Unix()
	uniq := []byte(strconv.FormatInt(time, 10) + params)
	checksum := sha256.Sum224(uniq)
	hexid := hex.EncodeToString(checksum[:8])
	return hexid
}
