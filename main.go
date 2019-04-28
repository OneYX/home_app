package main

import (
	"flag"
	"fmt"
	"log"
)

var (
	hostname string
	port     int
)

/* register command line options */
func init() {
	flag.StringVar(&hostname, "hostname", "0.0.0.0", "The hostname or IP on which the server will listen")
	flag.IntVar(&port, "port", 8080, "The port on which the server will listen")
}

func main() {
	flag.Parse()
	var address = fmt.Sprintf("%s:%d", hostname, port)
	log.Println("service listening on", address)

	a := App{}
	a.Initialize(
		"test",
		"test",
		"home")

	a.Run(address)
}
