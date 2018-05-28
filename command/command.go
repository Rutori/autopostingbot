package command

import (
	"errors"
	"fmt"

	"gitlab.com/shitposting/autoposting-bot/utility"

	"github.com/empetrone/telegram-bot-api"
	"gitlab.com/shitposting/autoposting-bot/algo"
)

// Handle e` il punto di entrata per il parsing e l'organizzazione dell'azione del bot
// su un messaggio entrante.
func Handle(update tgbotapi.Update, api *tgbotapi.BotAPI, manager *algo.Manager) error {
	if update.Message == nil && update.EditedMessage == nil {
		return errors.New("update Message or EditedMessage body was nil, most likely an error on Telegram side")
	}

	msg := update.Message
	editedMsg := update.EditedMessage

	if editedMsg != nil {
		switch {
		case editedMsg.Video != nil:
			modifyMedia(editedMsg.Video.FileID, editedMsg.Caption, manager, editedMsg.From.ID, editedMsg.MessageID, int(editedMsg.Chat.ID))
		case editedMsg.Photo != nil:
			photos := *editedMsg.Photo
			modifyMedia(photos[len(photos)-1].FileID, editedMsg.Caption, manager, editedMsg.From.ID, editedMsg.MessageID, int(editedMsg.Chat.ID))
		case editedMsg.Text != "":
			switch editedMsg.Command() {
			case "caption":
				editCaption(editedMsg, api, manager, false)
			case "credit":
				editCaption(editedMsg, api, manager, true)
			}
		}

		return nil
	}

	switch {
	case msg.Video != nil:
		saveMedia(msg.Video.FileID, msg.Caption, Video, manager, msg.From.ID, msg.MessageID, int(msg.Chat.ID))
	case msg.Photo != nil:
		photos := *msg.Photo
		saveMedia(photos[len(photos)-1].FileID, msg.Caption, Image, manager, msg.From.ID, msg.MessageID, int(msg.Chat.ID))
	case msg.Text != "":
		switch msg.Command() {
		case "status":
			statusSignal(msg, manager)
		case "delete":
			deleteMedia(msg, api, manager)
		case "caption":
			editCaption(msg, api, manager, false)
		case "credit":
			editCaption(msg, api, manager, true)
		}

	}

	return nil
}

// editCaption allows the user to edit the caption of a forwarded message or give the credit to the user.
// It is used both by caption and credit command in the bot.
func editCaption(msg *tgbotapi.Message, api *tgbotapi.BotAPI, manager *algo.Manager, isCredit bool) {

	var newcaption string

	fileID, err := checkReplyAndMedia(msg)

	if err != nil {
		utility.SendTelegramReply(int(msg.Chat.ID), msg.MessageID, api, err.Error())
		return
	}

	if msg.ReplyToMessage.ForwardFrom != nil && isCredit == true {
		newcaption = fmt.Sprintf("%s\n\n[By %s]", msg.CommandArguments(), msg.ReplyToMessage.ForwardFrom.FirstName)
	} else {
		newcaption = msg.CommandArguments()
	}

	modifyMedia(fileID, newcaption, manager, msg.From.ID, msg.MessageID, int(msg.Chat.ID))
}
