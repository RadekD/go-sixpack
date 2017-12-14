package sixpack

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

var expRe = regexp.MustCompile(`^[a-z0-9][a-z0-9\-_ ]*$`)

//Option define option type
type Option func(params url.Values)

//WithAlternatives sets traffic_fraction param
func WithAlternatives(alternatives ...string) Option {
	return func(params url.Values) {
		if len(alternatives) < 2 {
			panic("Must specify at least 2 alternatives")
		}

		for _, alt := range alternatives {
			if !expRe.MatchString(alt) {
				panic(fmt.Sprintf("Bad alternative name: %s", alt))
			}
			params.Add("alternatives", alt)
		}
	}
}

//WithForce sets force param
func WithForce(alternative string) Option {
	return func(params url.Values) {
		if !expRe.MatchString(alternative) {
			panic(fmt.Sprintf("Bad force name: %s", alternative))
		}
		params.Add("force", alternative)
	}
}

//WithRequest sets params from http.Request
func WithRequest(w http.ResponseWriter, r *http.Request) Option {
	return func(params url.Values) {
		var clientID string

		cookie, err := r.Cookie("sixpack_client_id")
		if cookie != nil && err == nil {
			clientID = cookie.String()
		} else {
			clientID = randomClientID(32)

			http.SetCookie(w, &http.Cookie{Name: "sixpack_client_id", Value: clientID, Expires: time.Now().Add(30 * 24 * time.Hour)})
		}

		WithClientID(clientID)(params)
		WithIP(r.RemoteAddr)(params)
		WithUserAgent(r.UserAgent())(params)
	}
}

//WithTrafficFraction sets traffic_fraction param
func WithTrafficFraction(fraction float64) Option {
	return func(params url.Values) {
		params.Set("traffic_fraction", fmt.Sprintf("%.2f", fraction))
	}
}

//WithClientID sets client_id param
func WithClientID(clientID string) Option {
	return func(params url.Values) {
		params.Set("client_id", clientID)
	}
}

//WithIP sets ip_address param
func WithIP(ip string) Option {
	return func(params url.Values) {
		params.Set("ip_address", ip)
	}
}

//WithUserAgent sets user_agent param
func WithUserAgent(userAgent string) Option {
	return func(params url.Values) {
		params.Set("user_agent", userAgent)
	}
}

//Client defines basic Sixpack functions
type Client interface {
	Participate(name string, opts ...Option) (string, error)
	Convert(name string, opts ...Option) error
}

func NewClient(baseURL string) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	return &client{
		baseURL: u,
	}, nil
}

type client struct {
	baseURL *url.URL
}

func (c *client) Participate(name string, opts ...Option) (string, error) {
	if !expRe.MatchString(name) {
		panic("Bad experiment name")
	}
	if len(opts) < 1 {
		panic("Required at least one option")
	}

	params := url.Values{}
	params.Set("experiment", name)
	for _, opt := range opts {
		opt(params)
	}
	if params.Get("alternatives") == "" {
		panic("WithAlternatives option is required")
	}
	if params.Get("client_id") == "" {
		params.Set("client_id", randomClientID(32))
	}

	return c.do("/participate", params)
}
func (c *client) Convert(name string, opts ...Option) error {
	if !expRe.MatchString(name) {
		panic("Bad experiment name")
	}
	if len(opts) < 1 {
		panic("Required at least one option")
	}

	params := url.Values{}
	params.Set("experiment", name)
	for _, opt := range opts {
		opt(params)
	}
	if params.Get("client_id") == "" {
		params.Set("client_id", randomClientID(32))
	}

	_, err := c.do("/convert", params)
	return err
}

func (c *client) do(endpoint string, params url.Values) (string, error) {
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	reqURL := c.baseURL.ResolveReference(endpointURL)
	reqURL.RawQuery = params.Encode()

	resp, err := http.Get(reqURL.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		Status      string `json:"status"`
		ClientID    string `json:"client_id"`
		Alternative struct {
			Name string `json:"name"`
		} `json:"alternative"`
		Experiment struct {
			Version int    `json:"version"`
			Name    string `json:"name"`
		} `json:"experiment"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}
	return response.Alternative.Name, nil
}

var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomClientID(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
