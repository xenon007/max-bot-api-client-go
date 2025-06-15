package maxbot

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/xenon007/max-bot-api-client-go/schemes"
)

type debugs struct {
	client *client
	chat   int64
}

func newDebugs(client *client, chat int64) *debugs {
	return &debugs{client: client, chat: chat}
}

// Send sends a message to a chat. As a result for this method new message identifier returns.
func (a *debugs) Send(upd schemes.UpdateInterface) (string, error) {
	return a.sendMessage(false, false, a.chat, 0, &schemes.NewMessageBody{Text: upd.GetDebugRaw()})
}

// Send sends a message to a chat. As a result for this method new message identifier returns.
func (a *debugs) SendErr(err error) (string, error) {
	return a.sendMessage(false, false, a.chat, 0, &schemes.NewMessageBody{Text: err.Error()})
}

func (a *debugs) sendMessage(vip bool, reset bool, chatID int64, userID int64, message *schemes.NewMessageBody) (string, error) {
	result := new(schemes.Error)
	values := url.Values{}
	if chatID != 0 {
		values.Set("chat_id", strconv.Itoa(int(chatID)))
	}
	if userID != 0 {
		values.Set("user_id", strconv.Itoa(int(userID)))
	}
	if reset {
		values.Set("access_token", message.BotToken)
	}
	mode := "messages"
	if vip {
		mode = "notify"
	}
	body, err := a.client.request(http.MethodPost, mode, values, reset, message)
	if err != nil {
		return "heir", err
	}
	defer body.Close()
	if err := json.NewDecoder(body).Decode(result); err != nil {
		// Message sent without errors
		return "err", err
	}
	if result.Code == "" {
		if mode == "notify" {
			return "ok", result
		} else {
			return "", nil
		}

	}
	return "", result
}
