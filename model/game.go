package model

import (
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"github.com/garyburd/redigo/redis"
	"log"
	"strconv"
	"strings"
	"time"
)

var staleGameTTL = 60 * 60 * 24 * 7
var colors = [2]string{"Black", "White"}
var hcPts = [9]Point{{15, 3}, {3, 15}, {15, 15}, {3, 3}, {9, 9}, {3, 9}, {15, 9}, {9, 3}, {9, 15}}

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

// New creates a game using a hash of the game parameters
func New(name string, color string, size int, hdcp int) (*Game, error) {
	if size > 19 || size < 9 || size%2 == 0 {
		return nil, errors.New("New game: invalid board size")
	} else if len(name) == 0 {
		return nil, errors.New("New game: name is required")
	} else if len(color) == 0 {
		return nil, errors.New("New game: color is required")
	} else if len(name) > 35 {
		return nil, errors.New("New game: name is too long")
	} else if hdcp > 0 && size != 19 {
		return nil, errors.New("New game: handicap only available on 19x19 boards")
	}
	key := hashGameParams(name, color, strconv.Itoa(size), strconv.Itoa(hdcp))
	g := &Game{Key: key, Size: size, Turn: 1, Ko: -1, Handicap: hdcp}
	if color == "black" {
		g.Black = name
	} else {
		g.White = name
	}
	args := redis.Args{}.Add("game:" + key).AddFlat(g)

	conn := pool.Get()
	defer conn.Close()

	conn.Send("HSET", args...)
	_, err := conn.Do("EXPIRE", "game:"+key, staleGameTTL)
	if err != nil {
		return nil, errors.New("new game: could not connect to database")
	}

	if hdcp > 0 {
		g.Board = parseBoard("", g.Size)
		for i := 0; i < hdcp; i++ {
			p := hcPts[i]
			g.Board[p.Y][p.X] = 1
		}
		bstr := gridBytes(g.Board)
		conn.Send("SET", "game:board:"+g.Key, bstr)
		conn.Send("HSET", "game:"+key, "Turn", 2)
	}

	return g, nil
}

// Load returns a game from the database for a provided key
func Load(key string) (*Game, error) {
	conn := pool.Get()
	defer conn.Close()
	if strings.Contains(key, "-vs-") {
		key, _ = redis.String(conn.Do("HGET", "games", key))
	}
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
	g.Board = parseBoard(grid, g.Size)

	return g, nil
}

// Join adds a player to a game
func (g *Game) Join(name string) string {
	conn := pool.Get()
	defer conn.Close()

	key := "game:" + g.Key
	var c string
	if g.Black == "" {
		c = "Black"
	} else {
		c = "White"
	}

	conn.Send("HSET", key, c, name)
	conn.Send("LPUSH", "games", g.Key)
	conn.Do("PUBLISH", key, "join")
	return strings.ToLower(c)
}

