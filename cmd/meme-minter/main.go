package main

import (
	"fmt"
	"github.com/kekzploit/MeMeMinter/pkg/checks"
	"github.com/kekzploit/MeMeMinter/pkg/openai"
	price2 "github.com/kekzploit/MeMeMinter/pkg/price"
	"github.com/kekzploit/MeMeMinter/pkg/start"
	"github.com/kekzploit/MeMeMinter/pkg/transactions"
	"github.com/spf13/viper"
	"log"
	"strings"
	"time"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

func ServeBot(tgToken string, mongoUri string, walletApi string, coingeckoUrl string, openaiApi string, openaiUrl string, beamCharge int, beamTxFee int, primaryAddress string) {

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

		msg := fmt.Sprintf("Welcome %s\n\nUse /start command to register with MeMe Minter and start changing the world, one MeMe at a time!", user.FirstName)
		return c.Send(msg)
	})

	b.Handle("/address", func(c tele.Context) error {

		var (
			user = c.Sender()
		)

		userRegistered := checks.Registered(user.ID, mongoUri)
		if userRegistered {

			address, info, failMsg := transactions.GetAddress(user.ID, mongoUri)
			if address {
				msg := fmt.Sprintf("Hi %s,\n\nYour address is:\n\n%s", user.FirstName, info)
				return c.Send(msg)
			}
			failedMsg := fmt.Sprintf("hey %s,\n\n%s to get address", failMsg, user.FirstName)
			return c.Send(failedMsg)
		}
		msg := fmt.Sprintf("Oops %s\n\nIt looks like you're not registered, to register, type the /start command", user.FirstName)
		return c.Send(msg)

	})

	b.Handle("/balance", func(c tele.Context) error {

		var (
			user = c.Sender()
		)

		userRegistered := checks.Registered(user.ID, mongoUri)
		if userRegistered {
			balance, info, _ := transactions.BalanceCheck(user.ID, mongoUri)
			if balance {
				msg := fmt.Sprintf("Hi %s,\n\nYour balance is:\n\n%d GROTH", user.FirstName, info)
				return c.Send(msg)
			}

			failedMsg := fmt.Sprintf("%s to get balance", info)
			return c.Send(failedMsg)
		}
		msg := fmt.Sprintf("Oops %s\n\nIt looks like you're not registered, to register, type the /start command", user.FirstName)
		return c.Send(msg)

	})

	b.Handle("/meme", func(c tele.Context) error {

		var (
			user    = c.Sender()
			payload = strings.TrimSpace(c.Text()[6:])
		)

		userRegistered := checks.Registered(user.ID, mongoUri)
		if userRegistered {
			balanceEnough := checks.Balance(user.ID, mongoUri, beamCharge, beamTxFee)
			if balanceEnough {
				imageCreated, imageUrl := openai.CreateImage(openaiApi, openaiUrl, payload)
				if imageCreated {
					// DEDUCT BALANCE AND SEND BEAM TO EXTERNAL WALLET
					go transactions.DeductAndSendTx(user.ID, mongoUri, beamCharge, beamTxFee, walletApi, primaryAddress)
					//
					imageHtml := fmt.Sprintf("<a href=\"%s\">%s:\n%s</a>", imageUrl, user.FirstName, payload)
					return c.Send(imageHtml, &tele.SendOptions{ParseMode: tele.ModeHTML})
				}
				msg := fmt.Sprintf("Unable to create Image")
				return c.Send(msg)
			}
			msg := fmt.Sprintf("You have insufficient funds!, top up your war chest!")
			return c.Send(msg)
		} else {
			msg := fmt.Sprintf("Oops %s\n\nIt looks like you're not registered, to register, type the /start command", user.FirstName)
			return c.Send(msg)
		}

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

	b.Handle("/withdraw", func(c tele.Context) error {

		var (
			user = c.Sender()
		)

		userRegistered := checks.Registered(user.ID, mongoUri)
		if userRegistered {
			msg := fmt.Sprintf("Uhhm.. %s,\n\nOne does not simply withdraw from MeMeMinter", user.FirstName)
			return c.Send(msg)
		}
		msg := fmt.Sprintf("Oops %s\n\nIt looks like you're not registered, to register, type the /start command", user.FirstName)
		return c.Send(msg)
	})

	b.Handle("/usage", func(c tele.Context) error {

		var (
			user = c.Sender()
		)

		msg := fmt.Sprintf("Hi %s\n\nSee below for MeMe Minters command list:\n\n/address - gets your deposit address\n/balance - gets your current balance\n/meme <description of image you want to create>\n/rate -  current BEAM price\n/start - register with MeMe Minter\n/withdraw - one simply doesnt withdraw from MeMe Minter\n/usage - get list of MeMe Minters commands", user.FirstName)
		return c.Send(msg)
	})

	b.Start()
}

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
	coingeckoUrl := viper.Get("COINGECKO.APIURL").(string)
	openaiApi := viper.Get("OPENAI.APIKEY").(string)
	openaiUrl := viper.Get("OPENAI.URL").(string)
	beamCharge := viper.Get("BEAM.CHARGE").(int)
	beamTxFee := viper.Get("BEAM.TXFEE").(int)
	primaryAddress := viper.Get("BEAM.PRIMARYADDRESS").(string)

	go ServeBot(tgToken, mongoUri, walletApi, coingeckoUrl, openaiApi, openaiUrl, beamCharge, beamTxFee, primaryAddress)
	transactions.MonitorTx(walletApi, mongoUri)
}
