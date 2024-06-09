package tgbot

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

func NewBot(cfg *Config) (*tb.Bot, error) {
	b, err := tb.NewBot(tb.Settings{
		Token:   cfg.Token,
		Verbose: cfg.Verbose,
	})
	if err != nil {
		return nil, fmt.Errorf("new bot: %w", err)
	}

	err = b.SetWebhook(&tb.Webhook{
		Endpoint: &tb.WebhookEndpoint{
			PublicURL: cfg.WebhookUrl,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("set webhook: %w", err)
	}

	http.HandleFunc("/bot", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Received webhook request")
		if r.Method != http.MethodPost {
			log.Warn("Invalid request method")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		update := &tb.Update{}
		if err := json.NewDecoder(r.Body).Decode(update); err != nil {
			log.Errorf("Failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Infof("Processing update: %+v", update)
		b.ProcessUpdate(*update)
		w.WriteHeader(http.StatusOK)
	})

	log.Infof("Successfuly connect to tg api and use bot with username %q", b.Me.Username)

	return b, nil
}
