package commands

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

const dialTimeout = 5 * time.Minute

// RedisManager manages redis server pools
type RedisManager struct {
	pools map[string]*redis.Pool
	host  string
}

// NewRedisManager creates an instance of RedisManager
func NewRedisManager() *RedisManager {
	return &RedisManager{
		pools: make(map[string]*redis.Pool),
	}
}

// AddPool adds a redis pool (new redis server connection)
func (r *RedisManager) AddPool(host string, password string) {
	if p, ok := r.pools[host]; ok {
		p.Close()
		delete(r.pools, host)
	}
	r.pools[host] = &redis.Pool{
		MaxActive:   500,
		MaxIdle:     500,
		IdleTimeout: 5 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", host, dialTimeout, dialTimeout, dialTimeout)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
	}
}

// Do executes an arbitrary redis command
func (r *RedisManager) Do(host, cmd string, a []string) (interface{}, error) {
	pool := r.pools[host]
	if pool == nil {
		return nil, fmt.Errorf("no pool for host %s", host)
	}
	p := pool.Get()
	defer p.Close()
	var args []interface{}
	for _, v := range a {
		args = append(args, v)
	}
	v, err := p.Do(cmd, args...)
	return v, err
}

// Sub subscribes to a given channel on redis
func (r *RedisManager) Sub(host string, channel string) chan string {
	for {
		result := make(chan string)
		pool := r.pools[host]
		if pool == nil {
			return nil
		}
		p := pool.Get()
		psc := redis.PubSubConn{p}

		psc.Subscribe(channel)

		go func() {
			// While not a permanent error on the connection.
			for p.Err() == nil {
				switch v := psc.Receive().(type) {
				case redis.Message:
					result <- fmt.Sprintf("%s: message: %s\n", v.Channel, v.Data)
				case redis.Subscription:
					result <- fmt.Sprintf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
				case error:
					result <- fmt.Sprintf("err: %s", v.Error())
				}
			}
		}()

		return result
	}
}

// Pub publishes a message to redis
func (r *RedisManager) Pub(host string, channel string, msg string) (interface{}, error) {
	for {
		pool := r.pools[host]
		if pool == nil {
			return nil, fmt.Errorf("no pool for host %s", host)
		}

		p := pool.Get()
		defer p.Close()

		return p.Do("PUBLISH", channel, msg)
	}
}
