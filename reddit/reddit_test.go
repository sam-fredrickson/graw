package reddit

import (
	"net/http"
	"net/url"
	"sync"
	"testing"
)

// testCase is an expectation for a resulting request from a single method call
// on a Bot interface.
type testCase struct {
	name               string
	err                error
	f                  func(Bot) error
	correct            http.Request
	expectedFormValues map[string]string // Used to validate form data in request body
}

func TestAccount(t *testing.T) {
	testRequests(
		[]testCase{
			testCase{
				name: "Reply",
				f: func(b Bot) error {
					return b.Reply("name", "text")
				},
				correct: http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/api/comment",
						RawQuery: "",
					},
					Host:   "reddit.com",
					Header: formEncoding,
				},
				expectedFormValues: map[string]string{
					"text":     "text",
					"thing_id": "name",
				},
			},
			testCase{
				name: "GetReply",
				f: func(b Bot) error {
					_, err := b.GetReply("name", "text")
					return err
				},
				correct: http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/api/comment",
						RawQuery: "",
					},
					Host:   "reddit.com",
					Header: formEncoding,
				},
				expectedFormValues: map[string]string{
					"api_type": "json",
					"text":     "text",
					"thing_id": "name",
				},
			},
			testCase{
				name: "SendMessage",
				f: func(b Bot) error {
					return b.SendMessage("user", "subject", "text")
				},
				correct: http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/api/compose",
						RawQuery: "",
					},
					Host:   "reddit.com",
					Header: formEncoding,
				},
				expectedFormValues: map[string]string{
					"subject": "subject",
					"text":    "text",
					"to":      "user",
				},
			},
			testCase{
				name: "PostSelf",
				f: func(b Bot) error {
					return b.PostSelf("self", "title", "text")
				},
				correct: http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/api/submit",
						RawQuery: "",
					},
					Host:   "reddit.com",
					Header: formEncoding,
				},
				expectedFormValues: map[string]string{
					"kind":  "self",
					"sr":    "self",
					"text":  "text",
					"title": "title",
				},
			},
			testCase{
				name: "GetPostSelf",
				f: func(b Bot) error {
					_, err := b.GetPostSelf("self", "title", "text")
					return err
				},
				correct: http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/api/submit",
						RawQuery: "",
					},
					Host:   "reddit.com",
					Header: formEncoding,
				},
				expectedFormValues: map[string]string{
					"api_type": "json",
					"kind":     "self",
					"sr":       "self",
					"text":     "text",
					"title":    "title",
				},
			},
			testCase{
				name: "PostLink",
				f: func(b Bot) error {
					return b.PostLink("link", "title", "url")
				},
				correct: http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/api/submit",
						RawQuery: "",
					},
					Host:   "reddit.com",
					Header: formEncoding,
				},
				expectedFormValues: map[string]string{
					"kind":  "link",
					"sr":    "link",
					"title": "title",
					"url":   "url",
				},
			},
			testCase{
				name: "GetPostLink",
				f: func(b Bot) error {
					_, err := b.GetPostLink("link", "title", "url")
					return err
				},
				correct: http.Request{
					Method: "POST",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/api/submit",
						RawQuery: "",
					},
					Host:   "reddit.com",
					Header: formEncoding,
				},
				expectedFormValues: map[string]string{
					"api_type": "json",
					"kind":     "link",
					"sr":       "link",
					"title":    "title",
					"url":      "url",
				},
			},
		}, t,
	)
}

func TestScanner(t *testing.T) {
	testRequests(
		[]testCase{
			testCase{
				name: "Listing",
				f: func(b Bot) error {
					_, err := b.Listing("/r/all", "ref")
					return err
				},
				correct: http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/r/all.json",
						RawQuery: "before=ref&limit=100&raw_json=1",
					},
					Host: "reddit.com",
				},
			},
		}, t,
	)
}

func TestLurker(t *testing.T) {
	testRequests(
		[]testCase{
			testCase{
				name: "Thread",
				err:  ThreadDoesNotExistErr,
				f: func(b Bot) error {
					_, err := b.Thread("/permalink")
					return err
				},
				correct: http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme:   "https",
						Host:     "reddit.com",
						Path:     "/permalink.json",
						RawQuery: "raw_json=1",
					},
					Host: "reddit.com",
				},
			},
		}, t,
	)
}

func testRequests(cases []testCase, t *testing.T) {
	c := &mockClient{}
	r := &reaperImpl{
		cli:        c,
		parser:     &mockParser{},
		hostname:   "reddit.com",
		reapSuffix: ".json",
		scheme:     "https",
		mu:         &sync.Mutex{},
	}
	b := &bot{
		Account: newAccount(r),
		Lurker:  newLurker(r),
		Scanner: newScanner(r),
	}
	for _, test := range cases {
		if err := test.f(b); err != test.err {
			t.Errorf("[%s] unexpected error: %v", test.name, err)
		}

		// We only verify the Method, Host, and Path parts of the URL
		// This allows our implementation to change between query params and body
		if c.request.Method != test.correct.Method {
			t.Errorf("[%s] wrong method: got %s, want %s",
				test.name, c.request.Method, test.correct.Method)
		}

		if c.request.Host != test.correct.Host {
			t.Errorf("[%s] wrong host: got %s, want %s",
				test.name, c.request.Host, test.correct.Host)
		}

		if c.request.URL.Path != test.correct.URL.Path {
			t.Errorf("[%s] wrong path: got %s, want %s",
				test.name, c.request.URL.Path, test.correct.URL.Path)
		}

		// For POST requests with a body, verify that the content length is > 0
		if c.request.Method == "POST" && test.expectedFormValues != nil {
			if c.request.ContentLength <= 0 {
				t.Errorf("[%s] ContentLength should be > 0 for POST request with form data", test.name)
			}
		}
	}
}
