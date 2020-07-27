package commands

import (
	"fmt"
	"github.com/zelenin/go-tdlib/client"
	"gitlab.com/shitposting/autoposting-bot/api"
	"gitlab.com/shitposting/autoposting-bot/documentstore/dbwrapper"
	"gitlab.com/shitposting/autoposting-bot/posting"
	"time"
)

const (
	//TODO: ASSOLUTAMENTE NON QUA, POSSIBILMENTE STRINGA TRADOTTA
	statusText = "📋 Posts enqueued: %d\n🕜 Post rate: %s\n\n🔮 Next post in: %s (%s)"
)

type StatusCommandHandler struct {}

func (StatusCommandHandler) Handle(arguments string, message, replyToMessage *client.Message) error {

	//
	nextPost := posting.GetNextPostTime()
	queueLength := dbwrapper.GetQueueLength()
	postingRate := posting.GetPostingRate().String()
	minutesUntilNextPost := time.Until(nextPost).Truncate(time.Minute)

	//
	text := fmt.Sprintf(statusText,
		queueLength,
		postingRate,
		minutesUntilNextPost,
		nextPost.Format("15:04"))

	//
	_, err := api.SendPlainText(message.ChatId, text)
	return err

}
