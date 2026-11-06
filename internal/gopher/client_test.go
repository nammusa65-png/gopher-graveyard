package gopher

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestParseURL(t *testing.T) {
	cases := []struct {
		raw      string
		host     string
		port     int
		typ      byte
		selector string
	}{
		{"gopher://gopher.floodgap.com", "gopher.floodgap.com", 70, '1', ""},
		{"gopher.floodgap.com", "gopher.floodgap.com", 70, '1', ""},
		{"gopher://sdf.org:70/1/users", "sdf.org", 70, '1', "/users"},
		{"gopher://example.com:7070/0/about.txt", "example.com", 7070, '0', "/about.txt"},
		{"gopher://gopher.floodgap.com/7/v2/vs", "gopher.floodgap.com", 70, '7', "/v2/vs"},
	}
	for _, c := range cases {
		u, err := ParseURL(c.raw)
		if err != nil {
			t.Fatalf("ParseURL(%q): %v", c.raw, err)
		}
		if u.Host != c.host || u.Port != c.port || u.Type != c.typ || u.Selector != c.selector {
			t.Errorf("ParseURL(%q) = %+v, want host=%s port=%d type=%c selector=%q",
				c.raw, u, c.host, c.port, c.typ, c.selector)
		}
	}
}

func TestParseURLRejectsHTTP(t *testing.T) {
	if _, err := ParseURL("https://example.com/"); err == nil {
		t.Fatal("expected error for https URL")
	}
}

func TestParseMenu(t *testing.T) {
	data := "iWelcome to the graveyard\tTITLE\tfake\t0\r\n" +
		"1Live holes\t/world\tgopher.floodgap.com\t70\r\n" +
		"0About this server\t/about.txt\tsdf.org\t70\r\n" +
		"7Search Gopherspace\t/v2/vs\tgopher.floodgap.com\t70\r\n" +
		"3This host is gone\t\terr\t70\r\n" +
		".\r\n" +
		"1After the dot, must be ignored\t/x\th\t70\r\n"
	items := ParseMenu([]byte(data))
	if len(items) != 5 {
		t.Fatalf("got %d items, want 5: %+v", len(items), items)
	}
	if items[0].Type != "i" || items[0].URL != "" {
		t.Errorf("info line should have no URL: %+v", items[0])
	}
	if items[1].Type != "1" || items[1].URL != "gopher://gopher.floodgap.com:70/1/world" {
		t.Errorf("menu item URL wrong: %+v", items[1])
	}
	if items[2].Type != "0" || items[2].Host != "sdf.org" {
		t.Errorf("text item wrong: %+v", items[2])
	}
	if items[4].Type != "3" || items[4].URL != "" {
		t.Errorf("error item should have no URL: %+v", items[4])
	}
}

// TestFetchAgainstLocalServer spins up a minimal in-process Gopher
// server and verifies the client's request/response handling.
func TestFetchAgainstLocalServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 256)
		n, _ := conn.Read(buf)
		req := strings.TrimSpace(string(buf[:n]))
		if req == "" {
			fmt.Fprint(conn, "ihello\tTITLE\tfake\t0\r\n.\r\n")
		} else {
			fmt.Fprintf(conn, "0echo %s\t/x\th\t70\r\n.\r\n", req)
		}
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	u := URL{Host: "127.0.0.1", Port: port, Type: '1'}
	raw, err := Fetch(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	items := ParseMenu(raw)
	if len(items) != 1 || items[0].Display != "hello" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

// TestFetchStopsAtTerminator covers servers (and proxies) that send a
// complete response but are slow to close the connection: Fetch must
// return at the "." terminator instead of waiting for EOF.
func TestFetchStopsAtTerminator(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	release := make(chan struct{})
	defer close(release)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 64)
		_, _ = conn.Read(buf)
		fmt.Fprint(conn, "islow server\tTITLE\tfake\t0\r\n.\r\n")
		<-release // hold the connection open on purpose
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	u := URL{Host: "127.0.0.1", Port: port, Type: '1'}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := time.Now()
	raw, err := Fetch(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("Fetch took %v; it should stop at the terminator", elapsed)
	}
	items := ParseMenu(raw)
	if len(items) != 1 || items[0].Display != "slow server" {
		t.Fatalf("unexpected items: %+v", items)
	}
}
