package commands

import (
	"fmt"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/orcaman/redibot/slack"
)

var currentHost string

// Redibot facilitates the Slack and Redis connections required for the bot to operate
type Redibot struct {
	Slack        *slack.Slack
	RedisManager *RedisManager
}

// NewRedibot creates a new instance of Redibot
func NewRedibot(token string) *Redibot {
	redisManager := NewRedisManager()

	sl := slack.NewSlack(token)

	return &Redibot{
		Slack:        sl,
		RedisManager: redisManager,
	}
}

// Connect sets up a new redis connection
func (r *Redibot) Connect(parts []string) {
	currentHost = sanitizeHost(parts[2])
	pwd := []string{}
	if len(parts) > 3 {
		pwd = parts[3:]
	}
	r.RedisManager.AddPool(currentHost, strings.Join(pwd, " "))
}

// GetWSMessage gets a message off the web socket connection
func (r *Redibot) GetWSMessage() (*slack.Message, *string, error) {
	m, err := r.Slack.GetMessage(r.connection())
	if err != nil {
		return nil, nil, err
	}
	return &m, &r.Slack.Conn.ID, nil
}

// Pub publishes a message over the websocket
func (r *Redibot) Pub(m *slack.Message, parts []string) {
	channel := sanitizeHost(parts[2])
	msg := []string{}
	if len(parts) > 3 {
		msg = parts[3:]
	}
	go func(m *slack.Message) {
		_, err := r.RedisManager.Pub(currentHost, channel, strings.Join(msg, " "))
		if err != nil {
			m.Text = err.Error()
			r.Slack.PostMessage(r.connection(), *m)
		}
	}(m)
}

// Sub subscribes to a new web socket connection
func (r *Redibot) Sub(m *slack.Message, parts []string) {
	channel := sanitizeHost(parts[2])
	c := r.RedisManager.Sub(currentHost, channel)
	go func() {
		for msg := range c {
			m.Text = msg
			r.Slack.PostMessage(r.connection(), *m)
		}
	}()
}

// Do executes an arbitrary redis command
func (r *Redibot) Do(m *slack.Message, parts []string) {
	args := parts[2:]
	cmd := parts[1]
	go func(m *slack.Message) {
		v, err := r.RedisManager.Do(currentHost, cmd, args)
		if err != nil {
			m.Text = err.Error()
		} else {
			m.Text = fmt.Sprintf("%s", v)
		}
		r.Slack.PostMessage(r.connection(), *m)
	}(m)
}

func (r *Redibot) connection() *websocket.Conn {
	return r.Slack.Conn.WS
}

func sanitizeHost(s string) string {
	idx := strings.Index(s, "|")
	if idx == -1 {
		return s
	}
	s = s[:idx]
	return strings.Replace(s, "<http://", "", -1)
}
