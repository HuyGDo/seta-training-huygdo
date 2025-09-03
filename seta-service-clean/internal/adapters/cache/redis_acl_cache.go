package cache

import (
	"context"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisACLCache implements the ACLCache port using Redis Hashes.
type RedisACLCache struct {
	rdb *redis.Client
}

// NewRedisACLCache creates a new instance of RedisACLCache.
func NewRedisACLCache(rdb *redis.Client) ports.ACLCache {
	return &RedisACLCache{rdb: rdb}
}

func (c *RedisACLCache) GetACL(ctx context.Context, assetID uuid.UUID) (map[common.UserID]common.Access, error) {
	key := "asset:" + assetID.String() + ":acl"
	aclStrings, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(aclStrings) == 0 {
		return nil, redis.Nil // Explicit cache miss
	}

	acl := make(map[common.UserID]common.Access)
	for userIDStr, accessStr := range aclStrings {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, err // Data corruption
		}
		acl[common.UserID(uid)] = common.Access(accessStr)
	}
	return acl, nil
}

func (c *RedisACLCache) SetACL(ctx context.Context, assetID uuid.UUID, acl map[common.UserID]common.Access) error {
	key := "asset:" + assetID.String() + ":acl"
	if len(acl) == 0 {
		return nil // Don't cache empty ACLs
	}

	aclStrings := make(map[string]interface{}, len(acl))
	for userID, access := range acl {
		aclStrings[userID.String()] = string(access)
	}

	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, key, aclStrings)
	pipe.Expire(ctx, key, 1*time.Hour) // Shorter TTL for ACLs as they can change more often
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisACLCache) InvalidateACL(ctx context.Context, assetID uuid.UUID) error {
	key := "asset:" + assetID.String() + ":acl"
	return c.rdb.Del(ctx, key).Err()
}
