package command

import (
	"errors"

	"github.com/go-telegram-bot-api/telegram-bot-api"
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
			manager.SendStatusInfo(msg.MessageID, int(msg.Chat.ID))
		case "delete":
			if msg.ReplyToMessage != nil && msg.ReplyToMessage.Photo != nil {
			photosID := *msg.ReplyToMessage.Photo
			manager.DeleteMedia(photosID[len(photosID)-1].FileID, msg.MessageID, int(msg.Chat.ID), msg.From.UserName)
			} else if msg.ReplyToMessage != nil && msg.ReplyToMessage.Video != nil {
			manager.DeleteMedia(msg.ReplyToMessage.Video.FileID, msg.MessageID, int(msg.Chat.ID), msg.From.UserName)
			} else {
				return nil
			}
		} 
	}

	return nil
}
