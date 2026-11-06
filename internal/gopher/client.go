// Package gopher implements a Gopher (RFC 1436) client from scratch:
// URL parsing, TCP fetching, and menu parsing. No third-party code.
package gopher

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Item is one line of a Gopher menu.
type Item struct {
	Type     string `json:"type"` // single character, e.g. "0" text, "1" menu, "7" search
	Display  string `json:"display"`
	Selector string `json:"selector"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	// URL is the gopher:// URL this item points to (empty for info lines).
	URL string `json:"url,omitempty"`
}

// URL is a parsed gopher:// URL.
type URL struct {
	Host     string
	Port     int
	Type     byte   // item type character; '1' (menu) when unspecified
	Selector string // path sent to the server; "" means the root menu
	Search   string // search terms for type-7 items
}

// ParseURL parses gopher://host[:port]/<type><selector> URLs.
// A bare host ("gopher.floodgap.com") is treated as its root menu.
func ParseURL(raw string) (URL, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return URL{}, fmt.Errorf("empty URL")
	}
	if !strings.Contains(s, "://") {
		s = "gopher://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return URL{}, err
	}
	if u.Scheme != "gopher" && u.Scheme != "gophers" {
		return URL{}, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	out := URL{Host: u.Hostname(), Port: 70, Type: '1'}
	if out.Host == "" {
		return URL{}, fmt.Errorf("missing host")
	}
	if p := u.Port(); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return URL{}, fmt.Errorf("invalid port %q", p)
		}
		out.Port = n
	}
	// Path is "/<type><selector...>"; the selector itself may contain slashes.
	path := strings.TrimPrefix(u.Path, "/")
	if len(path) >= 1 {
		out.Type = path[0]
		out.Selector = path[1:]
	}
	// Search terms arrive either %-encoded in the query or after a tab (%09).
	if q := u.Query().Get("search"); q != "" {
		out.Search = q
	}
	if i := strings.Index(out.Selector, "\t"); i >= 0 {
		out.Search = out.Selector[i+1:]
		out.Selector = out.Selector[:i]
	}
	return out, nil
}

// String reconstructs the canonical gopher:// URL.
func (u URL) String() string {
	s := fmt.Sprintf("gopher://%s:%d/%c%s", u.Host, u.Port, u.Type, u.Selector)
	if u.Search != "" {
		s += "%09" + url.QueryEscape(u.Search)
	}
	return s
}

// DefaultTimeout bounds a single fetch.
const DefaultTimeout = 8 * time.Second

// MaxResponse caps how much we read from one server.
const MaxResponse = 8 << 20 // 8 MiB

// Fetch retrieves the raw response for a Gopher URL: it opens a TCP
// connection, sends the selector (plus search terms for type 7), and
// reads the response, per RFC 1436.
//
// For menus and text, the response ends with a line containing only
// "." — Fetch stops there instead of waiting for the server to close
// the connection. Some servers (and some proxies in between) are slow
// to close, and waiting for EOF would hang an otherwise complete
// response. Binary items have no terminator and are read to EOF.
func Fetch(ctx context.Context, u URL) ([]byte, error) {
	d := net.Dialer{Timeout: DefaultTimeout}
	addr := net.JoinHostPort(u.Host, strconv.Itoa(u.Port))
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	} else {
		_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	}
	req := u.Selector
	if u.Search != "" {
		req += "\t" + u.Search
	}
	if _, err := io.WriteString(conn, req+"\r\n"); err != nil {
		return nil, err
	}
	if u.Type == '9' || u.Type == '5' || u.Type == '6' {
		// Binary payloads carry no terminator; read to EOF.
		return io.ReadAll(io.LimitReader(conn, MaxResponse))
	}
	var buf bytes.Buffer
	r := bufio.NewReader(io.LimitReader(conn, MaxResponse))
	for {
		line, err := r.ReadBytes('\n')
		buf.Write(line)
		if strings.TrimRight(string(line), "\r\n") == "." {
			break // protocol terminator: response complete
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			if buf.Len() > 0 && isTimeout(err) {
				// A complete-looking response whose connection is
				// being held open: deliver what arrived.
				break
			}
			return buf.Bytes(), err
		}
	}
	return buf.Bytes(), nil
}

func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// ParseMenu parses a Gopher menu response into items. It stops at the
// terminating "." line and tolerates both CRLF and LF line endings.
func ParseMenu(data []byte) []Item {
	var items []Item
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSuffix(sc.Text(), "\r")
		if line == "." {
			break
		}
		if line == "" {
			continue
		}
		it := Item{Type: line[:1]}
		parts := strings.Split(line[1:], "\t")
		it.Display = parts[0]
		if len(parts) > 1 {
			it.Selector = parts[1]
		}
		if len(parts) > 2 {
			it.Host = parts[2]
		}
		if len(parts) > 3 {
			if n, err := strconv.Atoi(strings.TrimSpace(parts[3])); err == nil {
				it.Port = n
			}
		}
		// Linkable items get a ready-to-use URL. Info lines ("i") and
		// errors ("3") are display-only.
		if it.Host != "" && it.Type != "i" && it.Type != "3" {
			port := it.Port
			if port == 0 {
				port = 70
			}
			child := URL{Host: it.Host, Port: port, Type: it.Type[0], Selector: it.Selector}
			it.URL = child.String()
		}
		items = append(items, it)
	}
	return items
}

// FetchMenu fetches and parses a menu URL.
func FetchMenu(ctx context.Context, u URL) ([]Item, []byte, error) {
	raw, err := Fetch(ctx, u)
	if err != nil {
		return nil, nil, err
	}
	return ParseMenu(raw), raw, nil
}
