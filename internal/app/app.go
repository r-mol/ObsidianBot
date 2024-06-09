// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/r-mol/ObsidianBot/internal/configs"
	"github.com/r-mol/ObsidianBot/internal/repository"
	"github.com/r-mol/ObsidianBot/internal/routes"
	"github.com/r-mol/ObsidianBot/internal/usecases"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"

	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"

	"github.com/r-mol/ObsidianBot/pkg/tgbot"
)

func Run(ctx context.Context, configPath string) error {
	config, err := configs.ParseConfig(configPath)
	if err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	b, err := tgbot.NewBot(config.TgBot)
	if err != nil {
		return fmt.Errorf("new telegram bot: %w", err)
	}

	// init repo
	repo := repository.New(config.Server.ObsidianAbsolutePath)

	// init usecases
	obsidianUsecase := usecases.NewObsidian(repo, config.Server.UserID)

	// init routes
	botRoute := routes.NewBot(obsidianUsecase, config.Server.UserID)

	// init tg menu
	tgMenu := map[string]routes.Command{
		"shopping_list": {
			DescRu:  "Показать весь список покупок",
			DescEn:  "Get shopping list",
			Handler: obsidianUsecase.GetShoppingList,
		},
		"clear_shopping_list": {
			DescRu:  "Очисть список покупок",
			DescEn:  "Clear shopping list",
			Handler: obsidianUsecase.ClearShoppingList,
		},
		"remove_item": {
			DescRu:  "Удалить из списка покупок",
			DescEn:  "Remove item from shopping list",
			Handler: obsidianUsecase.RemoveItemsFromShoppingList,
		},
		"wishing_list": {
			DescRu:  "Показать весь список желаний",
			DescEn:  "Get wishing list",
			Handler: obsidianUsecase.GetWishList,
		},
	}

	err = botRoute.SetMenu(ctx, b, tgMenu)
	if err != nil {
		return fmt.Errorf("set telegram menu: %w", err)
	}

	botRoute.TextMessageHandler(ctx, b)

	c := cron.New()
	cronExpr := "0 6 * * *" // Every Sunday and Wednesday at 6:00 AM

	// Schedule NotifyUser function
	_, err = c.AddFunc(cronExpr, func() {
		err := botRoute.NotifyUser(ctx, b)
		if err != nil {
			log.Printf("error notifying user: %v", err)
		} else {
			log.Println("Notification sent successfully")
		}
	})

	if err != nil {
		return fmt.Errorf("add func to cron job: %w", err)
	}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := http.ListenAndServe(":"+config.Server.Port, nil); err != http.ErrServerClosed {
			return err
		} else {
			log.Info("HTTP server stopped")
			return nil
		}
	})

	eg.Go(func() error {
		b.Start()
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		b.Stop()
		return nil
	})

	eg.Go(func() error {
		c.Start()
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		c.Stop()
		return nil
	})

	err = eg.Wait()
	if err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}

func GetApp() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start obsidian bot",
		Long:  `start obsidian bot with specified config file`,
		Run: func(cmd *cobra.Command, args []string) {
			err := Run(context.Background(), configPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
	
	cmd.Flags().StringVar(&configPath, "config", "configs/config.yaml", "path to config file")
	_ = cmd.MarkFlagRequired("config")

	return cmd
}
