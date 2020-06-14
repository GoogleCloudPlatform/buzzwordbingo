package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

type RedisPool interface {
	Get() redis.Conn
}

// Cacher is an interface for testing caching
type Cacher interface {
	Init(string, string)
	Clear() error
	SaveBoard(Board) error
	GetBoard(string) (Board, error)
	DeleteBoard(Board) error
	SaveGame(Game) error
	GetGame(string) (Game, error)
	SaveGamesForKey(string, Games) error
	GetGamesForKey(string) (Games, error)
	DeleteGamesForKey([]string) error
	UpdatePhrase(Game, Phrase) error
	log(string)
	SetRedisPool(RedisPool)
}

// ErrCacheMiss error indicates that an item is not in the cache
var ErrCacheMiss = fmt.Errorf("item is not in cache")

// NewCache returns an initialized cache ready to go.
func NewCache(redisHost, redisPort string, enabled bool) (*Cache, error) {
	c := &Cache{}
	c.Init(redisHost, redisPort)
	c.enabled = enabled
	return c, nil
}

// Cache abstracts all of the operations of caching for the application
type Cache struct {
	// redisPool *redis.Pool
	redisPool RedisPool
	enabled   bool
}

func (c *Cache) log(msg string) {
	if noisy {
		log.Printf("Cache     : %s\n", msg)
	}
}

// Init starts the cache off
func (c Cache) Init(redisHost, redisPort string) {
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	msg := fmt.Sprintf("Initialized Redis at %s", redisAddr)
	c.log(msg)
	const maxConnections = 10
	c.redisPool = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", redisAddr)
	}, maxConnections)
}

func (c *Cache) SetRedisPool(r RedisPool) {

}

// Clear removes all items from the cache.
func (c Cache) Clear() error {
	if !c.enabled {
		return nil
	}
	conn := c.redisPool.Get()
	defer conn.Close()

	if _, err := conn.Do("FLUSHALL"); err != nil {
		return err
	}
	return nil
}

func (c Cache) boardKeys(b Board) (boardkey string, playerkey string) {
	boardkey = "board-" + b.ID
	playerkey = "board-" + b.Game + "_" + b.Player.Email
	return boardkey, playerkey
}

func (c *Cache) gameKey(key string) string {
	return "game-" + key
}

func (c *Cache) gamesKey(key string) string {
	return "games-" + key
}

////////////////////////////////////////////////////////////////////////////////
// BOARDS
////////////////////////////////////////////////////////////////////////////////

// SaveBoard records a board into the cache.
func (c *Cache) SaveBoard(b Board) error {
	if !c.enabled {
		return nil
	}

	conn := c.redisPool.Get()
	defer conn.Close()

	json, err := b.JSON()
	if err != nil {
		return err
	}

	boardkey, playerkey := c.boardKeys(b)

	conn.Send("MULTI")
	conn.Send("SET", boardkey, json)
	conn.Send("SET", playerkey, json)

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}
	c.log("Successfully saved board to cache")
	return nil
}

// GetBoard retrieves an board from the cache
func (c *Cache) GetBoard(key string) (Board, error) {
	b := Board{}
	if !c.enabled {
		return b, ErrCacheMiss
	}
	conn := c.redisPool.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", "board-"+key))
	if err == redis.ErrNil {
		return Board{}, ErrCacheMiss
	} else if err != nil {
		return Board{}, err
	}

	if err := json.Unmarshal([]byte(s), &b); err != nil {
		return Board{}, err
	}
	c.log("Successfully retrieved board from cache")

	return b, nil
}

