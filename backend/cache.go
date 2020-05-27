package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gomodule/redigo/redis"
)

// ErrCacheMiss error indicates that an item is not in the cache
var ErrCacheMiss = fmt.Errorf("item is not in cache")

// NewCache returns an initialized cache ready to go.
func NewCache(redisHost, redisPort string) (Cache, error) {
	c := Cache{}
	c.Init(redisHost, redisPort)
	return c, nil
}

// Cache abstracts all of the operations of caching for the application
type Cache struct {
	redisPool *redis.Pool
}

func (c *Cache) log(msg string) {
	if noisy {
		log.Printf("Cache: %s\n", msg)
	}
}

// Init starts the cache off
func (c *Cache) Init(redisHost, redisPort string) {
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)
	msg := fmt.Sprintf("Initialized Redis at %s", redisAddr)
	c.log(msg)
	const maxConnections = 10
	c.redisPool = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", redisAddr)
	}, maxConnections)

}

// Clear removes all items from the cache.
func (c *Cache) Clear() error {
	conn := c.redisPool.Get()
	defer conn.Close()

	if _, err := conn.Do("FLUSHALL"); err != nil {
		return err
	}
	return nil
}

// SaveBoard records a board into the cache.
func (c *Cache) SaveBoard(b Board) error {

	conn := c.redisPool.Get()
	defer conn.Close()

	json, err := b.JSON()
	if err != nil {
		return err
	}

	if _, err := conn.Do("SET", b.ID, json); err != nil {
		return err
	}

	if _, err := conn.Do("SET", b.Game+"_"+b.Player.Email, json); err != nil {
		return err
	}
	c.log("Successfully saved board to cache")
	return nil
}

// SaveGame records a game in the cache.
func (c *Cache) SaveGame(g Game) error {

	conn := c.redisPool.Get()
	defer conn.Close()

	json, err := g.JSON()
	if err != nil {
		return err
	}

	if _, err := conn.Do("SET", g.ID, json); err != nil {
		return err
	}
	c.log("Successfully saved game to cache")
	return nil
}

// GetGame retrieves an game from the cache
func (c *Cache) GetGame(key string) (Game, error) {

	conn := c.redisPool.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", key))
	if err == redis.ErrNil {
		return Game{}, ErrCacheMiss
	} else if err != nil {
		return Game{}, err
	}

	g := Game{}
	if err := json.Unmarshal([]byte(s), &g); err != nil {
		return Game{}, err
	}
	c.log("Successfully retrieved game from cache")

	return g, nil
}

// GetBoard retrieves an board from the cache
func (c *Cache) GetBoard(key string) (Board, error) {
	conn := c.redisPool.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", key))
	if err == redis.ErrNil {
		return Board{}, ErrCacheMiss
	} else if err != nil {
		return Board{}, err
	}

	b := Board{}
	if err := json.Unmarshal([]byte(s), &b); err != nil {
		return Board{}, err
	}
	c.log("Successfully retrieved board from cache")

	return b, nil
}

// DeleteBoard will remove a board from the cache completely.
func (c *Cache) DeleteBoard(board Board) error {
	conn := c.redisPool.Get()
	defer conn.Close()

	if _, err := conn.Do("DEL", board.ID); err != nil {
		return err
	}

	if _, err := conn.Do("DEL", board.Game+"_"+board.Player.Email); err != nil {
		return err
	}

	c.log(fmt.Sprintf("Cleaning from cache %s", board.ID))
	c.log(fmt.Sprintf("Cleaning from cache %s", board.Game+"_"+board.Player.Email))
	return nil
}
