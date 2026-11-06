package gopherserver

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nammusa65-png/gopher-graveyard/internal/gopher"
)

func TestServeAndFetch(t *testing.T) {
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "gophermap"),
		[]byte("iTest hole\tTITLE\tfake\t0\r\n0About\t/about.txt\tHOST\tPORT\r\n"), 0o644)
	os.WriteFile(filepath.Join(root, "about.txt"), []byte("hello from the grave\n"), 0o644)
	os.Mkdir(filepath.Join(root, "docs"), 0o755)
	os.WriteFile(filepath.Join(root, "docs", "note.txt"), []byte("a note\n"), 0o644)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	s := &Server{Root: root, Host: "127.0.0.1", Port: port}
	go s.Serve(ln)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Root menu from gophermap (HOST/PORT placeholders pass through verbatim).
	u := gopher.URL{Host: "127.0.0.1", Port: port, Type: '1'}
	items, _, err := gopher.FetchMenu(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[1].Display != "About" {
		t.Fatalf("unexpected root menu: %+v", items)
	}

	// Text document.
	u2 := gopher.URL{Host: "127.0.0.1", Port: port, Type: '0', Selector: "/about.txt"}
	raw, err := gopher.Fetch(ctx, u2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "hello from the grave") {
		t.Fatalf("unexpected document: %q", raw)
	}

	// Generated listing for a directory without a gophermap.
	u3 := gopher.URL{Host: "127.0.0.1", Port: port, Type: '1', Selector: "/docs"}
	items, _, err = gopher.FetchMenu(ctx, u3)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Display != "note.txt" || items[0].Selector != "/docs/note.txt" {
		t.Fatalf("unexpected generated menu: %+v", items)
	}

	// Missing selector -> type 3 error item.
	u4 := gopher.URL{Host: "127.0.0.1", Port: port, Type: '1', Selector: "/nope"}
	items, _, err = gopher.FetchMenu(ctx, u4)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Type != "3" {
		t.Fatalf("expected error item, got %+v", items)
	}

	// Path traversal must not escape the root.
	u5 := gopher.URL{Host: "127.0.0.1", Port: port, Type: '0', Selector: "/../go.mod"}
	raw, err = gopher.Fetch(ctx, u5)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "module ") {
		t.Fatal("path traversal leaked a file outside the root")
	}
}
