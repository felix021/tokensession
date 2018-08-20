package tokensession

import (
	"encoding/json"
	"fmt"
)

type Store interface {
	Load(session *TokenSession) error
	Save(session *TokenSession) error
	Delete(session *TokenSession) error
}

type TokenSession struct {
	Token  string
	MaxAge int
	Values map[string]interface{}
	Store  Store
}

const DefaultTokenName = "X-USER-TOKEN"
const DefaultMaxAge = 30 * 24 * 36000 //30 days

var (
	tokenName string = DefaultTokenName
	maxAge    int    = DefaultMaxAge
)

func SetTokenName(name string) {
	tokenName = name
}

func SetMaxAge(age int) {
	maxAge = age
}

// Serialize to JSON. Will err if there are unmarshalable key values
func (s TokenSession) Serialize() ([]byte, error) {
	m := make(map[string]interface{}, len(s.Values))
	for k, v := range s.Values {
		m[k] = v
	}
	return json.Marshal(m)
}

// Deserialize from JSON back to map[string]interface{}
func (s TokenSession) Deserialize(d []byte) error {
	m := make(map[string]interface{})
	err := json.Unmarshal(d, &m)
	if err != nil {
		fmt.Printf("TokenSession.Deserialize() Error: %v", err)
		return err
	}
	for k, v := range m {
		s.Values[k] = v
	}
	return nil
}

func NewTokenSession(token string, store Store) *TokenSession {
	return &TokenSession{
		Token:  token,
		MaxAge: maxAge,
		Values: make(map[string]interface{}),
		Store:  store,
	}
}

func (s TokenSession) String() string {
	return fmt.Sprintf("TokenSession{Token:%s, MaxAge: %d, Values: %s, Store: %s}", s.Token, s.MaxAge, s.Values, s.Store)
}

func (s TokenSession) Set(key string, value interface{}) {
	fmt.Println(key, value)
	s.Values[key] = value
}

func (s TokenSession) Get(key string) (interface{}, bool) {
	value, ok := s.Values[key]
	return value, ok
}

func (s TokenSession) Load() error {
	return s.Store.Load(&s)
}

func (s TokenSession) Save() error {
	return s.Store.Save(&s)
}

func (s TokenSession) Delete() error {
	err := s.Store.Delete(&s)
	if err != nil {
		return err
	}

	// Clear session values.
	for k := range s.Values {
		delete(s.Values, k)
	}
	return nil
}

func (s TokenSession) MustGet(key string) interface{} {
	value, _ := s.Values[key]
	return value
}

func (s TokenSession) MustGetInt(key string, default_value int) int {
	value := s.MustGet(key)
	type_value, ok := value.(int)
	if ok {
		return type_value
	} else {
		return default_value
	}
}

func (s TokenSession) MustGetInt64(key string, default_value int64) int64 {
	value := s.MustGet(key)
	type_value, ok := value.(int64)
	if ok {
		return type_value
	} else {
		return default_value
	}
}

func (s TokenSession) MustGetString(key string, default_value string) string {
	value := s.MustGet(key)
	type_value, ok := value.(string)
	if ok {
		return type_value
	} else {
		return default_value
	}
}

func (s TokenSession) MustGetFloat32(key string, default_value float32) float32 {
	value := s.MustGet(key)
	type_value, ok := value.(float32)
	if ok {
		return type_value
	} else {
		return default_value
	}
}

func (s TokenSession) MustGetFloat64(key string, default_value float64) float64 {
	value := s.MustGet(key)
	type_value, ok := value.(float64)
	if ok {
		return type_value
	} else {
		return default_value
	}
}
