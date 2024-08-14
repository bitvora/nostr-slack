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

	"github.com/joho/godotenv"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

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

type Author struct {
	Npub string
	Name string
	Link string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
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
		Limit:   3,
	}}

	sub, err := relay.Subscribe(ctx, filters)
	if err != nil {
		panic(err)
	}

	for ev := range sub.Events {
		author := authorMap[ev.PubKey]
		njumpLink := fmt.Sprintf("https://njump.me/%s", ev.ID)
		content := strings.ReplaceAll(ev.Content, "\n", "\n>")
		slackMessage := fmt.Sprintf("*<%s|%s>* - <%s|njump>\n\n>%s", author.Link, author.Name, njumpLink, content)
		if err := postToSlack(slackMessage); err != nil {
			panic(err)
		}
	}
}
