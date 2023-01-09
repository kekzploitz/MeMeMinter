package main

import (
	"fmt"
	"github.com/kekzploit/MeMeMinter/pkg/start"
	"log"
	"time"

	"github.com/spf13/viper"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func main() {
	// initiate environment variable config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../configs/")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	tgToken := viper.Get("TELEGRAM.APIKEY").(string)
	mongoUri := viper.Get("MONGODB.URI").(string)
	walletApi := viper.Get("BEAM.WALLETAPI").(string)

	pref := tele.Settings{
		Token:  tgToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Global-scoped middleware:
	b.Use(middleware.Logger())
	b.Use(middleware.AutoRespond())

	// Group-scoped middleware:
	adminOnly := b.Group()

	admins := []int64{5578680936} // list of telegram IDs
	adminOnly.Use(middleware.Whitelist(admins...))

	adminOnly.Handle("/totals", func(c tele.Context) error {

		var (
		// payload = c.Message().Payload
		// user = c.Sender()
		// text = c.Text()
		)

		msg := fmt.Sprintln("admin only")
		return c.Send(msg)
	})

	b.Handle(tele.OnUserJoined, func(c tele.Context) error {

		var (
			user = c.Sender()
		)

		msg := fmt.Sprintf("Welcome %s\n\nUse /start command to register with MeMe Minter", user.FirstName)
		return c.Send(msg)
	})

	b.Handle("/start", func(c tele.Context) error {

		var (
			// payload = c.Message().Payload
			user = c.Sender()
			// text = c.Text()
		)
		register := start.Start(user.FirstName, user.ID, mongoUri, walletApi)
		if register {
			msg := fmt.Sprintln("Registration successful")
			return c.Send(msg)
		}

		msg := fmt.Sprintln("Registration unsuccessful")
		return c.Send(msg)
	})

	b.Start()
}
