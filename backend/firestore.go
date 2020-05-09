package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

var (
	client *firestore.Client
	ctx    = context.Background()
	noisy  = true
)

// Agent is a go between for the main application and firestore.
type Agent struct {
	ProjectID string
}

func (a *Agent) log(msg string) {
	if noisy {
		fmt.Printf("Firestore: %s\n", msg)
	}
}

func (a *Agent) getClient() (*firestore.Client, error) {
	if client != nil {
		return client, nil
	}
	a.log("Getting New Client")
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

	a.log("Getting Phrases")
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
	g.Active = true
	g.Master.Load(phrases)

	client, err := a.getClient()

	if err != nil {
		return g, fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Creating new game")
	doc, _, err := client.Collection("games").Add(ctx, g)
	if err != nil {
		return g, fmt.Errorf("failed to add game to firestore: %v", err)
	}

	g.ID = doc.ID

	a.log("Adding phrases to new game")
	batch := client.Batch()

	for _, v := range g.Master.Records {
		ref := client.Collection("games").Doc(g.ID).Collection("records").Doc(v.Phrase.ID)
		batch.Set(ref, v)
	}

	m := Message{}
	m.SetText("Game has begun!")
	m.SetAudience("all")

	mref := client.Collection("games").Doc(g.ID).Collection("messages").Doc("00001")
	batch.Set(mref, m)

	_, err = batch.Commit(ctx)
	if err != nil {
		return g, fmt.Errorf("failed to add records to database: %v", err)
	}

	return g, nil
}

func (a *Agent) GetGame(id string) (Game, error) {
	g := Game{}

	client, err := a.getClient()
	if err != nil {
		return g, fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Getting existing game")
	doc, err := client.Collection("games").Doc(id).Get(ctx)
	if err != nil {
		return g, fmt.Errorf("failed to get game: %v", err)
	}

	doc.DataTo(&g)
	g.ID = id
	g, err = a.loadGameWithRecords(g)
	if err != nil {
		return g, fmt.Errorf("failed to load records for game: %v", err)
	}

	return g, nil
}

func (a *Agent) AddMessageToGame(g Game, m Message) error {
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Adding message to game")
	_, err = client.Collection("games").Doc(g.ID).Collection("messages").Doc(timestamp).Set(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to send message : %v", err)
	}

	return nil
}

func (a *Agent) loadGameWithRecords(g Game) (Game, error) {
	client, err := a.getClient()
	if err != nil {
		return g, fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Loading records from game")
	iter := client.Collection("games").Doc(g.ID).Collection("records").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("failed getting game records: %v", err)
		}
		r := Record{}
		doc.DataTo(&r)
		g.Master.Records = append(g.Master.Records, r)
	}

	return g, nil
}

func (a *Agent) UpdateRecordOnGame(g Game, r Record) error {
	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Updating game record")
	ref := client.Collection("games").Doc(g.ID).Collection("records").Doc(r.Phrase.ID)

	if _, err := ref.Set(ctx, r); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

func (a *Agent) SaveGame(g Game) error {

	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// TODO: Add code to allow merging instead of overwrites.
	a.log("Save game")
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

	a.log("Getting board")
	doc, err := client.Collection("boards").Doc(id).Get(ctx)
	if err != nil {
		return b, fmt.Errorf("failed to get board: %v", err)
	}

	doc.DataTo(&b)
	b.ID = id
	b, err = a.loadBoardWithPhrases(b)
	if err != nil {
		return b, fmt.Errorf("failed to load phrases for board: %v", err)
	}
	return b, nil
}

func (a *Agent) loadBoardWithPhrases(b Board) (Board, error) {
	client, err := a.getClient()
	if err != nil {
		return b, fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Adding phrases to existing board")
	iter := client.Collection("boards").Doc(b.ID).Collection("phrases").OrderBy("DisplayOrder", firestore.Asc).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return b, fmt.Errorf("failed getting board records: %v", err)
		}
		p := Phrase{}
		doc.DataTo(&p)
		b.Phrases = append(b.Phrases, p)
	}

	return b, nil
}

func (a *Agent) SaveBoard(b Board) (Board, error) {
	client, err := a.getClient()
	if err != nil {
		return b, fmt.Errorf("failed to create client: %v", err)
	}

	if b.ID != "" {
		_, err = client.Collection("boards").Doc(b.ID).Set(ctx, b)
		if err != nil {
			return b, fmt.Errorf("failed to update board: %v", err)
		}

		return b, nil
	}
	a.log("Saving board")
	doc, _, err := client.Collection("boards").Add(ctx, b)
	if err != nil {
		return b, fmt.Errorf("failed to add board: %v", err)
	}
	b.ID = doc.ID

	if err := a.savePhrases(b); err != nil {
		return b, fmt.Errorf("failed to save phrases: %v", err)
	}

	return b, nil

}

func (a *Agent) savePhrases(b Board) error {

	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Adding phrases to board")
	batch := client.Batch()

	for _, v := range b.Phrases {
		ref := client.Collection("boards").Doc(b.ID).Collection("phrases").Doc(v.ID)
		batch.Set(ref, v)
	}
	_, err = batch.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to add records to database: %v", err)
	}

	return nil
}

