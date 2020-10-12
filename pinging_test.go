package gohealthchecks_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/crazy-max/gohealthchecks"
)

const logs = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."

func TestSuccess(t *testing.T) {
	client, mux, _, teardown := newClient(t)
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/%s", uuid), func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`OK`))
	})

	if err := client.Success(context.Background(), gohealthchecks.PingingOptions{
		UUID: uuid,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSuccessLogs(t *testing.T) {
	client, mux, _, teardown := newClient(t)
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/%s", uuid), func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, "body", string(b), logs)
		w.Write([]byte(`OK`))
	})

	if err := client.Success(context.Background(), gohealthchecks.PingingOptions{
		UUID: uuid,
		Logs: logs,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestFail(t *testing.T) {
	client, mux, _, teardown := newClient(t)
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/%s/fail", uuid), func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`OK`))
	})

	if err := client.Fail(context.Background(), gohealthchecks.PingingOptions{
		UUID: uuid,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestFailLogs(t *testing.T) {
	client, mux, _, teardown := newClient(t)
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/%s/fail", uuid), func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, "body", string(b), logs)
		w.Write([]byte(`OK`))
	})

	if err := client.Fail(context.Background(), gohealthchecks.PingingOptions{
		UUID: uuid,
		Logs: logs,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestStart(t *testing.T) {
	client, mux, _, teardown := newClient(t)
	defer teardown()

	mux.HandleFunc(fmt.Sprintf("/%s/start", uuid), func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`OK`))
	})

	if err := client.Start(context.Background(), gohealthchecks.PingingOptions{
		UUID: uuid,
	}); err != nil {
		t.Fatal(err)
	}
}

func assertEqual(t *testing.T, name string, got, want interface{}) {
	t.Helper()

	if name != "" {
		name += ": "
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%sgot %+v, want %+v", name, got, want)
	}
}
