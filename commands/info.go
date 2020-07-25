package commands

import (
	"fmt"
	"github.com/hako/durafmt"
	"github.com/zelenin/go-tdlib/client"
	"gitlab.com/shitposting/autoposting-bot/api"
	"gitlab.com/shitposting/autoposting-bot/documentstore/dbwrapper"
	"gitlab.com/shitposting/autoposting-bot/posting"
	"gitlab.com/shitposting/autoposting-bot/telegram"
	"gitlab.com/shitposting/autoposting-bot/utility"
	"strconv"
	"time"
)

type InfoCommandHandler struct {
}

//TODO: RIMUOVERE LE PRINT E TIRARE FUORI LE STRINGHE CABLATE
func (InfoCommandHandler) Handle(arguments string, message, replyToMessage *client.Message) error {

	fmt.Println("HANDLING INFO")
	fileInfo, err := api.GetMediaFileInfo(replyToMessage)
	if err != nil {
		return err
	}
	fmt.Println("UniqueID: ", fileInfo.Remote.UniqueId)

	post, err := dbwrapper.FindPostByUniqueID(fileInfo.Remote.UniqueId)
	if err != nil {
		return err
	}

	fmt.Println("Found post: ", post.AddedAt)

	var reply, name string

	user, err := api.GetUserByID(post.AddedBy)
	if err != nil {
		name = strconv.Itoa(int(post.AddedBy))
	} else {
		name = telegram.GetNameFromUser(user)
	}

	if post.PostedAt != nil {

		fmt.Println("POSTED")
		reply = fmt.Sprintf("Post added by <a href=\"tg://user?id=%d\">%s</a> on %s\nPosted on %s\nLink: t.me/%s/%d",
			post.AddedBy, name, utility.FormatDate(post.AddedAt), utility.FormatDate(*post.PostedAt), posting.GetPostingManager().GetEditionName(), post.MessageID)

		ft, err := api.GetFormattedText(reply)
		if err != nil {
			return err
		}

		_, err = api.SendText(replyToMessage.ChatId, replyToMessage.Id, ft.Text, ft.Entities)
		return err

	}

	fmt.Println("NOT POSTED")
	position := dbwrapper.GetQueuePositionByAddTime(post.AddedAt)
	fmt.Println("POSITION FOUND: ", position)
	timeToPost := posting.GetNextPostTime().Add(posting.GetPostingManager().EstimatePostTime(position - 1))
	durationUntilPost := durafmt.Parse(time.Until(timeToPost).Truncate(time.Minute))

	reply = fmt.Sprintf("📋 The post is number %d in the queue\n👤 Added by <a href=\"tg://user?id=%d\">%s</a> on %s\n\n🕜 It should be posted roughly in %s\n📅 On %s",
		position, post.AddedBy, name, utility.FormatDate(post.AddedAt), durationUntilPost.String(), utility.FormatDate(timeToPost))

	ft, err := api.GetFormattedText(reply)
	if err != nil {
		return err
	}

	_, err = api.SendText(replyToMessage.ChatId, replyToMessage.Id, ft.Text, ft.Entities)
	return err

}
