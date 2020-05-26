package main

import (
	"fmt"
	"log"
)

// ErrCacheMiss error indicates that an item is not in the cache
var ErrCacheMiss = fmt.Errorf("item is not in cache")

// NewCache returns an initialized cache ready to go.
func NewCache() (Cache, error) {
	c := Cache{}
	c.Init()
	return c, nil
}

// Cache abstracts all of the operations of caching for the application
type Cache struct {
	boards map[string]Board
	games  map[string]Game
}

func (c *Cache) log(msg string) {
	if noisy {
		log.Printf("Cache: %s\n", msg)
	}
}

// Init starts the cache off
func (c *Cache) Init() {
	c.boards = make(map[string]Board)
	c.games = make(map[string]Game)
}

// Clear removes all items from the cache.
func (c *Cache) Clear() error {
	c.Init()

	return nil
}

// SaveBoard records a board into the cache.
func (c *Cache) SaveBoard(b Board) error {
	c.boards[b.ID] = b
	c.boards[b.Game+"_"+b.Player.Email] = b

	return nil
}

// SaveGame records a game in the cache.
func (c *Cache) SaveGame(g Game) error {
	c.games[g.ID] = g
	c.games["active"] = g

	return nil
}

// GetGame retrieves an game from the cache
func (c *Cache) GetGame(key string) (Game, error) {
	g, ok := c.games[key]
	if !ok {
		return Game{}, ErrCacheMiss
	}

	return g, nil
}

// GetBoard retrieves an board from the cache
func (c *Cache) GetBoard(key string) (Board, error) {
	b, ok := c.boards[key]
	if !ok {
		return Board{}, ErrCacheMiss
	}

	return b, nil
}

// DeleteBoard will remove a board from the cache completely.
func (c *Cache) DeleteBoard(board Board) error {
	delete(c.boards, board.ID)
	delete(c.boards, board.Game+"_"+board.Player.Email)
	c.log(fmt.Sprintf("Cleaning from cache %s", board.ID))
	c.log(fmt.Sprintf("Cleaning from cache %s", board.Game+"_"+board.Player.Email))
	return nil
}
