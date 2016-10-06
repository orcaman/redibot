package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/orcaman/redibot/commands"
)

const (
	cmdConnect    = "connect"
	cmdSubscribe  = "subscribe"
	cmdPublish    = "publish"
	cmdDisconnect = "disconnect"

	maxWSErrorsAllowed = 10
)

var errors = 0

func main() {
	token := os.Getenv("redibot_token")

	if token == "" {
		log.Fatalf("please set the redibot_token (your slack bot access token) and try again")
	}

	go handleSlack(token)
	handleHTTP()
}

func handleHTTP() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})

	log.Println("http listening")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSlack(token string) {
	redibot := commands.NewRedibot(token)

	go func() {
		for {
			m, connID, err := redibot.GetWSMessage()

			if err != nil {
				if errors >= maxWSErrorsAllowed {
					log.Fatalln(err.Error())
				}
				log.Println(err.Error())
				errors++
				continue
			}

			id := *connID

			if err != nil {
				log.Fatal(err)
			}

			if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {
				parts := strings.Fields(m.Text)
				cmd := parts[1]
				switch cmd {
				case cmdConnect:
					redibot.Connect(parts)
				case cmdPublish:
					redibot.Pub(m, parts)
				case cmdSubscribe:
					redibot.Sub(m, parts)
				default:
					redibot.Do(m, parts)
				}
			}
		}
	}()
}
