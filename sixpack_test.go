package sixpack_test

import (
	"fmt"
	"net/http"
	"testing"

	sixpack "github.com/RadekD/go-sixpack"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

func TestParicipate(t *testing.T) {
	experiment := "test-package"
	alternaives := []string{"a", "b", "c", "d"}
	clientID := "my-user-client-id"
	trafficFraction := 0.55
	force := "b"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://sixpack.test/participate",
		func(r *http.Request) (*http.Response, error) {

			params := r.URL.Query()

			for key, value := range params {
				if key == "alternatives" {
					for i, v := range value {
						if alternaives[i] != v {
							t.Fatalf("invalid alternative (got: %s, expected: %s)", v, alternaives[i])
						}
					}
				}
			}

			if params.Get("client_id") != clientID {
				t.Fatalf("invalid client id (got: %s, expected: %s)", params.Get("client_id"), clientID)
			}
			if params.Get("experiment") != experiment {
				t.Fatalf("invalid experiment (got: %s, expected: %s)", params.Get("experiment"), experiment)
			}
			if params.Get("traffic_fraction") != fmt.Sprintf("%.2f", trafficFraction) {
				t.Fatalf("invalid traffic_fraction (got: %s, expected: %.2f)", params.Get("traffic_fraction"), trafficFraction)
			}
			if params.Get("force") != force {
				t.Fatalf("invalid force (got: %s, expected: %s)", params.Get("force"), force)
			}

			return httpmock.NewStringResponse(200, `{"status": "OK", "client_id": "123456", "alternative": {"name": "my-alternative"}, "experiment": {"version": 1, "name": "my-test"}}`), nil
		},
	)

	client, err := sixpack.NewClient("https://sixpack.test")
	if err != nil {
		t.Error(err)
	}

	userAlt, err := client.Participate(experiment, sixpack.WithClientID(clientID),
		sixpack.WithAlternatives(alternaives...), sixpack.WithForce(force), sixpack.WithTrafficFraction(trafficFraction))
	if err != nil {
		t.Error(err)
	}
	if userAlt != "my-alternative" {
		t.Fatalf("invalid force (got: %s, expected: %s)", userAlt, "my-alternative")
	}
}
func TestConvert(t *testing.T) {
	experiment := "test-package"
	clientID := "my-user-client-id"
	trafficFraction := 0.55
	force := "b"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://sixpack.test/convert",
		func(r *http.Request) (*http.Response, error) {

			params := r.URL.Query()

			if params.Get("client_id") != clientID {
				t.Fatalf("invalid client id (got: %s, expected: %s)", params.Get("client_id"), clientID)
			}
			if params.Get("experiment") != experiment {
				t.Fatalf("invalid experiment (got: %s, expected: %s)", params.Get("experiment"), experiment)
			}
			if params.Get("traffic_fraction") != fmt.Sprintf("%.2f", trafficFraction) {
				t.Fatalf("invalid traffic_fraction (got: %s, expected: %.2f)", params.Get("traffic_fraction"), trafficFraction)
			}
			if params.Get("force") != force {
				t.Fatalf("invalid force (got: %s, expected: %s)", params.Get("force"), force)
			}

			return httpmock.NewStringResponse(200, `{"status": "OK", "client_id": "123456", "alternative": {"name": "my-alternative"}, "experiment": {"version": 1, "name": "my-test"}}`), nil
		},
	)

	client, err := sixpack.NewClient("https://sixpack.test")
	if err != nil {
		t.Fatal(err)
	}

	err = client.Convert(experiment, sixpack.WithClientID(clientID), sixpack.WithForce(force), sixpack.WithTrafficFraction(trafficFraction))
	if err != nil {
		t.Fatal(err)
	}
}
