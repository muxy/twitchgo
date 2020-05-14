package twitch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const baseUrl = "https://api.twitch.tv/kraken"
const helixBaseUrl = "https://api.twitch.tv/helix"

type TwitchClient struct {
	HttpClient *http.Client
	ClientID   string
	Token      string
}

type TwitchClientOption func(*TwitchClient)

type RequestOptions struct {
	Limit     int64  `url:"limit"`
	Offset    int64  `url:"offset"`
	Direction string `url:"direction"`
	Nonce     int64  `url:"_"`
	Channel   string `url:"channel"`
	Version   string
	Extra     *url.Values
}

/* TwitchClient options */
func WithClientID(clientID string) TwitchClientOption {
	return func(cl *TwitchClient) {
		cl.ClientID = clientID
	}
}

func WithHTTPClient(httpClient *http.Client) TwitchClientOption {
	return func(cl *TwitchClient) {
		cl.HttpClient = httpClient
	}
}

func WithBearerToken(token string) TwitchClientOption {
	return func(cl *TwitchClient) {
		cl.Token = token
	}
}

func NewTwitchClient(opts ...TwitchClientOption) *TwitchClient {
	var (
		defaultClient = &http.Client{}
		defaultID     = ""
		defaultToken  = ""
	)

	cl := &TwitchClient{
		HttpClient: defaultClient,
		ClientID:   defaultID,
		Token:      defaultToken,
	}

	for _, opt := range opts {
		opt(cl)
	}

	return cl
}

func (client *TwitchClient) getRequest(endpoint string, options *RequestOptions, out interface{}) error {
	targetUrl := baseUrl + endpoint
	targetVersion := "3"

	if options != nil {
		if options.Version == "helix" {
			targetUrl = helixBaseUrl + endpoint
			targetVersion = "5"
		}

		v := url.Values{}

		if options.Direction != "" {
			v.Add("direction", options.Direction)
		}

		if options.Limit != 0 {
			v.Add("limit", fmt.Sprintf("%d", options.Limit))
		}

		if options.Offset != 0 {
			v.Add("offset", fmt.Sprintf("%d", options.Offset))
		}

		if options.Nonce != 0 {
			v.Add("_", fmt.Sprintf("%d", options.Nonce))
		}

		if options.Channel != "" {
			v.Add("channel", options.Channel)
		}

		if len(v) != 0 {
			targetUrl += "?" + v.Encode()
		}

		if options.Extra != nil {
			if len(v) != 0 {
				targetUrl += "&" + options.Extra.Encode()
			} else {
				targetUrl += "?" + options.Extra.Encode()
			}
		}
	}

	req, _ := http.NewRequest("GET", targetUrl, nil)
	req.Header.Set("Accept", fmt.Sprintf("application/vnd.twitchtv.v%s+json", targetVersion))
	req.Header.Set("Client-ID", client.ClientID)

	if len(client.Token) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.Token))
	}

	res, err := client.HttpClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Request failed with status: %v", res.StatusCode)
	}

	body, _ := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(body, out)
	if err != nil {
		return err
	}

	return nil
}
