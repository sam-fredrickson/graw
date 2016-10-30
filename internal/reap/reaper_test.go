package reaper

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/kylelemons/godebug/pretty"

	"github.com/turnage/graw/internal/client"
	"github.com/turnage/graw/internal/data"
)

type mockParser struct {
	comments []*data.Comment
	posts    []*data.Post
	messages []*data.Message
}

func (m *mockParser) Parse(
	blob json.RawMessage,
) ([]*data.Comment, []*data.Post, []*data.Message, error) {
	return m.comments, m.posts, m.messages, nil
}

func parserWhich(h Harvest) data.Parser {
	return &mockParser{
		comments: h.Comments,
		posts:    h.Posts,
		messages: h.Messages,
	}
}

type mockClient struct {
	request *http.Request
}

func (m *mockClient) Do(r *http.Request) ([]byte, error) {
	m.request = r
	return nil, nil
}

func newMockClient() client.Client {
	return &mockClient{}
}

func TestNew(t *testing.T) {
	cli := &mockClient{}
	par := &mockParser{}
	cfg := Config{
		Client:   cli,
		Parser:   par,
		Hostname: "reddit.com",
		TLS:      true,
	}
	expected := &reaper{
		cli:      cli,
		parser:   par,
		hostname: "reddit.com",
		scheme:   "https",
	}

	if diff := pretty.Compare(New(cfg), expected); diff != "" {
		t.Errorf("reaper construction incorrect; diff: %s", diff)
	}
}

func TestReaper(t *testing.T) {
	for i, test := range []struct {
		path    string
		values  map[string]string
		correct http.Request
	}{
		{"", nil, http.Request{
			Method: "GET",
			URL: &url.URL{
				Scheme:   "http",
				Host:     "",
				Path:     "",
				RawQuery: "",
			},
			Host: "",
		}},
		{"", map[string]string{"key": "value"}, http.Request{
			Method: "GET",
			URL: &url.URL{
				Scheme:   "http",
				Host:     "",
				Path:     "",
				RawQuery: "key=value",
			},
			Host: "",
		}},
		{"path", nil, http.Request{
			Method: "GET",
			URL: &url.URL{
				Scheme:   "http",
				Host:     "",
				Path:     "path",
				RawQuery: "",
			},
			Host: "",
		}},
	} {
		expected := Harvest{
			Comments: []*data.Comment{
				&data.Comment{
					Body: "comment",
				},
			},
			Posts: []*data.Post{
				&data.Post{
					SelfText: "post",
				},
			},
			Messages: []*data.Message{
				&data.Message{
					Body: "message",
				},
			},
		}
		c := &mockClient{}
		r := &reaper{
			cli:    c,
			parser: parserWhich(expected),
			scheme: "http",
		}

		harvest, err := r.Reap(test.path, test.values)
		if err != nil {
			t.Errorf("Error reaping input %d: %v", i, err)
		}

		if diff := pretty.Compare(harvest, expected); diff != "" {
			t.Errorf("harvest incorrect; diff: %s", diff)
		}

		if diff := pretty.Compare(c.request, test.correct); diff != "" {
			t.Errorf("request incorrect; diff: %s", diff)
		}
	}
}