// Recent returns a slice of recently created games
func Recent() []*Game {
	conn := pool.Get()
	defer conn.Close()
	keys, err := redis.Strings(conn.Do("LRANGE", "games", 0, -1))
	if err != nil {
		panic(err)
	}

	games := make([]*Game, 0, len(keys))
	for _, k := range keys {
		if g, err := Load(k); err == nil {
			games = append(games, g)
		} else {
			conn.Do("LREM", "games", 0, k)
		}
	}

	return games
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
func (g *Game) Move(color string, p Point) error {
	player := colorInt(color)
	if g.Last == "f" {
		return errors.New("Illegal move: game over")
	} else if player != 2-g.Turn%2 {
		return errors.New("Illegal move: not your turn")
	} else if g.Board[p.Y][p.X] != 0 {
		return errors.New("Illegal move: point already occupied")
	} else if g.Ko == p.X*19+p.Y {
		return errors.New("Illegal move: ko")
	}

	g.Board[p.Y][p.X] = player

	captured, err := p.CheckLife(g.Board)
	if err != nil {
		return err
	}
	if len(captured) == 1 {
		g.Ko = captured[0].X*19 + captured[0].Y
	} else {
		g.Ko = -1
	}

	bstr := gridBytes(g.Board)
	g.Last = strconv.Itoa(p.X*19 + p.Y)
	g.save(len(captured), bstr, 0, 0)

	return nil
}

// Pass increments the turn number without making a move
func (g *Game) Pass(color string) error {
	player := colorInt(color)
	if g.Last == "f" {
		return errors.New("Illegal move: game over")
	} else if player != 2-g.Turn%2 {
		return errors.New("Illegal move: not your turn")
	}

	g.Ko = -2
	blackArea, whiteArea := 0, 0
	if g.Last == "p" {
		blackArea, whiteArea = g.scoreBoard()
		g.Last = "f"
	} else {
		g.Last = "p"
	}
	g.save(0, "", blackArea, whiteArea)

	return nil
}

// ZeroSize returns one less than the game board size
func (g *Game) ZeroSize() int {
	return g.Size - 1
}

// Save persists the game to the database
func (g *Game) save(cap int, grid string, blackPts int, whitePts int) {
	log.Printf("save %d %d\n", blackPts, whitePts)
	conn := pool.Get()
	defer conn.Close()

	key := "game:" + g.Key
	if grid != "" {
		conn.Send("SET", "game:board:"+g.Key, grid)
	}
	if cap > 0 {
		conn.Send("HINCRBY", key, colors[1-g.Turn%2]+"Scr", cap)
	}
	if blackPts > 0 {
		conn.Send("HINCRBY", key, "BlackScr", blackPts)
	}
	if whitePts > 0 {
		conn.Send("HINCRBY", key, "WhiteScr", whitePts)
	}
	conn.Send("HINCRBY", key, "Turn", 1)
	conn.Send("HSET", key, "Ko", g.Ko)
	conn.Send("HSET", key, "Last", g.Last)
	conn.Send("EXPIRE", key, staleGameTTL)
	conn.Send("EXPIRE", "game:board:"+g.Key, staleGameTTL)
	conn.Do("PUBLISH", key, "move")
}

// Returns area controlled by each color
func (g *Game) scoreBoard() (blackArea int, whiteArea int) {
	log.Printf("scoreBoard %d %d\n", blackArea, whiteArea)
	countedPoints := make([]Point, 0, g.Size*g.Size)

	for y, row := range g.Board {
		for x, color := range row {
			p := Point{x, y}
			if p.inSet(countedPoints) {
				continue
			}

			switch color {
			case empty:
				points, owner := p.searchArea(g.Board, nil, empty)
				switch owner {
				case black:
					blackArea += len(points)
				case white:
					whiteArea += len(points)
				}
				countedPoints = append(countedPoints, points...)
			case black:
				blackArea += 1
			case white:
				whiteArea += 1
			}
		}
	}

	log.Printf("scoreBoard %d %d\n", blackArea, whiteArea)

	return blackArea, whiteArea
}

func colorInt(color string) int {
	if color == "black" {
		return 1
	}
	return 2
}

// Converts a binary string into a two-dimensional slice
func parseBoard(grid string, size int) [][]int {
	board := make([][]int, size)
	for y := range board {
		board[y] = make([]int, size)
		if len(grid) > 0 {
			for x := range board[y] {
				bit := (y*size + x) * 2
				board[y][x] = int((grid[bit/8] >> uint(bit%8)) & 3)
			}
		}
	}
	return board
}

// Converts a two-dimensional slice into a binary string
func gridBytes(board [][]int) string {
	l := len(board)
	bytesize := l*l/4 + 1
	grid := make([]byte, bytesize, bytesize)
	for row, y := range board {
		for col, x := range y {
			bit := (row*l + col) * 2
			grid[bit/8] |= byte(x) << uint(bit%8)
		}
	}
	return string(grid[:bytesize])
}

// Returns the checksum of the game parameters truncated to 48 bits
func hashGameParams(params ...string) string {
	time := time.Now().Unix()
	pstr := strings.Join(params, "\n")
	uniq := []byte(strconv.FormatInt(time, 10) + pstr)
	checksum := sha256.Sum224(uniq)
	return strings.ToLower(base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString(checksum[:6]))
}
