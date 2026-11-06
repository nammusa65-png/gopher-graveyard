// Package gopherserver implements a small Gopher server (RFC 1436)
// from scratch. It serves a directory tree: files named "gophermap"
// are sent verbatim as menus, other files are sent as documents, and
// directories without a gophermap get a generated listing.
package gopherserver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Server serves Gopher content rooted at Root.
type Server struct {
	Root string // directory containing the gopherhole content
	Host string // host name written into generated menus
	Port int    // port written into generated menus
}

// Serve accepts connections on ln until it is closed.
func (s *Server) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	r := bufio.NewReader(conn)
	line, err := r.ReadString('\n')
	if err != nil && line == "" {
		return
	}
	selector := strings.TrimRight(line, "\r\n")
	// Search requests append "\t<terms>"; this server does not index,
	// so answer searches with a polite error item.
	if strings.Contains(selector, "\t") {
		fmt.Fprintf(conn, "3Search is not supported on this server\t\t%s\t%d\r\n.\r\n", s.Host, s.Port)
		return
	}
	s.serve(conn, selector)
}

func (s *Server) serve(w io.Writer, selector string) {
	clean := path.Clean("/" + selector)
	full := filepath.Join(s.Root, filepath.FromSlash(clean))
	if rel, err := filepath.Rel(s.Root, full); err != nil || strings.HasPrefix(rel, "..") {
		errorItem(w, "Not found")
		return
	}
	info, err := os.Stat(full)
	if err != nil {
		errorItem(w, "Not found: "+selector)
		return
	}
	if info.IsDir() {
		gophermap := filepath.Join(full, "gophermap")
		if data, err := os.ReadFile(gophermap); err == nil {
			w.Write(data)
			if !strings.HasSuffix(string(data), ".\r\n") {
				io.WriteString(w, ".\r\n")
			}
			return
		}
		s.generatedMenu(w, full, clean)
		return
	}
	if filepath.Base(full) == "gophermap" {
		data, err := os.ReadFile(full)
		if err != nil {
			errorItem(w, "Not readable")
			return
		}
		w.Write(data)
		io.WriteString(w, ".\r\n")
		return
	}
	f, err := os.Open(full)
	if err != nil {
		errorItem(w, "Not readable")
		return
	}
	defer f.Close()
	io.Copy(w, f)
	// Text documents are conventionally terminated with a lone dot.
	if strings.HasSuffix(full, ".txt") || strings.HasSuffix(full, ".md") {
		io.WriteString(w, ".\r\n")
	}
}

func (s *Server) generatedMenu(w io.Writer, dir, selectorPrefix string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		errorItem(w, "Cannot list directory")
		return
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if e.Name() == "gophermap" {
			continue
		}
		typ := byte('9')
		if e.IsDir() {
			typ = '1'
		} else if strings.HasSuffix(e.Name(), ".txt") || strings.HasSuffix(e.Name(), ".md") {
			typ = '0'
		}
		sel := path.Join(selectorPrefix, e.Name())
		fmt.Fprintf(w, "%c%s\t%s\t%s\t%d\r\n", typ, e.Name(), sel, s.Host, s.Port)
	}
	io.WriteString(w, ".\r\n")
}

func errorItem(w io.Writer, msg string) {
	fmt.Fprintf(w, "3%s\t\tlocalhost\t70\r\n.\r\n", msg)
}
