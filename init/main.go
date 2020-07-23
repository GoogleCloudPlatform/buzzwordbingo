// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"cloud.google.com/go/firestore"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
)

var (
	projectID = ""
	err       error
	ctx       = context.Background()
	a         Agent
)

func main() {
	projectID, err = getProjectID()
	if err != nil {
		log.Fatal(err)
	}

	a, err = NewAgent(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}

	list, err := a.GetPhrases()
	if err != nil {
		log.Fatal(err)
	}

	if len(list) != 25 {
		defaultlist := getDefaultList()
		a.log("Updating Default List")
		for _, v := range defaultlist {
			err := a.UpdateMasterPhrase(v)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	a.log("Intializing admin email.")
	player := Player{"", "notrealemail"}
	err = a.AddAdmin(player)
	if err != nil {
		log.Fatal(err)
	}

}

// NewAgent intializes and returns a fresh agent.
func NewAgent(ctx context.Context, projectID string) (Agent, error) {
	var err error
	rand.Seed(time.Now().UTC().UnixNano())
	a := Agent{}
	a.ProjectID = projectID
	a.ctx = ctx
	a.client, err = firestore.NewClient(a.ctx, a.ProjectID)
	if err != nil {
		return a, fmt.Errorf("Failed to create client: %v", err)
	}
	return a, nil
}

type Agent struct {
	ProjectID string
	ctx       context.Context
	client    *firestore.Client
}

func (a *Agent) log(msg string) {
	log.Printf("Firestore : %s\n", msg)
}

// Phrase represents a statement, event or other such thing that we are on the
// lookout for in this game of bingo.
type Phrase struct {
	ID           string `json:"id" firestore:"id"`
	Text         string `json:"text" firestore:"text"`
	Selected     bool   `json:"selected" firestore:"selected"`
	Row          string `json:"row" firestore:"row"`
	Column       string `json:"column" firestore:"column"`
	DisplayOrder int    `json:"displayorder" firestore:"displayorder"`
}

// Player is a human user who is playing the game.
type Player struct {
	Name  string `json:"name"  firestore:"name"`
	Email string `json:"email"  firestore:"email"`
}

// GetPhrases fetches the master list of Phrases for populating Games
func (a *Agent) GetPhrases() ([]Phrase, error) {

	p := []Phrase{}

	a.log("Getting Phrases")
	iter := a.client.Collection("phrases").Documents(a.ctx)
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
		phrase.Text = dataMap["text"].(string)

		p = append(p, phrase)
	}

	return p, nil
}

func getProjectID() (string, error) {
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return "", fmt.Errorf("could not determine this project id: %v", err)
	}
	return credentials.ProjectID, nil
}

func getDefaultList() []Phrase {
	a.log("Getting the default phrase list")
	phrases := []Phrase{
		{"101", "Greg tells a dad joke", false, "", "", 0},
		{"102", "Greg references airplanes/piloting", false, "", "", 0},
		{"103", "\"We’re all in this together\"", false, "", "", 0},
		{"104", "\"the new ladder\"", false, "", "", 0},
		{"105", "Someone's child/S.O. on screen", false, "", "", 0},
		{"106", "\"OKRs\"", false, "", "", 0},
		{"107", "\"Increased (better, clearer) focus\"", false, "", "", 0},
		{"108", "\"These uncertain times\"", false, "", "", 0},
		{"109", "Someone’s pet on screen", false, "", "", 0},
		{"110", "\"working from home\"", false, "", "", 0},
		{"111", "Someone speaks when muted", false, "", "", 0},
		{"112", "\"Wash your hands\"", false, "", "", 0},
		{"113", "FREE", false, "", "", 0},
		{"114", "Awkward silence", false, "", "", 0},
		{"115", "Sports metaphor", false, "", "", 0},
		{"116", "Start at least 5 min late", false, "", "", 0},
		{"117", "Joke made, but no one laughs", false, "", "", 0},
		{"118", "Someone eats on screen", false, "", "", 0},
		{"119", "Answer all Dory questions", false, "", "", 0},
		{"120", "\"self care\"", false, "", "", 0},
		{"121", "\"Can you see my screen?\"", false, "", "", 0},
		{"122", "\"headcount\"", false, "", "", 0},
		{"123", "\"Sundar\"", false, "", "", 0},
		{"124", "\"TK\"", false, "", "", 0},
		{"125", "\"Amr\"", false, "", "", 0},
	}

	return phrases
}

// UpdateMasterPhrase updates a phrase in the master collection of phrases
func (a *Agent) UpdateMasterPhrase(phrase Phrase) error {
	a.log("Adding Phrase " + phrase.Text)
	if _, err := a.client.Collection("phrases").Doc(phrase.ID).Set(a.ctx, phrase); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

// AddAdmin adds an admin to the over all system
func (a *Agent) AddAdmin(player Player) error {
	if _, err := a.client.Collection("admins").Doc(player.Email).Set(ctx, player); err != nil {
		return fmt.Errorf("unable to add admin: %s", err)
	}
	return nil
}
