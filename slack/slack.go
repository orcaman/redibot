package slack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

// WSConnection holds the websocket struct and an ID for the connection
type WSConnection struct {
	WS *websocket.Conn
	ID string
}

// Slack manages access to the Slack API functionality
type Slack struct {
	Token string
	Conn  *WSConnection
}

// These two structures represent the response of the Slack API rtm.start.
// Only some fields are included. The rest are ignored by json.Unmarshal.

type responseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	URL   string       `json:"url"`
	Self  responseSelf `json:"self"`
}

type responseSelf struct {
	ID string `json:"id"`
}

// NewSlack initiates an instane of Slack
func NewSlack(token string) *Slack {
	sl := &Slack{Token: token}
	ws, id := sl.Connect(token)
	sl.Conn = &WSConnection{ws, id}
	return sl
}

// rtmStart does a rtm.start, and returns a websocket URL and user ID. The
// websocket URL can be used to initiate an RTM session.
func rtmStart(token string) (wsurl, id string, err error) {
	url := fmt.Sprintf("https://slack.com/api/rtm.start?token=%s", token)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", resp.StatusCode)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return
	}
	var respObj responseRtmStart
	err = json.Unmarshal(body, &respObj)
	if err != nil {
		return
	}

	if !respObj.Ok {
		err = fmt.Errorf("Slack error: %s", respObj.Error)
		return
	}

	wsurl = respObj.URL
	id = respObj.Self.ID
	return
}

// Message reads off and is written into the websocket. Since this
// struct serves as both read and write, we include the "Id" field which is
// required only for writing.
type Message struct {
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

// GetMessage gets a message from the rtm channel
func (*Slack) GetMessage(ws *websocket.Conn) (m Message, err error) {
	err = ws.ReadJSON(&m)
	return
}

var counter uint64

// PostMessage posts a message to the rtm channel
func (*Slack) PostMessage(ws *websocket.Conn, m Message) error {
	m.ID = atomic.AddUint64(&counter, 1)

	return ws.WriteJSON(m)
}

// Connect to Slack RTM WebSocket API
func (*Slack) Connect(token string) (*websocket.Conn, string) {
	wsurl, id, err := rtmStart(token)
	if err != nil {
		log.Fatal(err)
	}

	ws, _, err := websocket.DefaultDialer.Dial(wsurl, nil)

	if err != nil {
		log.Fatal(err)
	}

	return ws, id
}
