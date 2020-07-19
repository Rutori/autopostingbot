package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gitlab.com/shitposting/autoposting-bot/analysisadapter"
	"gitlab.com/shitposting/autoposting-bot/api"
	"gitlab.com/shitposting/autoposting-bot/documentstore"
	updates2 "gitlab.com/shitposting/autoposting-bot/updates"
	"net/http"

	"gitlab.com/shitposting/autoposting-bot/edition"

	"gitlab.com/shitposting/autoposting-bot/manager"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

	configuration "gitlab.com/shitposting/autoposting-bot/config"
	"gitlab.com/shitposting/autoposting-bot/messages"
	"gitlab.com/shitposting/autoposting-bot/repository"
)

var (
	// config file path, if not specified it will read ./config.toml
	configFilePath string

	// Version is the autoposting-bot version, a compile-time value
	Version string

	// Build is the git tag for the current version
	Build string

	// testing is a bool value to enable polling instead of webhook
	testing bool

	//debug --
	debug bool

	//polling
	polling bool

	//sushiEdition
	sushiEdition bool
)

func main() {

	/* LOAD CLI ARGUMENTS */
	loadCLIParams()

	/* LOAD CONFIGURATION */
	cfg, err := configuration.Load(configFilePath, !polling)
	if err != nil {
		log.Fatal(err)
	}

	tdlibClient, err := api.Authorize(cfg.BotToken, &cfg.Tdlib)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}
	repository.Tdlib = tdlibClient
	repository.Me, err = tdlibClient.GetMe()
	if err != nil {
		log.Fatal("GetMe")
	}
	listener := tdlibClient.GetListener()
	defer listener.Close()

	analysisadapter.Start(cfg.AnalysisAPI)

	go updates2.HandleUpdates(listener)

	/* INITIALIZE BOT */
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Error(err.Error())
		return
	}

	/* SET EDITION */
	if sushiEdition {
		edition.SetEdition(edition.Sushiporn)
	} else {
		edition.SetEdition(edition.Shitpost)
	}

	/* PRINT INFO */
	bot.Debug = debug
	log.Info(fmt.Sprintf("Shitposting autoposting-bot version v%s, build %s, %s", Version, Build, edition.GetEditionString()))
	log.Info(fmt.Sprintf("Authorized on account @%s", bot.Self.UserName))

	/* CONNECT TO THE DATABASE */
	db, err := gorm.Open("mysql", cfg.DB.DatabaseConnectionString())
	if err != nil {
		log.Error(err.Error())
		return
	}

	documentstore.Connect(&cfg.DocumentStore)

	/* CREATE Repository */
	repo := repository.SetVariables(bot, db, &cfg)

	/* GET UPDATES CHANNEL */
	updates := getUpdatesChannel(repo)
	if updates == nil {
		log.Error("Update channel nil")
		return
	}

	err = manager.StartManager(repo.Bot, repo.Db, repo.Config, debug, testing)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to start manager: %s", err.Error()))
		return
	}

	/* HANDLE UPDATES */
	handleUpdates(updates, repo)
}

//handleUpdates iterates on the updates and passes them onto the handlers
func handleUpdates(updates tgbotapi.UpdatesChannel, repo *repository.Repository) {
	for update := range updates {
		switch {
		case update.EditedMessage != nil:
			go messages.HandleEdited(update.EditedMessage, repo)
		case update.Message != nil:
			go messages.HandleNew(update.Message, repo)
		}
	}
}

func loadCLIParams() {
	flag.StringVar(&configFilePath, "config", "./config.toml", "configuration file path")
	flag.BoolVar(&testing, "testing", false, "activate testing features")
	flag.BoolVar(&debug, "debug", false, "activate all the debug features")
	flag.BoolVar(&polling, "polling", false, "use polling instead of webhoooks")
	flag.BoolVar(&sushiEdition, "sushi", false, "set the current edition as sushiporn")
	flag.Parse()
}

//getUpdatesChannel uses webhooks or polling to get an `UpdatesChannel`
func getUpdatesChannel(repo *repository.Repository) tgbotapi.UpdatesChannel {

	/* WEBHOOKS IF WE'RE NOT TESTING */
	if !polling {
		return useWebhook(repo)
	}

	/* POLLING OTHERWISE */
	_, err := repo.Bot.Request(tgbotapi.RemoveWebhookConfig{})
	if err != nil {
		log.Error(fmt.Sprintf("Unable to remove webhook: %s", err.Error()))
		return nil
	}

	return usePolling(repo)
}

//usePolling gets an `UpdatesChannel` using polling
func usePolling(repo *repository.Repository) tgbotapi.UpdatesChannel {

	updateConfig := tgbotapi.UpdateConfig{
		Offset:  0,
		Limit:   0,
		Timeout: 60,
	}

	return repo.Bot.GetUpdatesChan(updateConfig)
}

//useWebhook ets an `UpdatesChannel` using webhooks
func useWebhook(repo *repository.Repository) tgbotapi.UpdatesChannel {

	go startServer(repo.Config.Server)

	/* TRY TO RETRIEVE WEBHOOK INFORMATION FROM TELEGRAM */
	webhook, err := repo.Bot.GetWebhookInfo()

	/* SET UP NEW WEBHOOK */
	if err != nil || !webhook.IsSet() {
		newWebhook := tgbotapi.NewWebhook(repo.Config.WebHookURL())
		webhookConfig := tgbotapi.WebhookConfig{
			URL:            newWebhook.URL,
			Certificate:    newWebhook.Certificate,
			MaxConnections: newWebhook.MaxConnections,
			AllowedUpdates: newWebhook.AllowedUpdates,
		}

		_, err := repo.Bot.Request(webhookConfig)
		if err != nil {
			log.Error(fmt.Sprintf("Unable to request webhookConfig: %s", err.Error()))
			return nil
		}
	}

	return repo.Bot.ListenForWebhook(repo.Config.WebHookPath())
}

//startServer starts serving HTTP requests with or without TLS
func startServer(config configuration.ServerDetails) {
	if config.TLS {
		log.Error((http.ListenAndServeTLS(config.BindString(), config.TLSCertPath, config.TLSKeyPath, nil)).Error())
	} else {
		log.Error((http.ListenAndServe(config.BindString(), nil)).Error())
	}
}
