package moon

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	items map[string]interface{}
}

func (c *Config) hasKey(key string) bool {
	if c.items == nil {
		return false
	}
	_, ok := c.items[key]
	return ok
}

func (c *Config) setUnique(key string, value interface{}) error {
	if c.hasKey(key) {
		return fmt.Errorf("the name %s is already defined in this scope", key)
	}
	c.set(key, value)
	return nil
}

func (c *Config) set(key string, value interface{}) {
	if c.items == nil {
		c.items = make(map[string]interface{}, 12)
	}
	c.items[key] = value
}

func (c *Config) Get(key string) interface{} {
	if c.items == nil {
		return nil
	}
	return c.items[key]
}

func (c *Config) GetString(key string) string {
	v := c.Get(key)
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if ok {
		return s
	}
	return ""
}

func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.items)
}
