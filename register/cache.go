package register

type Cache interface {
	GetValue(key string) (interface{}, bool)
	SetValue(key string, val interface{})
}

type memoryCache struct {
	data map[string]interface{}
}

func newMemoryCache() Cache {
	return &memoryCache{
		data: make(map[string]interface{}),
	}
}

func (c *memoryCache) GetValue(key string) (interface{}, bool) {
	val, ok := c.data[key]
	return val, ok
}
func (c *memoryCache) SetValue(key string, val interface{}) {
	c.data[key] = val
}
