package fake

import (
	redis "gopkg.in/redis.v5"
)

// RedisClient - mock implementaiton of a redis client
type RedisClient struct {
	InfoRes *redis.StringCmd
}

// Info - mock implementation of info
func (client *RedisClient) Info(section ...string) *redis.StringCmd {
	return client.InfoRes
}
