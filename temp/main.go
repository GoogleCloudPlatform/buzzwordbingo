package main

import (
	"fmt"

	"google.golang.org/api/iterator"
)

var (
	randseedfunc = randomseed
	a            = Agent{ProjectID: "bingo-collab"}
)

func main() {
	game, err := a.GetActiveGame()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Printf("Game Retrieved %v\n", game.ID)
	boards := []Board{}

	iter := client.Collection("boards").Where("game", "==", game.ID).Documents(ctx)

	for {
		b := Board{}
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		doc.DataTo(&b)
		b.ID = doc.Ref.ID

		b, err = a.loadBoardWithPhrases(b)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		boards = append(boards, b)

	}

	batch := client.Batch()
	for _, b := range boards {
		fmt.Printf("Board updating %v\n", b.ID)
		ref := client.Collection("games").Doc(game.ID).Collection("boards").Doc(b.ID)
		batch.Set(ref, b)

	}
	_, err = batch.Commit(ctx)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	for _, b := range boards {
		fmt.Printf("Phrases updating %v\n", b.ID)
		batch := client.Batch()
		for _, p := range b.Phrases {
			fmt.Printf("Phrase updating %v\n", p.ID)
			pref := client.Collection("games").Doc(game.ID).Collection("boards").Doc(b.ID).Collection("phrases").Doc(p.ID)
			batch.Set(pref, b)
		}

		_, err = batch.Commit(ctx)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

	}

}
