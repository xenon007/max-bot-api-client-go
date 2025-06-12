// Package maxbot implements MAX Bot API.
// Official documentation: https://dev.max.ru/
package maxbot

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/max-messenger/max-bot-api-client-go/configservice"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

const (
	defaultTimeout = 120
	defaultPause   = 1

	maxRetries    = 3
)

// Api implements main part of Max Bot API
type Api struct {
	Bots          *bots
	Chats         *chats
	Debugs        *debugs
	Messages      *messages
	Subscriptions *subscriptions
	Uploads       *uploads
	client        *client
	timeout       int
	pause         int
	debug         bool
}

// New Max Bot Api object
func New(key string) *Api {
	u, err := url.Parse("https://botapi.max.ru/")
	if err != nil {
		fmt.Println(err.Error())
	}

	cl := newClient(key, "1.2.5", u, &http.Client{Timeout: time.Duration(defaultTimeout) * time.Second})
	return &Api{
		Bots:          newBots(cl),
		Chats:         newChats(cl),
		Uploads:       newUploads(cl),
		Messages:      newMessages(cl),
		Subscriptions: newSubscriptions(cl),
		Debugs:        newDebugs(cl, 0),
		client:        cl,
		timeout:       defaultTimeout,
		pause:         1,
	}
}

func NewFormConfig(cfg configservice.ConfigInterface) *Api {
	timeout := cfg.GetHttpBotAPITimeOut()
	u, err := url.Parse(cfg.GetHttpBotAPIUrl())
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	key := cfg.BotTokenCheckString()
	if key == "" {
		key = os.Getenv("TOKEN")
		if key == "" {
			fmt.Println("token is nil")
			return nil
		}
	}

	cl := newClient(key, cfg.GetHttpBotAPIVersion(), u, &http.Client{Timeout: time.Duration(timeout) * time.Second})
	return &Api{
		Bots:          newBots(cl),
		Chats:         newChats(cl),
		Uploads:       newUploads(cl),
		Messages:      newMessages(cl),
		Subscriptions: newSubscriptions(cl),
		Debugs:        newDebugs(cl, cfg.GetDebugLogChat()),
		client:        cl,
		timeout:       timeout,
		pause:         1,
		debug:         cfg.GetDebugLogMode(),
	}
}

func (a *Api) bytesToProperUpdate(b []byte) schemes.UpdateInterface {
	u := &schemes.Update{}
	_ = json.Unmarshal(b, u)
	if a.debug {
		u.DebugRaw = string(b)
	}
	switch u.GetUpdateType() {
	case schemes.TypeMessageCallback:
		upd := &schemes.MessageCallbackUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	case schemes.TypeMessageCreated:
		upd := &schemes.MessageCreatedUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		for _, att := range upd.Message.Body.RawAttachments {
			upd.Message.Body.Attachments = append(upd.Message.Body.Attachments, a.bytesToProperAttachment(att))
		}
		return upd
	case schemes.TypeMessageRemoved:
		upd := &schemes.MessageRemovedUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	case schemes.TypeMessageEdited:
		upd := &schemes.MessageEditedUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		for _, att := range upd.Message.Body.RawAttachments {
			upd.Message.Body.Attachments = append(upd.Message.Body.Attachments, a.bytesToProperAttachment(att))
		}
		return upd
	case schemes.TypeBotAdded:
		upd := &schemes.BotAddedToChatUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	case schemes.TypeBotRemoved:
		upd := &schemes.BotRemovedFromChatUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	case schemes.TypeUserAdded:
		upd := &schemes.UserAddedToChatUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	case schemes.TypeUserRemoved:
		upd := &schemes.UserRemovedFromChatUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	case schemes.TypeBotStarted:
		upd := &schemes.BotStartedUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	case schemes.TypeChatTitleChanged:
		upd := &schemes.ChatTitleChangedUpdate{Update: schemes.Update{DebugRaw: u.DebugRaw}}
		_ = json.Unmarshal(b, upd)
		return upd
	}
	return nil
}

func (a *Api) bytesToProperAttachment(b []byte) schemes.AttachmentInterface {
	attachment := new(schemes.Attachment)
	_ = json.Unmarshal(b, attachment)
	switch attachment.GetAttachmentType() {
	case schemes.AttachmentAudio:
		res := new(schemes.AudioAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentContact:
		res := new(schemes.ContactAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentFile:
		res := new(schemes.FileAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentImage:
		res := new(schemes.PhotoAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentKeyboard:
		res := new(schemes.InlineKeyboardAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentLocation:
		res := new(schemes.LocationAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentShare:
		res := new(schemes.ShareAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentSticker:
		res := new(schemes.StickerAttachment)
		_ = json.Unmarshal(b, res)
		return res
	case schemes.AttachmentVideo:
		res := new(schemes.VideoAttachment)
		_ = json.Unmarshal(b, res)
		return res
	}
	return attachment
}

func (a *Api) getUpdates(ctx context.Context, limit int, timeout int, marker int64, types []string) (*schemes.UpdateList, error) {
	result := new(schemes.UpdateList)
	values := url.Values{}
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}
	if timeout > 0 {
		values.Set("timeout", strconv.Itoa(timeout))
	}
	if marker > 0 {
		values.Set("marker", strconv.Itoa(int(marker)))
	}
	if len(types) > 0 {
		for _, t := range types {
			values.Add("types", t)
		}
	}
	
	body, err := a.client.requestWithContext(ctx, http.MethodGet, "updates", values, false, nil)
	if err != nil {
		if err == errLongPollTimeout {
			return result, nil
		}
		return result, fmt.Errorf("failed to request updates: %w", err)
	}
	defer func() {
		if err := body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()
	
	jb, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	if err := json.Unmarshal(jb, result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal updates: %w", err)
	}
	
	return result, nil
}

func (a *Api) getUpdatesWithRetry(ctx context.Context, limit int, timeout int, marker int64, types []string) (*schemes.UpdateList, error) {
	var result *schemes.UpdateList
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		result, lastErr = a.getUpdates(ctx, limit, timeout, marker, types)
		if lastErr == nil || lastErr == errLongPollTimeout {
			return result, lastErr
		}
		
		if attempt < maxRetries-1 {
			retryWait := time.Duration(1<<attempt)
			log.Printf("Attempt %d failed, retrying in %v: %v", attempt+1, retryWait, lastErr)
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryWait):
			}
		}
	}
	
	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// GetUpdates returns updates channel
func (a *Api) GetUpdates(ctx context.Context) chan schemes.UpdateInterface {
	ch := make(chan schemes.UpdateInterface)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case <-time.After(time.Duration(a.pause) * time.Second):
				var marker int64
				for {
					upds, err := a.getUpdatesWithRetry(ctx, 50, a.timeout, marker, []string{})
					if err != nil {
						select {
						case <-ctx.Done():
							return
						default:
							log.Printf("Error getting updates: %v", err)
							break
						}
					}
					
					if upds == nil || len(upds.Updates) == 0 {
						break
					}

					for _, u := range upds.Updates {
						ch <- a.bytesToProperUpdate(u)
					}

					if upds.Marker != nil {
						marker = *upds.Marker
					}
				}
			}
		}
	}()
	return ch
}

// GetHandler returns http handler for webhooks
func (a *Api) GetHandler(updates chan interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := r.Body.Close(); err != nil {
				log.Println(err)
			}
		}()
		b, _ := ioutil.ReadAll(r.Body)
		updates <- a.bytesToProperUpdate(b)
	}
}
