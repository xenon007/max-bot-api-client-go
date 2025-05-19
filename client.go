package maxbot

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

var (
	errLongPollTimeout = errors.New("timeout")
)

type client struct {
	key        string
	version    string
	url        *url.URL
	httpClient *http.Client
}

func newClient(key string, version string, url *url.URL, httpClient *http.Client) *client {
	return &client{key: key, version: version, url: url, httpClient: httpClient}
}

func (cl *client) request(method, path string, query url.Values, reset bool, body interface{}) (io.ReadCloser, error) {
	if body != nil {
		j, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		return cl.requestReader(method, path, query, reset, bytes.NewReader(j))
	}
	return cl.requestReader(method, path, query, reset, nil)
}

func (cl *client) requestReader(method, path string, query url.Values, reset bool, body io.Reader) (io.ReadCloser, error) {
	u := *cl.url
	u.Path = path
	if !reset {
		query.Set("access_token", cl.key)
	}
	query.Set("v", cl.version)
	u.RawQuery = query.Encode()
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	resp, err := cl.httpClient.Do(req)
	if err != nil {
		err, ok := err.(*url.Error)
		if ok {
			if err.Timeout() {
				return nil, errLongPollTimeout
			}
		}
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		errObj := new(schemes.Error)
		err = json.NewDecoder(resp.Body).Decode(errObj)
		if err != nil {
			return nil, err
		}
		return resp.Body, errObj
	}
	return resp.Body, err
}
