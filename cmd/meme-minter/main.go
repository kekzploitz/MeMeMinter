package main

import (
	"fmt"
	"github.com/kekzploit/MeMeMinter/pkg/checks"
	price2 "github.com/kekzploit/MeMeMinter/pkg/price"
	"github.com/kekzploit/MeMeMinter/pkg/start"
	"log"
	"time"

	"github.com/spf13/viper"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func ServeBot() {
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
	coingeckoUrl := viper.Get("COINGECKO.APIURL").(string)

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
	//b.Use(middleware.Logger())
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

	b.Handle("/rate", func(c tele.Context) error {

		var (
			user = c.Sender()
		)

		priceFetched, gbp, usd, eur := price2.GetPrice(coingeckoUrl)

		if priceFetched {
			msg := fmt.Sprintf("Hi %s,\n\nThe current price of BEAM:\n\nGBP: £%f\nUSD: $%f\nEUR: €%f\n\n Price data provided by coingecko.com", user.FirstName, gbp, usd, eur)
			return c.Send(msg)
		}
		msg := fmt.Sprintf("Hi %s,\n\nUnfortunately we've been unable to fetch price data right now, try again soon", user.FirstName)
		return c.Send(msg)
	})

	b.Handle("/start", func(c tele.Context) error {

		var (
			user = c.Sender()
		)

		userExists, _ := checks.CheckUserExists(user.ID)
		if !userExists {
			register, addressMsg := start.Start(user.FirstName, user.ID, mongoUri, walletApi)
			if register {
				msg := fmt.Sprintf("Welcome to MeMe Minter %s\n\nYour MeMe Minter BEAM address:%s\n\nUse the /usage command for instructions on how to use MeMe Minter", user.FirstName, addressMsg)
				return c.Send(msg)
			}

			msg := fmt.Sprintln("Registration unsuccessful")
			return c.Send(msg)
		}

		msg := fmt.Sprintf("Hi %s,\n\nYou're already registered with MeMe Minter, go change the world!", user.FirstName)
		return c.Send(msg)
	})

	b.Start()
}

func main() {
	ServeBot()
}
