package redisstore

// modified from https://github.com/boj/redistore

import (
	"errors"
	"fmt"
	"time"

	"github.com/felix021/tokensession"
	"github.com/gomodule/redigo/redis"
)

const DefaultIdleTime = 240

// Amount of time for redis keys to expire.
const DefaultExpire = 30 * 86400

const DefaultMaxLength = 4096

const DefaultPrefix = "sess_"

// RedisStore stores sessions in a redis backend.
type RedisStore struct {
	Pool          *redis.Pool
	DefaultMaxAge int // default Redis TTL for a MaxAge == 0 session
	maxLength     int
	keyPrefix     string
}

func (s RedisStore) String() string {
	return fmt.Sprintf("RedisStore{keyPrefix:%s}", s.keyPrefix)
}

// SetMaxLength sets RedisStore.maxLength if the `l` argument is greater or equal 0
// maxLength restricts the maximum length of new sessions to l.
// If l is 0 there is no limit to the size of a session, use with caution.
// The default for a new RedisStore is 4096. Redis allows for max.
// value sizes of up to 512MB (http://redis.io/topics/data-types)
// Default: 4096,
func (s *RedisStore) SetMaxLength(l int) {
	if l >= 0 {
		s.maxLength = l
	}
}

// SetKeyPrefix set the prefix
func (s *RedisStore) SetKeyPrefix(p string) {
	s.keyPrefix = p
}

func dial(network, address, password string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

// NewRedisStore returns a new RedisStore.
// size: maximum number of idle connections.
func NewRedisStore(size int, network, address, password string) (*RedisStore, error) {
	return NewRedisStoreWithPool(&redis.Pool{
		MaxIdle:     size,
		IdleTimeout: DefaultIdleTime * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dial(network, address, password)
		},
	})
}

func dialWithDB(network, address, password, DB string) (redis.Conn, error) {
	c, err := dial(network, address, password)
	if err != nil {
		return nil, err
	}
	if _, err := c.Do("SELECT", DB); err != nil {
		c.Close()
		return nil, err
	}
	return c, err
}

// NewRedisStoreWithDB - like NewRedisStore but accepts `DB` parameter to select
// redis DB instead of using the default one ("0")
func NewRedisStoreWithDB(size int, network, address, password, DB string) (*RedisStore, error) {
	return NewRedisStoreWithPool(&redis.Pool{
		MaxIdle:     size,
		IdleTimeout: DefaultIdleTime * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dialWithDB(network, address, password, DB)
		},
	})
}

// NewRedisStoreWithPool instantiates a RedisStore with a *redis.Pool passed in.
func NewRedisStoreWithPool(pool *redis.Pool) (*RedisStore, error) {
	rs := &RedisStore{
		// http://godoc.org/github.com/garyburd/redigo/redis#Pool
		Pool:          pool,
		DefaultMaxAge: DefaultExpire,
		maxLength:     DefaultMaxLength,
		keyPrefix:     DefaultPrefix,
	}
	_, err := rs.ping()
	return rs, err
}

// Close closes the underlying *redis.Pool
func (s *RedisStore) Close() error {
	return s.Pool.Close()
}

// ping does an internal ping against a server to check if it is alive.
func (s *RedisStore) ping() (bool, error) {
	conn := s.Pool.Get()
	defer conn.Close()
	data, err := conn.Do("PING")
	if err != nil || data == nil {
		return false, err
	}
	return (data == "PONG"), nil
}

// load reads the session from redis.
// returns true if there is a sessoin data in DB
func (s *RedisStore) Load(session *tokensession.TokenSession) error {
	conn := s.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return err
	}
	data, err := conn.Do("GET", s.keyPrefix+session.Token)
	if err != nil {
		return err
	}
	if data == nil {
		return nil // no data was associated with this key
	}
	b, err := redis.Bytes(data, err)
	if err != nil {
		return err
	}
	return session.Deserialize(b)
}

// save stores the session in redis.
func (s *RedisStore) Save(session *tokensession.TokenSession) error {
	b, err := session.Serialize()
	if err != nil {
		return err
	}
	if s.maxLength != 0 && len(b) > s.maxLength {
		return errors.New("RedisStore: the value to store is too big")
	}
	conn := s.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return err
	}
	age := session.MaxAge
	if age == 0 {
		age = s.DefaultMaxAge
	}
	_, err = conn.Do("SETEX", s.keyPrefix+session.Token, age, b)
	return err
}

// delete removes keys from redis if MaxAge<0
func (s *RedisStore) Delete(session *tokensession.TokenSession) error {
	conn := s.Pool.Get()
	defer conn.Close()
	if _, err := conn.Do("DEL", s.keyPrefix+session.Token); err != nil {
		return err
	}
	return nil
}