// DeleteBoard will remove a board from the cache completely.
func (c *Cache) DeleteBoard(board Board) error {
	if !c.enabled {
		return nil
	}
	conn := c.redisPool.Get()
	defer conn.Close()

	boardkey, playerkey := c.boardKeys(board)
	gameskey := c.gamesKey(board.Player.Email)

	conn.Send("MULTI")

	if err := conn.Send("DEL", boardkey); err != nil {
		return err
	}

	if err := conn.Send("DEL", playerkey); err != nil {
		return err
	}

	if err := conn.Send("DEL", gameskey); err != nil {
		return err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	c.log(fmt.Sprintf("Cleaning from cache %s", boardkey))
	c.log(fmt.Sprintf("Cleaning from cache %s", playerkey))
	c.log(fmt.Sprintf("Cleaning from cache %s", gameskey))
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// GAMES
////////////////////////////////////////////////////////////////////////////////

// SaveGame records a game in the cache.
func (c *Cache) SaveGame(g Game) error {
	if !c.enabled {
		return nil
	}

	conn := c.redisPool.Get()
	defer conn.Close()

	json, err := g.JSON()
	if err != nil {
		return err
	}

	gamekey := c.gameKey(g.ID)

	if _, err := conn.Do("SET", gamekey, json); err != nil {
		return err
	}
	c.log("Successfully saved game to cache")

	if len(g.Boards) == 0 {
		c.log("WARNING game saved to cache without the boards.")
	}

	return nil
}

// GetGame retrieves an game from the cache
func (c *Cache) GetGame(key string) (Game, error) {
	g := Game{}
	if !c.enabled {
		return g, ErrCacheMiss
	}

	conn := c.redisPool.Get()
	defer conn.Close()

	gamekey := c.gameKey(key)

	s, err := redis.String(conn.Do("GET", gamekey))
	if err == redis.ErrNil {
		return Game{}, ErrCacheMiss
	} else if err != nil {
		return Game{}, err
	}

	if err := json.Unmarshal([]byte(s), &g); err != nil {
		return Game{}, err
	}
	c.log("Successfully retrieved game from cache")

	return g, nil
}

// SaveGamesForKey saves a list of all of the games a player is in.
func (c *Cache) SaveGamesForKey(key string, g Games) error {
	if !c.enabled {
		return nil
	}

	conn := c.redisPool.Get()
	defer conn.Close()

	json, err := g.JSON()
	if err != nil {
		return err
	}

	rkey := c.gamesKey(key)

	if _, err := conn.Do("SET", rkey, json); err != nil {
		return err
	}
	c.log("Successfully saved game list to cache")
	return nil
}

// GetGamesForKey retrieves a list of games from the cache
func (c *Cache) GetGamesForKey(key string) (Games, error) {
	g := []Game{}
	if !c.enabled {
		return g, ErrCacheMiss
	}
	conn := c.redisPool.Get()
	defer conn.Close()

	rkey := c.gamesKey(key)

	s, err := redis.String(conn.Do("GET", rkey))
	if err == redis.ErrNil {
		return g, ErrCacheMiss
	} else if err != nil {
		return g, err
	}

	if err := json.Unmarshal([]byte(s), &g); err != nil {
		return g, err
	}
	c.log("Successfully retrieved games from cache")

	return g, nil
}

// DeleteGamesForKey will remove the list of games for a particular player
func (c *Cache) DeleteGamesForKey(keys []string) error {
	if !c.enabled {
		return nil
	}
	conn := c.redisPool.Get()
	defer conn.Close()

	conn.Send("MULTI")

	for _, v := range keys {
		rkey := c.gamesKey(v)
		if err := conn.Send("DEL", rkey); err != nil {
			return err
		}
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	c.log(fmt.Sprintf("Cleaning games for key from cache"))
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// PHRASES
////////////////////////////////////////////////////////////////////////////////

// UpdatePhrase will update all of the versions of a phrase in a game and all
// of the boards in that game.
func (c *Cache) UpdatePhrase(g Game, p Phrase) error {
	c.log("Update Phrase " + p.Text)
	conn := c.redisPool.Get()
	defer conn.Close()

	gamekey := c.gameKey(g.ID)

	gjson, err := g.JSON()
	if err != nil {
		return err
	}

	conn.Send("MULTI")
	conn.Send("SET", gamekey, gjson)

	for _, b := range g.Boards {

		boardkey, playerkey := c.boardKeys(b)

		json, err := b.JSON()
		if err != nil {
			return err
		}

		if err := conn.Send("SET", boardkey, json); err != nil {
			return err
		}

		if err := conn.Send("SET", playerkey, json); err != nil {
			return err
		}

	}

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	return nil
}
