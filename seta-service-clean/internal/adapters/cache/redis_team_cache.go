package cache

import (
	"context"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RedisTeamCache implements the TeamMemberCache port using Redis.
type RedisTeamCache struct {
	rdb *redis.Client
}

// NewRedisTeamCache creates a new instance of RedisTeamCache.
func NewRedisTeamCache(rdb *redis.Client) ports.TeamMemberCache {
	return &RedisTeamCache{rdb: rdb}
}

func (c *RedisTeamCache) GetTeamMembers(ctx context.Context, teamID common.TeamID) ([]common.UserID, error) {
	key := "team:" + teamID.String() + ":members"
	idStrings, err := c.rdb.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(idStrings) == 0 {
		return nil, redis.Nil // Return redis.Nil on cache miss to be explicit
	}

	userIDs := make([]common.UserID, len(idStrings))
	for i, idStr := range idStrings {
		uid, err := uuid.Parse(idStr)
		if err != nil {
			return nil, err // Data corruption in cache
		}
		userIDs[i] = common.UserID(uid)
	}
	return userIDs, nil
}

func (c *RedisTeamCache) SetTeamMembers(ctx context.Context, teamID common.TeamID, memberIDs []common.UserID) error {
	key := "team:" + teamID.String() + ":members"
	if len(memberIDs) == 0 {
		return nil // Don't cache empty sets
	}

	stringIDs := make([]interface{}, len(memberIDs))
	for i, id := range memberIDs {
		stringIDs[i] = id.String()
	}

	pipe := c.rdb.Pipeline()
	pipe.SAdd(ctx, key, stringIDs...)
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

func (c *RedisTeamCache) InvalidateTeamMembers(ctx context.Context, teamID common.TeamID) error {
	key := "team:" + teamID.String() + ":members"
	return c.rdb.Del(ctx, key).Err()
}