func (a *Agent) UpdatePhraseOnBoard(b Board, p Phrase) error {
	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Updating phrase on board")
	ref := client.Collection("boards").Doc(b.ID).Collection("phrases").Doc(p.ID)

	if _, err := ref.Set(ctx, p); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

func (a *Agent) UpdateBingoOnBoard(b Board, bingo bool) error {
	client, err := a.getClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Updating board to bingo")
	ref := client.Collection("boards").Doc(b.ID)

	if _, err := ref.Set(ctx, map[string]interface{}{"BingoDeclared": bingo}, firestore.MergeAll); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

func (a *Agent) GetActiveGame() (Game, error) {
	client, err := a.getClient()
	g := Game{}

	if err != nil {
		return g, fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Gettign active game")
	iter := client.Collection("games").Where("Active", "==", true).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("failed to iterate over response from firestore: %v", err)
		}
		doc.DataTo(&g)
		g.ID = doc.Ref.ID
		break
	}

	if g.ID == "" {
		return a.NewGame("Default Bingo Board")
	}

	g, err = a.loadGameWithRecords(g)
	if err != nil {
		return g, fmt.Errorf("failed to load records for game: %v", err)
	}

	return g, nil
}

func (a *Agent) ResetActiveGame() (Game, error) {
	g := Game{}
	client, err := a.getClient()
	if err != nil {
		return g, fmt.Errorf("failed to create client: %v", err)
	}

	a.log("Reset game")
	g, err = a.GetActiveGame()
	if err != nil {
		return g, fmt.Errorf("failed to get active game: %v", err)
	}

	iter := client.Collection("games").Where("Active", "==", true).Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("failed to iterate over response from firestore: %v", err)
		}
		doc.DataTo(&g)
		g.ID = doc.Ref.ID
		break
	}

	a.log("removing messages from game")
	ref := client.Collection("games").Doc(g.ID).Collection("messages")
	for {
		// Get a batch of documents
		iter := ref.Limit(100).Documents(ctx)
		numDeleted := 0

		// Iterate through the documents, adding
		// a delete operation for each one to a
		// WriteBatch.
		batch := client.Batch()
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return g, fmt.Errorf("failed to clean messages from firestore: %v", err)
			}

			batch.Delete(doc.Ref)
			numDeleted++
		}

		// If there are no documents to delete,
		// the process is over.
		if numDeleted == 0 {
			break
		}

		_, err := batch.Commit(ctx)
		if err != nil {
			return g, fmt.Errorf("failed to clean messages from firestore: %v", err)
		}
	}

	return g, nil
}

func (a *Agent) GetBoardForPlayer(id string, email string) (Board, error) {
	b := Board{}
	client, err := a.getClient()
	if err != nil {
		return b, fmt.Errorf("failed to create client: %v", err)
	}

	a.log("get board from player")
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

	if b.ID != "" {
		b, err = a.loadBoardWithPhrases(b)
		if err != nil {
			return b, fmt.Errorf("failed to load phrases for board: %v", err)
		}
	}

	return b, nil

}
