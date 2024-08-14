package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

type Author struct {
	Npub string
	Name string
	Link string
}

type PostedNotes struct {
	NoteIDs map[string]bool `json:"note_ids"`
	mu      sync.Mutex
}

func loadAuthors(filename string) ([]Author, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	var authors []Author
	if err := json.Unmarshal(byteValue, &authors); err != nil {
		return nil, err
	}

	return authors, nil
}

func loadPostedNotes(filename string) (*PostedNotes, error) {
	file, err := os.Open(filename)
	if os.IsNotExist(err) {
		// If the file doesn't exist, return an empty PostedNotes struct
		return &PostedNotes{NoteIDs: make(map[string]bool)}, nil
	} else if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	var postedNotes PostedNotes
	if err := json.Unmarshal(byteValue, &postedNotes); err != nil {
		return nil, err
	}

	return &postedNotes, nil
}

func savePostedNotes(filename string, postedNotes *PostedNotes) error {
	postedNotes.mu.Lock()
	defer postedNotes.mu.Unlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := json.Marshal(postedNotes)
	if err != nil {
		return err
	}

	_, err = file.Write(byteValue)
	return err
}

func postToSlack(message string) error {
	webhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if webhookURL == "" {
		return fmt.Errorf("SLACK_WEBHOOK_URL environment variable is not set")
	}

	payload := map[string]string{"text": message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status code %d", resp.StatusCode)
	}

	return nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Load the posted notes from the JSON file
	postedNotes, err := loadPostedNotes("posted_notes.json")
	if err != nil {
		log.Fatalf("Failed to load posted notes: %v", err)
	}

	ctx := context.Background()
	relayUrl := os.Getenv("RELAY_URL")
	relay, err := nostr.RelayConnect(ctx, relayUrl)
	if err != nil {
		panic(err)
	}

	authors, err := loadAuthors("authors.json")
	if err != nil {
		log.Fatalf("Failed to load authors: %v", err)
	}

	var authorsPubKeys []string
	authorMap := make(map[string]Author)

	for _, author := range authors {
		if _, v, err := nip19.Decode(author.Npub); err == nil {
			pub := v.(string)
			authorsPubKeys = append(authorsPubKeys, pub)
			authorMap[pub] = author
		} else {
			panic(err)
		}
	}

	filters := []nostr.Filter{{
		Kinds:   []int{nostr.KindTextNote},
		Authors: authorsPubKeys,
		Limit:   1,
	}}

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		panic(err)
	}

	for ev := range sub.Events {
		postedNotes.mu.Lock()
		if postedNotes.NoteIDs[ev.ID] {
			postedNotes.mu.Unlock()
			continue // Skip if already posted
		}
		postedNotes.mu.Unlock()

		author := authorMap[ev.PubKey]
		njumpLink := fmt.Sprintf("https://njump.me/%s", ev.ID)
		content := strings.ReplaceAll(ev.Content, "\n", "\n>")
		slackMessage := fmt.Sprintf("*<%s|%s>* - <%s|njump>\n\n>%s", author.Link, author.Name, njumpLink, content)

		if err := postToSlack(slackMessage); err != nil {
			log.Printf("Failed to post to Slack: %v", err)
			continue
		}

		// Mark the note as posted
		postedNotes.mu.Lock()
		postedNotes.NoteIDs[ev.ID] = true
		postedNotes.mu.Unlock()

		// Save the updated note IDs to the JSON file
		if err := savePostedNotes("posted_notes.json", postedNotes); err != nil {
			log.Printf("Failed to save posted notes: %v", err)
		}
	}
}
