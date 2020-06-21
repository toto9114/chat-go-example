package chat

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

type Message struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

type Room struct {
	roomId    string
	clientMap map[*Client]bool
	ChanEnter chan *Client
	ChanLeave chan *Client
	broadcast chan []byte
}

func newRoom(roomId string) *Room {
	return &Room{
		roomId:    roomId,
		clientMap: make(map[*Client]bool, 5),
		broadcast: make(chan []byte),
	}
}

func (w *Room) run() {
	w.ChanEnter = make(chan *Client)
	w.ChanLeave = make(chan *Client)

	for {
		select {
		case client := <-w.ChanEnter:
			fmt.Println("클라이언트 입장")
			w.clientMap[client] = true
		case client := <-w.ChanLeave:
			if _, ok := w.clientMap[client]; ok {
				delete(w.clientMap, client)
				fmt.Println("클라이언트 퇴장")
				close(client.send)
				if len(w.clientMap) <= 0 {
					fmt.Println("모두 퇴장함.")
					delete(roomPool, w.roomId)
					break
				}
			}
		case message := <-w.broadcast:
			for client := range w.clientMap {
				client.send <- message
			}
		}
	}
}

type Client struct {
	room *Room
	conn *websocket.Conn
	send chan []byte
}

func NewClient(room *Room, c *websocket.Conn) (client *Client) {
	client = &Client{
		room: room,
		conn: c,
		send: make(chan []byte, 256),
	}
	go client.readPump()
	go client.writePump()
	return client
}

func (c *Client) readPump() {
	defer func() {
		c.room.ChanLeave <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		var result map[string]string
		err = json.Unmarshal(message, &result)
		if err != nil {
			break
		}
		fmt.Println("Read: " + result["message"])

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.room.broadcast <- []byte(result["message"])
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			fmt.Println("Write: " + string(message))
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

var addr = flag.String("addr", "0.0.0.0:1213", "http service address")
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// func serveHome(w http.ResponseWriter, r *http.Request) {
// 	log.Println(r.URL)
// 	if r.URL.Path != "/" {
// 		http.Error(w, "Not found", http.StatusNotFound)
// 		return
// 	}
// 	if r.Method != "GET" {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	http.ServeFile(w, r, "home.html")
// }

// func main() {
// 	room := newRoom("1")
// 	go room.run()

// 	http.HandleFunc("/", serveHome)

// 	http.HandleFunc("/ws/{}", func(w http.ResponseWriter, r *http.Request) {
// 		conn, err := upgrader.Upgrade(w, r, nil)
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		client := NewClient(room, conn)
// 		client.room.ChanEnter <- client
// 	})
// }

var roomPool = make(map[string]*Room, 0)

func Start(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	paths := strings.FieldsFunc(r.URL.Path, func(c rune) bool { return c == '/' })
	roomId := paths[len(paths)-1]
	fmt.Println("roomId: ", roomId)

	var room *Room
	if _, ok := roomPool[roomId]; !ok {
		room = newRoom(roomId)
		roomPool[roomId] = room
	} else {
		room = roomPool[roomId]
	}
	go room.run()

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	client := NewClient(room, conn)
	client.room.ChanEnter <- client

	return nil
}
