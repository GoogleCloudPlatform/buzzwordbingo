package main

import (
	"fmt"
)

var ErrCacheMiss = fmt.Errorf("item is not in cache")

type Cache struct {
}

func (c *Cache) Clear() error {
	boards = make(map[string]Board)
	games = make(map[string]Game)

	return nil
}

func (c *Cache) SaveBoard(b Board) error {
	boards[b.ID] = b
	boards[b.Game+"_"+b.Player.Email] = b

	return nil
}

func (c *Cache) SaveGame(g Game) error {
	games[g.ID] = g
	games["active"] = g

	return nil
}

func (c *Cache) GetGame(key string) (Game, error) {
	g, ok := games[key]
	if !ok {
		return Game{}, ErrCacheMiss
	}

	return g, nil
}

func (c *Cache) GetBoard(key string) (Board, error) {
	b, ok := boards[key]
	if !ok {
		return Board{}, ErrCacheMiss
	}

	return b, nil
}
