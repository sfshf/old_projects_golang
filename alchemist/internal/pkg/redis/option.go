package redis

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Option configure the redis client
type Option struct {
	dns    string
	logger *zap.Logger
	Client *redis.Client
}

// NewOption create option
func NewOption(dns string, logger *zap.Logger) (*Option, error) {
	opt, err := redis.ParseURL(dns)
	if err != nil {
		return nil, err
	}
	return &Option{
		dns:    dns,
		logger: logger,
		Client: redis.NewClient(opt),
	}, nil
}
