package cache

import (
	"context"
	"encoding/json"
	"seta/internal/application/ports"
	"seta/internal/domain/common"
	"seta/internal/domain/folder"
	"seta/internal/domain/note"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisAssetMetaCache implements the AssetMetaCache port using Redis.
type RedisAssetMetaCache struct {
	rdb *redis.Client
}

// NewRedisAssetMetaCache creates a new instance of RedisAssetMetaCache.
func NewRedisAssetMetaCache(rdb *redis.Client) ports.AssetMetaCache {
	return &RedisAssetMetaCache{rdb: rdb}
}

// --- Folder Cache ---

func (c *RedisAssetMetaCache) GetFolder(ctx context.Context, folderID common.FolderID) (*folder.Folder, error) {
	key := "folder:" + folderID.String()
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var f folder.Folder
	if err := json.Unmarshal([]byte(val), &f); err != nil {
		return nil, err // Data corruption
	}
	return &f, nil
}

func (c *RedisAssetMetaCache) SetFolder(ctx context.Context, f *folder.Folder) error {
	key := "folder:" + f.ID.String()
	val, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, val, 24*time.Hour).Err()
}

func (c *RedisAssetMetaCache) InvalidateFolder(ctx context.Context, folderID common.FolderID) error {
	key := "folder:" + folderID.String()
	return c.rdb.Del(ctx, key).Err()
}

// --- Note Cache ---

func (c *RedisAssetMetaCache) GetNote(ctx context.Context, noteID common.NoteID) (*note.Note, error) {
	key := "note:" + noteID.String()
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var n note.Note
	if err := json.Unmarshal([]byte(val), &n); err != nil {
		return nil, err // Data corruption
	}
	return &n, nil
}

func (c *RedisAssetMetaCache) SetNote(ctx context.Context, n *note.Note) error {
	key := "note:" + n.ID.String()
	val, err := json.Marshal(n)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, val, 24*time.Hour).Err()
}

func (c *RedisAssetMetaCache) InvalidateNote(ctx context.Context, noteID common.NoteID) error {
	key := "note:" + noteID.String()
	return c.rdb.Del(ctx, key).Err()
}
