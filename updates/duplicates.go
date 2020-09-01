package updates

import (
	"fmt"
	"github.com/zelenin/go-tdlib/client"
	"gitlab.com/shitposting/autoposting-bot/api"
	"gitlab.com/shitposting/autoposting-bot/documentstore/entities"
	l "gitlab.com/shitposting/autoposting-bot/localization"
	"gitlab.com/shitposting/autoposting-bot/repository"
	"gitlab.com/shitposting/autoposting-bot/telegram"
	"gitlab.com/shitposting/autoposting-bot/utility"
	"strconv"
)

const(
	telegramMessageIDConversionFactor = 1048576
)

func getDuplicateCaption(duplicatePost *entities.Post) (*client.FormattedText, error) {

	var userName string
	user, err := api.GetUserByID(duplicatePost.AddedBy)
	if err != nil {
		userName = strconv.Itoa(int(duplicatePost.AddedBy))
	} else {
		userName = telegram.GetNameFromUser(user)
	}

	caption := fmt.Sprintf(
		l.GetString(l.UPDATES_DUPLICATES_DUPLICATE_ADDED_BY),
		duplicatePost.AddedBy, userName, utility.FormatDate(duplicatePost.AddedAt))

	if duplicatePost.MessageID != 0 {

		// In order to work well with private channels, we need to use the
		// t.me/c/chatid/messageid format
		// We need a few changes, though.
		// For channels we need to substring the chatID from the 4th position,
		// effectively removing the prefix -100
		// We then need to convert the messageID from tdlib to normal telegram
		// Since bots cannot call the getLink method, we need to divide
		// our message id by the magic number and add 1
		chatIDStr := strconv.FormatInt(repository.Config.Autoposting.ChannelID, 10)
		link := fmt.Sprintf("t.me/c/%s/%d", chatIDStr[4:], duplicatePost.MessageID/telegramMessageIDConversionFactor + 1)
		captionEnd := fmt.Sprintf(l.GetString(l.UPDATES_DUPLICATE_DUPLICATE_ADDED_AT), utility.FormatDate(*duplicatePost.PostedAt), link)
		caption = fmt.Sprintf("%s\n%s", caption, captionEnd)

	}

	ft, err := api.GetFormattedText(caption)
	return ft, err

}
