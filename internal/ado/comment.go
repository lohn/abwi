package ado

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/yuin/goldmark"
)

// Comment is a work item comment, decoded from the comments REST API. The SDK
// comments client cannot request Markdown format, so abwi calls REST directly.
type Comment struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	CreatedBy struct {
		DisplayName string `json:"displayName"`
	} `json:"createdBy"`
	CreatedDate string `json:"createdDate"`
}

// commentsURL builds the comments endpoint; format ("markdown" or "html") is
// added only when non-empty.
func commentsURL(orgURL, project string, id int, format string) string {
	u := fmt.Sprintf("%s/%s/_apis/wit/workItems/%d/comments?api-version=7.1-preview.4",
		strings.TrimRight(orgURL, "/"), url.PathEscape(project), id)
	if format != "" {
		u += "&format=" + format
	}
	return u
}

// AddComment posts a comment in the configured format. In html mode the
// Markdown input is converted before posting.
func (c *Client) AddComment(ctx context.Context, id int, text string) (*Comment, error) {
	if c.Format == FormatHTML {
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(text), &buf); err != nil {
			return nil, err
		}
		text = buf.String()
	}
	var out Comment
	err := c.rest(ctx, http.MethodPost, commentsURL(c.Org, c.Project, id, c.Format),
		map[string]string{"text": text}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ListComments fetches all comments of a work item.
func (c *Client) ListComments(ctx context.Context, id int) ([]Comment, error) {
	var out struct {
		Comments []Comment `json:"comments"`
	}
	err := c.rest(ctx, http.MethodGet, commentsURL(c.Org, c.Project, id, ""), nil, &out)
	if err != nil {
		return nil, err
	}
	return out.Comments, nil
}

// rest performs a plain REST call reusing the connection's authorization, for
// APIs the SDK does not cover.
func (c *Client) rest(ctx context.Context, method, u string, body, out any) error {
	var rd io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.Conn.AuthorizationString)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: %s: %s", method, u, resp.Status, bytes.TrimSpace(b))
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
