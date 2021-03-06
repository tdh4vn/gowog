package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type clientImpl struct {
	hub Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// ID
	id int32
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *clientImpl) ReadPump() {
	defer func() {
		fmt.Println("Close readpump")
		c.hub.UnRegister(c)
		c.conn.Close()
	}()
	for {
		fmt.Println("Waiting for message")
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// Client disconnect
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.hub.ReceiveMessage(message)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *clientImpl) WritePump() {
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *clientImpl) Close() {
	close(c.send)
}

// Send pushes message event to channel, so it can be processed concurrently
func (c *clientImpl) Send(message []byte) {
	c.send <- message
}

// GetSend returns Send channel
func (c *clientImpl) GetSend() chan []byte {
	return c.send
}

// GetID returns client ID
func (c *clientImpl) GetID() int32 {
	return c.id
}

// NewClient returns new client given hub
func NewClient(upgrader websocket.Upgrader, hub Hub, w http.ResponseWriter, r *http.Request) int32 {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return -1
	}
	client := &clientImpl{hub: hub, conn: conn, send: make(chan []byte, 256)}
	// We need to register client from hub.
	clientIDChan := client.hub.Register(client)
	client.id = <-clientIDChan

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
	return client.id
}
