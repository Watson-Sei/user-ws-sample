package main

import (
	"encoding/json"
	"github.com/gofiber/websocket/v2"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strconv"
)

type client struct{}

type message struct {
	data []byte
	conn *websocket.Conn
}

type hub struct {
	clients map[*websocket.Conn]client
	register chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast chan message
}

var h = hub{
	clients: make(map[*websocket.Conn]client),
	register: make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
	broadcast: make(chan message),
}

// チャット参加者一覧
var Member = make(map[*websocket.Conn]map[string]interface{})
// チャット延べ参加者
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

			// ユーザーリストに追加
			Member[connection] = make(map[string]interface{})
			Member[connection]["token"] = token
			Member[connection]["name"] = nil
			Member[connection]["count"] = MemberCount

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

		case m := <- h.broadcast:
			log.Println("message received:", string(m.data))

			var dataMap map[string]interface{}
			if err := json.Unmarshal(m.data, &dataMap); err != nil {
				log.Println("error unmarshal: ", err)
			}

			if dataMap["event"] == "join" {
				if authToken(m.conn, dataMap["token"].(string)) {
					log.Println("The token is valid.")
					// 入室OK + 現在の入室者一覧を通知
					memberlist := getMemberList()
					bytes, _ := json.Marshal(map[string]interface{}{
						"event": "join-result",
						"status": true,
						"list": memberlist,
					})
					if err := m.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
						log.Println("write error:", err)

						h.unregister <- m.conn
						m.conn.WriteMessage(websocket.TextMessage, []byte{})
						m.conn.Close()
					}

					// メンバー一覧に追加
					Member[m.conn]["name"] = dataMap["name"]

					// 入室通知
					// sender
					dataMap["event"] = "member-join"
					bytes, _ = json.Marshal(dataMap)
					if err := m.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
						log.Println("write error:", err)

						m.conn.WriteMessage(websocket.TextMessage, []byte{})
						m.conn.Close()
					}
					// others
					dataMap["token"] = Member[m.conn]["count"]
					bytes, _ = json.Marshal(dataMap)
					for connection:= range h.clients {
						if connection != m.conn {
							if err := connection.WriteMessage(websocket.TextMessage, bytes); err != nil {
								log.Println("write error:", err)

								connection.WriteMessage(websocket.TextMessage, []byte{})
								connection.Close()
							}
						}
					}
				} else {  // トークンが誤っていた場合
					// NG notification to sender
					bytes, _ := json.Marshal(map[string]interface{}{
						"event": "join-result",
						"status": false,
					})
					if err := m.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
						log.Println("write error:", err)

						m.conn.WriteMessage(websocket.TextMessage, []byte{})
						m.conn.Close()
					}
				}
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

			if dataMap["event"] == "quit" {
				// トークンが正しければ
				if authToken(m.conn, dataMap["token"].(string)) {
					// 本人に通知
					bytes, _ := json.Marshal(map[string]interface{}{
						"event": "quit-result",
						"status": true,
					})
					log.Println(string(bytes))
					if err := m.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
						log.Println("write error:", err)

						h.unregister <- m.conn
						m.conn.WriteMessage(websocket.TextMessage, []byte{})
						m.conn.Close()
					}

					// 本人以外
					bytes, _ = json.Marshal(map[string]interface{}{
						"event": "member-quit",
						"token": Member[m.conn]["count"],
					})
					log.Println(string(bytes))
					for connection := range h.clients {
						if connection != m.conn {
							if err := connection.WriteMessage(websocket.TextMessage, bytes); err != nil {
								log.Println("write error:", err)

								h.unregister <- connection
								connection.WriteMessage(websocket.TextMessage, []byte{})
								connection.Close()
							}
						}
					}

					// 削除
					delete(Member, m.conn)
				} else {
					// 本人にNG通知
					bytes, _ := json.Marshal(map[string]interface{}{
						"event": "quit-result",
						"status": false,
					})
					if err := m.conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
						log.Println("write error:", err)

						h.unregister <- m.conn
						m.conn.WriteMessage(websocket.TextMessage, []byte{})
						m.conn.Close()
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

func authToken(conn *websocket.Conn, token string) bool {
	if _, ok := Member[conn]; ok {
		if token == Member[conn]["token"] {
			return true
		}
		return false
	}
	return false
}

func getMemberList() []map[string]interface{} {
	var list []map[string]interface{}
	for key, _ := range Member {
		cur := Member[key]
		if cur["name"] != nil {
			list = append(list, map[string]interface{}{
				"token": cur["count"],
				"name": cur["name"],
			})
		}
	}
	return list
}