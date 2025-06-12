package maxbot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (cl *client) requestWithContext(ctx context.Context, method, path string, query url.Values, reset bool, body interface{}) (io.ReadCloser, error) {
	if body != nil {
		j, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		return cl.requestReader(ctx, method, path, query, reset, bytes.NewReader(j))
	}
	return cl.requestReader(ctx, method, path, query, reset, nil)
}

// request - обратно совместимый метод, использующий context.Background()
// Deprecated: Use requestWithContext instead
func (cl *client) request(method, path string, query url.Values, reset bool, body interface{}) (io.ReadCloser, error) {
	return cl.requestWithContext(context.Background(), method, path, query, reset, body)
}

func (cl *client) requestReader(ctx context.Context, method, path string, query url.Values, reset bool, body io.Reader) (io.ReadCloser, error) {
	u := *cl.url
	u.Path = path
	if !reset {
		query.Set("access_token", cl.key)
	}
	query.Set("v", cl.version)
	u.RawQuery = query.Encode()
	
	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := cl.httpClient.Do(req)
	if err != nil {
		urlErr, ok := err.(*url.Error)
		if ok && urlErr.Timeout() {
			return nil, errLongPollTimeout
		}
		
		return nil, fmt.Errorf("request failed: %w", err)
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
