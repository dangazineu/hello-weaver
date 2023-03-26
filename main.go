package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ServiceWeaver/weaver"
)

func reverse(c context.Context, reverser Reverser, name string, channel chan string, e chan error) {
	message, err := reverser.Reverse(c, name)
	if err != nil {
		e <- err
	}
	channel <- message
}

func main() {
	// Get a network listener on address "localhost:12345".
	root := weaver.Init(context.Background())
	opts := weaver.ListenerOptions{LocalAddress: "localhost:12345"}
	lis, err := root.Listener("hello", opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("hello listener available on %v\n", lis)

	reverser, err := weaver.Get[Reverser](root)
	if err != nil {
		log.Fatal(err)
	}

	// Serve the /hello endpoint.
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		result, err := make(chan string), make(chan error)

		go reverse(r.Context(), reverser, r.URL.Query().Get("name"), result, err)

		select {
		case message := <-result:
			fmt.Fprintf(w, "Hello, %s!\n", message)
		case e := <-err:
			fmt.Fprintf(w, "An error has occurred:  %s\n", e)
			log.Fatal(e)
		}
	})
	http.Serve(lis, nil)
}
