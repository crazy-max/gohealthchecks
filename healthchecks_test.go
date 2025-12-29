package gohealthchecks_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/crazy-max/gohealthchecks"
)

const uuid = "5bf66975-d4c7-4bf5-bcc8-b8d8a82ea278"

func TestClient(t *testing.T) {
	client, mux, _, teardown := newClient(t)
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/%s", uuid), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`OK`))
	})

	if err := client.Success(context.Background(), gohealthchecks.PingingOptions{
		UUID: uuid,
	}); err != nil {
		t.Fatal(err)
	}
}

func newClient(t testing.TB) (client *gohealthchecks.Client, mux *http.ServeMux, baseURL *url.URL, teardown func()) {
	t.Helper()

	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	client = gohealthchecks.NewClient(&gohealthchecks.ClientOptions{
		BaseURL:    baseURL,
		HTTPClient: server.Client(),
	})

	teardown = func() {
		server.Close()
	}

	return client, mux, baseURL, teardown
}
