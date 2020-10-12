package gohealthchecks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const (
	version        = "0.1.0"
	defaultBaseURL = "https://hc-ping.com/"
)

// ClientOptions holds optional parameters for the Client.
type ClientOptions struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
}

// Client manages communication with the Healthchecks API.
type Client struct {
	httpClient *http.Client
}

// NewClient constructs a new Client.
func NewClient(o *ClientOptions) *Client {
	if o == nil {
		o = new(ClientOptions)
	}

	c := o.HTTPClient
	if c == nil {
		c = new(http.Client)
	}

	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	baseURL := o.BaseURL
	if baseURL == nil {
		baseURL, _ = url.Parse(defaultBaseURL)
	}
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	c.Timeout = 10 * time.Second
	c.Transport = roundTripperFunc(func(r *http.Request) (resp *http.Response, err error) {
		r.Header.Set("User-Agent", fmt.Sprintf("gohealthchecks/%s go/%s %s", version, runtime.Version()[2:], strings.Title(runtime.GOOS)))
		u, err := baseURL.Parse(r.URL.String())
		if err != nil {
			return nil, err
		}
		r.URL = u
		return transport.RoundTrip(r)
	})

	return &Client{
		httpClient: c,
	}
}

// request handles the HTTP request response cycle. It creates an HTTP request
// with provided method on a path.
func (c *Client) request(ctx context.Context, method, path string, body []byte) (err error) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer drain(r.Body)

	if err = responseErrorHandler(r); err != nil {
		return err
	}

	return nil
}

// drain discards all of the remaining data from the reader and closes it,
// asynchronously.
func drain(r io.ReadCloser) {
	go func() {
		// Panicking here does not put data in
		// an inconsistent state.
		defer func() {
			_ = recover()
		}()

		_, _ = io.Copy(ioutil.Discard, r)
		r.Close()
	}()
}

// responseErrorHandler returns an error based on the HTTP status code or nil if
// the status code is from 200 to 299.
func responseErrorHandler(r *http.Response) (err error) {
	if r.StatusCode/100 == 2 {
		return nil
	}
	switch r.StatusCode {
	case http.StatusBadRequest:
		return decodeBadRequest(r)
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrTooManyRequests
	case http.StatusInternalServerError:
		return ErrInternalServerError
	case http.StatusServiceUnavailable:
		return ErrMaintenance
	default:
		return errors.New(strings.ToLower(r.Status))
	}
}

// decodeBadRequest parses the body of HTTP response that contains a list of
// errors as the result of bad request data.
func decodeBadRequest(r *http.Response) (err error) {

	type badRequestResponse struct {
		Errors []string `json:"errors"`
	}

	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return NewBadRequestError("bad request")
	}
	var e badRequestResponse
	if err = json.NewDecoder(r.Body).Decode(&e); err != nil {
		if err == io.EOF {
			return NewBadRequestError("bad request")
		}
		return err
	}
	return NewBadRequestError(e.Errors...)
}

// roundTripperFunc type is an adapter to allow the use of ordinary functions as
// http.RoundTripper interfaces. If f is a function with the appropriate
// signature, roundTripperFunc(f) is a http.RoundTripper that calls f.
type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip calls f(r).
func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
