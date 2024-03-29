package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func getAllTricks() (string, error) {
	req, err := http.NewRequest("GET", "/tricks", nil)
	if err != nil {
		return "", err
	}

	rr := httptest.NewRecorder()

	handleTricks(rr, req)

	if rr.Code != http.StatusOK {
		return "", fmt.Errorf("failed to retrieve tricks, status code: %d", rr.Code)
	}

	var tricks []Trick
	if err := json.NewDecoder(rr.Body).Decode(&tricks); err != nil {
		return "", err
	}

	var tricksText strings.Builder
	for _, trick := range tricks {
		tricksText.WriteString("Trick Name: ")
		tricksText.WriteString(trick.Name)
		tricksText.WriteString("\n")
		tricksText.WriteString("Trick Description: ")
		tricksText.WriteString(trick.Description)
		tricksText.WriteString("\n")
		tricksText.WriteString("Difficulty: ")
		tricksText.WriteString(trick.Difficulty)
		tricksText.WriteString("\n")
		tricksText.WriteString("Progress: ")
		tricksText.WriteString(trick.Progress)
		tricksText.WriteString("\n")
		tricksText.WriteString("\n")
	}

	return tricksText.String(), nil
}

func InsertNewTrick(newTrick string) (string, error) {
	cleanString := strings.Replace(newTrick, "/NewTrick", "", -1)
	splittedTrick := strings.Split(cleanString, ",")
	fmt.Println("InsertNewTrick debug splitted trick string: ", splittedTrick)

	newTrickJSON, err := json.Marshal(splittedTrick)
	if err != nil {
		return "", err
	}

	resp, err := http.Post("http://localhost:8080/tricks", "application/json", bytes.NewBuffer(newTrickJSON))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var insertedTrick Trick
	if err := json.NewDecoder(resp.Body).Decode(&insertedTrick); err != nil {
		return "", err
	}

	var trickText strings.Builder
	trickText.WriteString(insertedTrick.Name)
	trickText.WriteString(insertedTrick.Description)
	trickText.WriteString(insertedTrick.Difficulty)
	trickText.WriteString(insertedTrick.Progress)

	return trickText.String(), nil
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./tricks.db")
	if err != nil {
		panic(err)
	}
	createTable()
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("GARYS_TRICKS_TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			switch update.Message.Command() {
			case "Tricks":
				msg.Text, err = getAllTricks()
				if err != nil {
					msg.Text = "Failed to retrieve tricks"
				}
			case "NewTrick":
				msg.Text, err = InsertNewTrick(update.Message.Text)
				if err != nil {
					msg.Text = "Failed to insert new trick"
				}
			case "TagebuchUebersicht":
				msg.Text = "Hier kommen Garys TagebuchEinträge"
			case "TagebuchEintrag":
				msg.Text = "Hier kannst du einen Eintrag hinzufügen"
			default:
				msg.Text = "I don't know that command"
			}
			bot.Send(msg)
		}
	}
}
