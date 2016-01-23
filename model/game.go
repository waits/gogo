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
	Size     uint8
	Handicap uint8
	Turn     uint16
	Board    [19][19]uint8
	Captured [2]uint8
}

// Creates a game using a hash of the game parameters as an ID
func NewGame(black string, white string, size uint8) *Game {
	sizestr := strconv.Itoa(int(size))
	turnstr := strconv.Itoa(1)
	hexid := hashGameParams(black + white + turnstr)
	g := &Game{Id: hexid, White: white, Black: black, Size: size, Turn: 1}
	conn.Do("HMSET", "game:"+g.Id, "black", g.Black, "white", g.White, "size", sizestr, "turn", turnstr)
	conn.Do("EXPIRE", "game:"+g.Id, 86400*7)
	return g
}

// Loads a game for a provided id
func LoadGame(id string) (*Game, error) {
	resp, err := redis.StringMap(conn.Do("HGETALL", "game:"+id))
	if err != nil {
		return nil, err
	} else if len(resp) == 0 {
		return nil, errors.New("LoadGame: game not found")
	}
	size, _ := strconv.Atoi(resp["size"])
	turn, _ := strconv.Atoi(resp["turn"])
	g := &Game{Id: id, Black: resp["black"], White: resp["white"], Size: uint8(size), Turn: uint16(turn)}
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

func hashGameParams(params string) string {
	time := time.Now().Unix()
	uniq := []byte(strconv.FormatInt(time, 10) + params)
	checksum := sha256.Sum224(uniq)
	hexid := hex.EncodeToString(checksum[:8])
	return hexid
}
