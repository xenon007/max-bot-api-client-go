package maxbot

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/xenon007/max-bot-api-client-go/schemes"
)

type messages struct {
	client *client
}

func newMessages(client *client) *messages {
	return &messages{client: client}
}

// GetMessages returns messages in chat: result page and marker referencing to the next page. Messages traversed in reverse direction so the latest message in chat will be first in result array. Therefore if you use from and to parameters, to must be less than from
func (a *messages) GetMessages(chatID int64, messageIDs []string, from int, to int, count int) (*schemes.MessageList, error) {
	result := new(schemes.MessageList)
	values := url.Values{}
	if chatID != 0 {
		values.Set("chat_id", strconv.Itoa(int(chatID)))
	}
	if len(messageIDs) > 0 {
		for _, mid := range messageIDs {
			values.Add("message_ids", mid)
		}
	}
	if from != 0 {
		values.Set("from", strconv.Itoa(from))
	}
	if to != 0 {
		values.Set("to", strconv.Itoa(to))
	}
	if count > 0 {
		values.Set("count", strconv.Itoa(count))
	}
	body, err := a.client.request(http.MethodGet, "messages", values, false, nil)
	if err != nil {
		return result, err
	}
	defer func() {
		if err := body.Close(); err != nil {
			log.Println(err)
		}
	}()
	return result, json.NewDecoder(body).Decode(result)
}

// EditMessage updates message by id
func (a *messages) EditMessage(messageID int64, message *Message) error {
	s, err := a.editMessage(messageID, message.message)
	if err != nil {
		return err
	}
	if !s.Success {
		return errors.New(s.Message)
	}
	return nil
}

// DeleteMessage deletes message by id
func (a *messages) DeleteMessage(messageID int64) (*schemes.SimpleQueryResult, error) {
	result := new(schemes.SimpleQueryResult)
	values := url.Values{}
	values.Set("message_id", strconv.Itoa(int(messageID)))
	body, err := a.client.request(http.MethodDelete, "messages", values, false, nil)
	if err != nil {
		return result, err
	}
	defer func() {
		if err := body.Close(); err != nil {
			log.Println(err)
		}
	}()
	return result, json.NewDecoder(body).Decode(result)
}

// AnswerOnCallback should be called to send an answer after a user has clicked the button. The answer may be an updated message or/and a one-time user notification.
func (a *messages) AnswerOnCallback(callbackID string, callback *schemes.CallbackAnswer) (*schemes.SimpleQueryResult, error) {
	result := new(schemes.SimpleQueryResult)
	values := url.Values{}
	values.Set("callback_id", callbackID)
	body, err := a.client.request(http.MethodPost, "answers", values, false, callback)
	if err != nil {
		return result, err
	}
	defer func() {
		if err := body.Close(); err != nil {
			log.Println(err)
		}
	}()
	return result, json.NewDecoder(body).Decode(result)
}

// NewKeyboardBuilder returns new keyboard builder helper
func (a *messages) NewKeyboardBuilder() *Keyboard {
	return &Keyboard{
		rows: make([]*KeyboardRow, 0),
	}
}

// Send sends a message to a chat. As a result for this method new message identifier returns.
func (a *messages) Send(m *Message) (string, error) {
	return a.sendMessage(m.vip, m.reset, m.chatID, m.userID, m.message)
}

// Send sends a message to a chat. As a result for this method new message identifier returns.
func (a *messages) SendMessageResult(m *Message) (schemes.Message, error) {
	_, err := a.sendMessage(m.vip, m.reset, m.chatID, m.userID, m.message)
	switch message := err.(type) {
	case *schemes.Error:
		return message.Message, nil
	}
	return schemes.Message{}, err
}

func (a *messages) sendMessage(vip bool, reset bool, chatID int64, userID int64, message *schemes.NewMessageBody) (string, error) {
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
			return result.Message.Body.Mid, result
		}

	}
	return "", result
}

func (a *messages) editMessage(messageID int64, message *schemes.NewMessageBody) (*schemes.SimpleQueryResult, error) {
	result := new(schemes.SimpleQueryResult)
	values := url.Values{}
	values.Set("message_id", strconv.Itoa(int(messageID)))
	body, err := a.client.request(http.MethodPut, "messages", values, false, message)
	if err != nil {
		return result, err
	}
	defer func() {
		if err := body.Close(); err != nil {
			log.Println(err)
		}
	}()
	return result, json.NewDecoder(body).Decode(result)
}

// Check posiable to send a message to a chat.
func (a *messages) Check(m *Message) (bool, error) {
	return a.checkUser(m.reset, m.message)
}

func (a *messages) checkUser(reset bool, message *schemes.NewMessageBody) (bool, error) {
	result := new(schemes.Error)
	values := url.Values{}
	if reset {
		values.Set("access_token", message.BotToken)
	}
	mode := "notify/exists"

	if message.PhoneNumbers != nil {
		values.Set("phone_numbers", strings.Join(message.PhoneNumbers, ","))
	}

	body, err := a.client.request(http.MethodGet, mode, values, reset, nil)
	if err != nil {
		return false, err
	}
	defer body.Close()
	if err := json.NewDecoder(body).Decode(result); err != nil {
		// Message sent without errors
		return false, err
	}
	if len(result.NumberExist) > 0 {
		return true, result
	}
	return false, result
}
