package commands

import (
	"errors"
	"github.com/shitpostingio/autopostingbot/analysisadapter"
	"github.com/shitpostingio/autopostingbot/api"
	"github.com/shitpostingio/autopostingbot/caption"
	"github.com/shitpostingio/autopostingbot/documentstore/dbwrapper"
	"github.com/shitpostingio/autopostingbot/documentstore/entities"
	l "github.com/shitpostingio/autopostingbot/localization"
	"github.com/shitpostingio/autopostingbot/posting"
	"github.com/zelenin/go-tdlib/client"
)

// AddCommandHandler represents the handler of the /add command.
type AddCommandHandler struct{}

// Handle handles the /add command.
// /add forces the addition of a post to the database, skipping fingerprinting checks.
func (AddCommandHandler) Handle(_ string, message, replyToMessage *client.Message) error {

	//
	if replyToMessage == nil {
		_, _ = api.SendPlainReplyText(message.ChatId, message.Id, l.GetString(l.COMMANDS_REPLY_TO_MEDIA_FILE))
		return errors.New("reply to message nil")
	}

	//
	fileInfo, err := api.GetMediaFileInfo(replyToMessage)
	if err != nil {
		_, _ = api.SendPlainReplyText(message.ChatId, message.Id, l.GetString(l.COMMANDS_REPLY_TO_MEDIA_FILE))
		return err
	}

	//
	mediaType := api.GetTypeFromMessageType(replyToMessage.Content.MessageContentType())
	fingerprint, err := analysisadapter.Request(fileInfo.Local.Path, mediaType, fileInfo.Remote.UniqueId)
	if err != nil {
		_, _ = api.SendPlainReplyText(message.ChatId, message.Id, l.GetString(l.ANALYSIS_NO_MEDIA_FINGERPRINT))
		return err
	}

	//
	avg, sum := entities.GetHistogramAverageAndSum(fingerprint.Histogram)

	//
	media := entities.Media{
		Type:             mediaType,
		TdlibID:          fileInfo.Id,
		FileUniqueID:     fileInfo.Remote.UniqueId,
		FileID:           fileInfo.Remote.Id,
		Histogram:        fingerprint.Histogram,
		HistogramAverage: avg,
		HistogramSum:     sum,
		PHash:            fingerprint.PHash,
	}

	// If the message is a forward, remove the caption
	postCaption := ""

	// if the message is not a forward, get the caption
	if message.ForwardInfo == nil {
		ft := api.GetMediaFormattedText(replyToMessage)
		postCaption = caption.ToHTMLCaption(ft)
	}

	senderUserID, err := api.GetSenderUserID(replyToMessage)
	if err != nil {
		return err
	}

	err = dbwrapper.AddPost(senderUserID, media, postCaption)
	if err != nil {
		_, _ = api.SendPlainReplyText(message.ChatId, message.Id, l.GetString(l.COMMANDS_ADD_ERROR))
	} else {
		_, _ = api.SendPlainReplyText(message.ChatId, message.Id, l.GetString(l.MEDIA_ADDED_CORRECTLY))

		// If the queue was empty, re-schedule
		if dbwrapper.GetQueueLength() == 1 {
			posting.ForcePostScheduling()
		}

	}

	return err

}
