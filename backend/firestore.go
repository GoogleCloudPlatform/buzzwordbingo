package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var (
	client *firestore.Client
	ctx    = context.Background()
)

// Agent is a go between for the main application and firestore.
type Agent struct {
	ProjectID string
}

func (a *Agent) getClient() (*firestore.Client, error) {
	if client != nil {
		return client, nil
	}
	return firestore.NewClient(context.Background(), a.ProjectID)
}

// GetPhrases fetches the list of nodes from Firestore and arranges them
// into a route.Route
func (a *Agent) GetPhrases() ([]Phrase, error) {

	p := []Phrase{}

	client, err := a.getClient()
	if err != nil {
		return p, fmt.Errorf("Failed to create client: %v", err)
	}

	iter := client.Collection("phrases").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return p, fmt.Errorf("Failed to iterate: %v", err)
		}
		var phrase Phrase
		dataMap := doc.Data()

		phrase.ID = dataMap["id"].(string)
		phrase.Text = dataMap["phrase"].(string)

		p = append(p, phrase)
	}

	return p, nil
}

func (a *Agent) NewGame(name string) (Game, error) {
	g := Game{}

	phrases, err := a.GetPhrases()
	if err != nil {
		return g, fmt.Errorf("failed to get phrases client: %v", err)
	}

	g.Name = name
	g.Master.Load(phrases)

	client, err := a.getClient()

	if err != nil {
		return g, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	doc, _, err := client.Collection("games").Add(ctx, g)
	if err != nil {
		return g, fmt.Errorf("failed to add game to firestore: %v", err)
	}

	g.ID = doc.ID

	return g, nil
}

func (a *Agent) GetGame(id string) (Game, error) {
	g := Game{}

	client, err := a.getClient()
	if err != nil {
		return g, fmt.Errorf("failed to create client: %v", err)
	}

	doc, err := client.Collection("games").Doc(id).Get(ctx)
	if err != nil {
		return g, fmt.Errorf("failed to get game: %v", err)
	}

	doc.DataTo(&g)
	g.ID = id

	return g, nil
}

func (a *Agent) SaveGame(g Game) error {

	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// TODO: Add code to allow merging instead of overwrites.
	_, err = client.Collection("games").Doc(g.ID).Set(ctx, g)
	if err != nil {
		return fmt.Errorf("failed to get game: %v", err)
	}

	return nil
}

func (a *Agent) GetBoard(id string) (Board, error) {
	b := Board{}

	client, err := a.getClient()
	if err != nil {
		return b, fmt.Errorf("failed to create client: %v", err)
	}

	doc, err := client.Collection("boards").Doc(id).Get(ctx)
	if err != nil {
		return b, fmt.Errorf("failed to get game: %v", err)
	}

	doc.DataTo(&b)
	b.ID = id

	return b, nil
}

func (a *Agent) SaveBoard(b Board) (Board, error) {
	client, err := a.getClient()
	if err != nil {
		return b, fmt.Errorf("failed to create client: %v", err)
	}

	// TODO: Add code to allow merging instead of overwrites.
	if b.ID != "" {
		_, err = client.Collection("boards").Doc(b.ID).Set(ctx, b)
		if err != nil {
			return b, fmt.Errorf("failed to update board: %v", err)
		}

		return b, nil
	}
	doc, _, err := client.Collection("boards").Add(ctx, b)
	if err != nil {
		return b, fmt.Errorf("failed to add board: %v", err)
	}

	b.ID = doc.ID
	return b, nil

}

func (a *Agent) GetActiveGame() (Game, error) {
	client, err := a.getClient()
	response := Game{}

	if err != nil {
		return response, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	iter := client.Collection("games").Where("Active", "==", true).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return response, fmt.Errorf("failed to iterate over response from firestore: %v", err)
		}
		doc.DataTo(&response)
		break
	}

	return response, nil
}

func (a *Agent) GetBoardForPlayer(id string, email string) (Board, error) {
	client, err := a.getClient()
	b := Board{}

	if err != nil {
		return b, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	iter := client.Collection("boards").Where("Game", "==", id).Where("Player.Email", "==", email).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return b, fmt.Errorf("failed to iterate over b from firestore: %v", err)
		}
		doc.DataTo(&b)
		b.ID = doc.Ref.ID
		break
	}

	return b, nil

}
