// Command gopherd runs the project's own Gopher server, serving the
// gopherhole in ./content (or -root) on TCP port 70 (or -addr).
package main

import (
	"flag"
	"log"
	"net"
	"strconv"

	"github.com/nammusa65-png/gopher-graveyard/internal/gopherserver"
)

func main() {
	addr := flag.String("addr", ":70", "listen address (host:port)")
	root := flag.String("root", "content", "gopherhole content directory")
	host := flag.String("host", "localhost", "public host name written into generated menus")
	flag.Parse()

	_, portStr, _ := net.SplitHostPort(*addr)
	port, _ := strconv.Atoi(portStr)

	s := &gopherserver.Server{Root: *root, Host: *host, Port: port}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}
	log.Printf("gopherd serving %s on %s (public host %s)", *root, *addr, *host)
	log.Fatal(s.Serve(ln))
}
