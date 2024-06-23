package routes

import (
	"context"
	"fmt"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	log "github.com/sirupsen/logrus"
	tb "gopkg.in/telebot.v3"
)

type ObsidianUsecase interface {
	ParseMessage(ctx context.Context, msg string) (string, error)
	CreateNewNoteToInbox(ctx context.Context, msg string) (string, error)
	AddAction(ctx context.Context, msg string) (string, error)
	GetWishList(ctx context.Context, msg string) (string, error)
	GetShoppingList(ctx context.Context, msg string) (string, error)
	AddItemsToShoppingList(ctx context.Context, msg string) (string, error)
	ClearShoppingList(ctx context.Context, msg string) (string, error)
	RemoveItemsFromShoppingList(ctx context.Context, msg string) (string, error)
	RememberAboutInbox(ctx context.Context) string
}

type bot struct {
	ObsidianUsecase ObsidianUsecase
	UserID          int64
}

func NewBot(obsidianUsecase ObsidianUsecase, userID int64) *bot {
	return &bot{
		ObsidianUsecase: obsidianUsecase,
		UserID:          userID,
	}
}

func (br *bot) TextMessageHandler(ctx context.Context, b *tb.Bot) {
	b.Handle(tb.OnText, func(c tb.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		user := c.Sender()

		if user == nil {
			return fmt.Errorf("nil sender for updateID=%d", c.Update().ID)
		}

		log.Infof("receive tg message: updateID=%d userID=%d username=%s text=%s",
			c.Update().ID, user.ID, user.Username, c.Text())

		var userFriendlyMessage string
		var err error
		if br.checkUser(user.ID) {
			userFriendlyMessage, err = br.ObsidianUsecase.ParseMessage(ctx, c.Text())
			if err != nil {
				log.Errorf("Message proccess error from handler: %v", err)
				userFriendlyMessage = fmt.Errorf("**Error occurred in proccessing message.**\n\n%w", err).Error()
			}
		} else {
			userFriendlyMessage = fmt.Sprintf("**You are not allowed to use this bot.**")
		}

		_, err = b.Send(c.Sender(), userFriendlyMessage, tb.ModeMarkdown)
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		return nil
	})
}

type Command struct {
	DescRu  string
	DescEn  string
	Handler func(ctx context.Context, msg string) (string, error)
}

func (br *bot) SetMenu(ctx context.Context, bot *tb.Bot, menu map[string]Command) error {
	cmdsRu := make([]tb.Command, 0, len(menu))
	cmdsEn := make([]tb.Command, 0, len(menu))

	keys := maps.Keys(menu)
	slices.Sort(keys)

	for _, key := range keys {
		cmd, info := key, menu[key]

		cmdsRu = append(cmdsRu, tb.Command{Text: cmd, Description: info.DescRu})
		cmdsEn = append(cmdsEn, tb.Command{Text: cmd, Description: info.DescEn})

		// Ошибки от этого хендлера логируются через bot.OnError.
		bot.Handle("/"+cmd, func(c tb.Context) error {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			user := c.Sender()

			if user == nil {
				return fmt.Errorf("nil sender for updateID=%d", c.Update().ID)
			}

			log.Infof("receive tg command: /%s updateID=%d userID=%d username=%s text=%s",
				cmd, c.Update().ID, user.ID, user.Username, c.Text())

			var userFriendlyMessage string
			var err error
			if br.checkUser(user.ID) {
				userFriendlyMessage, err = info.Handler(ctx, c.Text())
				if err != nil {
					log.Errorf("command %q get error from handler: %v", cmd, err)
					userFriendlyMessage = fmt.Errorf("**Error occurred in command %q.**\n\n%w", cmd, err).Error()
				}
			} else {
				userFriendlyMessage = fmt.Sprintf("**You are not allowed to use this bot.**")
			}

			_, err = bot.Send(c.Sender(), userFriendlyMessage, tb.ModeMarkdown)
			if err != nil {
				return fmt.Errorf("send message: %w", err)
			}

			return nil
		})
	}

	err := bot.SetCommands(cmdsRu, "ru")
	if err != nil {
		return fmt.Errorf("set ru commands: %w", err)
	}

	err = bot.SetCommands(cmdsEn)
	if err != nil {
		return fmt.Errorf("set default commands: %w", err)
	}

	return nil
}

func (br *bot) NotifyUser(ctx context.Context, b *tb.Bot) error {
	msg := br.ObsidianUsecase.RememberAboutInbox(ctx)

	_, err := b.Send(&tb.User{ID: br.UserID}, msg, tb.ModeMarkdown)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (br *bot) checkUser(userID int64) bool {
	return br.UserID == userID
}
