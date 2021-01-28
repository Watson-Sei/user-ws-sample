package main

import (
	"encoding/json"
	"github.com/gofiber/websocket/v2"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strconv"
)

type client struct{}

type hub struct {
	clients map[*websocket.Conn]client
	register chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast chan string
}

var h = hub{
	clients: make(map[*websocket.Conn]client),
	register: make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
	broadcast: make(chan string),
}

var MemberCount int = 1

func (h *hub) run() {
	for {
		select {
		case connection := <- h.register:
			h.clients[connection] = client{}
			log.Println("connection registered")

			// create token and send token
			SID := strconv.Itoa(MemberCount)
			token := makeToken(SID)
			MemberCount += 1

			bytes, err := json.Marshal(map[string]interface{}{
				"event": "token",
				"token": token,
			})
			if err != nil {
				log.Println(err)
			}

			if err = connection.WriteMessage(websocket.TextMessage, bytes); err != nil {
				log.Println("write error:", err)

				h.unregister <- connection
				connection.WriteMessage(websocket.TextMessage, []byte{})
				connection.Close()
			}

		case connection := <- h.unregister:
			delete(h.clients, connection)
			log.Println("connection unregistered")

		case message := <- h.broadcast:
			log.Println("message received:", message)

			var dataMap map[string]interface{}
			if err := json.Unmarshal([]byte(message), &dataMap); err != nil {
				log.Println("error unmarshal: ", err)
			}

			if dataMap["event"] == "post" {
				bytes, err := json.Marshal(map[string]interface{}{
					"event": "member-post",
					"message": dataMap["message"],
					"token": dataMap["token"],
					"name": dataMap["name"],
				})
				if err != nil {
					log.Println(err)
				}

				for connection := range h.clients {
					if err = connection.WriteMessage(websocket.TextMessage, bytes); err != nil {
						log.Println("write error:", err)

						h.unregister <- connection
						connection.WriteMessage(websocket.TextMessage, []byte{})
						connection.Close()
					}
				}
			}
		}
	}
}

func makeToken(id string) string {
	str := SecretKey + id
	hash, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
