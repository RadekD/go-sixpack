package sixpack

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/nu7hatch/gouuid"
)

/*

	test := sixpack.Test{
		URL: "http://localhost:8000",
		Name: "my-test",
		Alternatives: []string{"a", "b", "c"},
		Fraction: 0.1,
	}

*/

var expRe = regexp.MustCompile(`^[a-z0-9][a-z0-9\-_ ]*$`)

type Test struct {
	URL          string
	Name         string
	Alternatives []string
	Fraction     float32
}

func (six Test) request(endpoint *url.URL, params url.Values) (string, *Response, error) {
	def := params.Get("force")
	if def == "" {
		def = six.Alternatives[0]
	}

	timeout := time.Duration(250 * time.Millisecond)
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, timeout)
		},
	}
	client := &http.Client{
		Transport: &transport,
	}

	baseurl, err := url.Parse(six.URL)
	if err != nil {
		return def, nil, err
	}

	u := baseurl.ResolveReference(endpoint)
	u.RawQuery = params.Encode()

	resp, err := client.Get(u.String())
	if err != nil {
		return def, nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return def, nil, err
	}

	if resp.StatusCode == http.StatusInternalServerError {
		return def, nil, fmt.Errorf("StatusInternalServerError")
	}
	var r *Response
	err = json.Unmarshal(b, &r)
	if err != nil {
		return def, r, err
	}
	def = r.Alternative.Name
	return def, r, err
}

func (six Test) Participate(clientID, ip, userAgent, force string) (string, *Response, error) {
	if len(clientID) == 0 {
		id, err := uuid.NewV4()
		if err != nil {
			return force, nil, err
		}
		clientID = id.String()
	}

	if !expRe.MatchString(six.Name) {
		return force, nil, errors.New("Bad experiment name")
	}

	if len(six.Alternatives) < 2 {
		return force, nil, errors.New("Must specify at least 2 alternatives")
	}

	for _, alt := range six.Alternatives {
		if !expRe.MatchString(alt) {
			return force, nil, fmt.Errorf("Bad alternative name: %s", alt)
		}
	}

	endpoint, _ := url.Parse("/participate")

	params := url.Values{}
	params.Set("client_id", clientID)
	if ip != "" {
		params.Set("ip_address", ip)
	}
	if userAgent != "" {
		params.Set("user_agent", userAgent)
	}
	params.Set("experiment", six.Name)
	params.Set("traffic_fraction", fmt.Sprintf("%.2f", six.Fraction))
	if force != "" {
		params.Set("force", force)
	}

	for _, alt := range six.Alternatives {
		params.Add("alternatives", alt)
	}

	return six.request(endpoint, params)
}

func (six Test) ParticipateFromRequest(w http.ResponseWriter, r *http.Request) (string, *Response, error) {
	var clientID string

	cookie, err := r.Cookie("sixpack_client_id")
	if cookie != nil && err == nil {
		clientID = cookie.String()
	} else {
		id, _ := uuid.NewV4()
		clientID = id.String()
		http.SetCookie(w, &http.Cookie{Name: "sixpack_client_id", Value: clientID, Expires: time.Now().Add(30 * 24 * time.Hour)})
	}
	ip := r.RemoteAddr
	userAgent := r.UserAgent()
	force := r.URL.Query().Get("sixpack-force-" + six.Name)

	return six.Participate(clientID, ip, userAgent, force)
}

func (six Test) Convert(clientID, ip, userAgent, kpi string) (string, *Response, error) {
	if len(clientID) == 0 {
		id, err := uuid.NewV4()
		if err != nil {
			return "", nil, err
		}
		clientID = id.String()
	}

	endpoint, _ := url.Parse("/convert")

	params := url.Values{}
	params.Set("client_id", clientID)
	if ip != "" {
		params.Set("ip_address", ip)
	}
	if userAgent != "" {
		params.Set("user_agent", userAgent)
	}
	params.Set("experiment", six.Name)
	if kpi != "" {
		params.Set("kpi", kpi)
	}

	return six.request(endpoint, params)
}

func (six Test) ConvertFromRequest(w http.ResponseWriter, r *http.Request, kpi string) (string, *Response, error) {
	var clientID string

	cookie, err := r.Cookie("sixpack_client_id")
	if cookie != nil && err == nil {
		clientID = cookie.String()
	} else {
		id, _ := uuid.NewV4()
		clientID = id.String()
		http.SetCookie(w, &http.Cookie{Name: "sixpack_client_id", Value: clientID, Expires: time.Now().Add(30 * 24 * time.Hour)})
	}
	ip := r.RemoteAddr
	userAgent := r.UserAgent()

	return six.Convert(clientID, ip, userAgent, kpi)
}

//Response is a response from sixpack server
type Response struct {
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
