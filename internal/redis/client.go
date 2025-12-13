package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"redis-explorer/internal/models"
)

// Client wraps the Redis client with additional functionality
type Client struct {
	rdb        *redis.Client
	connection *models.ServerConnection
	ctx        context.Context
}

// New creates a new Redis client from a server connection
func New(conn *models.ServerConnection) *Client {
	return &Client{
		connection: conn,
		ctx:        context.Background(),
	}
}

// Connect establishes a connection to the Redis server
func (c *Client) Connect() error {
	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.connection.Host, c.connection.Port),
		Password: c.connection.Password,
		DB:       c.connection.Database,
	}

	if c.connection.UseTLS {
		opts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: c.connection.Host, // Required for SNI verification
		}
	}

	c.rdb = redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	_, err := c.rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis at %s:%d: %w", c.connection.Host, c.connection.Port, err)
	}
	return nil
}

// Disconnect closes the Redis connection
func (c *Client) Disconnect() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

// IsConnected checks if the client is connected
func (c *Client) IsConnected() bool {
	if c.rdb == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(c.ctx, 2*time.Second)
	defer cancel()
	_, err := c.rdb.Ping(ctx).Result()
	return err == nil
}

// SelectDatabase changes the current database
func (c *Client) SelectDatabase(db int) error {
	return c.rdb.Do(c.ctx, "SELECT", db).Err()
}

// ScanKeys returns keys matching the pattern with pagination
func (c *Client) ScanKeys(pattern string, cursor uint64, count int64) ([]string, uint64, error) {
	if pattern == "" {
		pattern = "*"
	}
	keys, nextCursor, err := c.rdb.Scan(c.ctx, cursor, pattern, count).Result()
	return keys, nextCursor, err
}

// GetAllKeys returns all keys matching the pattern (use with caution on large databases)
func (c *Client) GetAllKeys(pattern string, maxKeys int) ([]models.RedisKey, error) {
	if pattern == "" {
		pattern = "*"
	}

	var keys []models.RedisKey
	var cursor uint64

	// Optimize scan count based on maxKeys
	scanCount := int64(100)
	if maxKeys > 0 && maxKeys < 100 {
		scanCount = int64(maxKeys)
	}

	for {
		result, nextCursor, err := c.rdb.Scan(c.ctx, cursor, pattern, scanCount).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		for _, key := range result {
			keyType, err := c.rdb.Type(c.ctx, key).Result()
			if err != nil {
				log.Printf("warning: failed to get type for key %s: %v", key, err)
				keyType = "unknown"
			}

			ttl, err := c.rdb.TTL(c.ctx, key).Result()
			if err != nil {
				log.Printf("warning: failed to get TTL for key %s: %v", key, err)
				ttl = -2 * time.Second
			}

			keys = append(keys, models.RedisKey{
				Key:  key,
				Type: keyType,
				TTL:  int64(ttl.Seconds()),
			})

			if maxKeys > 0 && len(keys) >= maxKeys {
				return keys, nil
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// GetKeyType returns the type of a key
func (c *Client) GetKeyType(key string) (string, error) {
	return c.rdb.Type(c.ctx, key).Result()
}

// GetTTL returns the TTL of a key in seconds
func (c *Client) GetTTL(key string) (int64, error) {
	ttl, err := c.rdb.TTL(c.ctx, key).Result()
	if err != nil {
		return -2, err
	}
	return int64(ttl.Seconds()), nil
}

// SetTTL sets the TTL for a key
func (c *Client) SetTTL(key string, seconds int64) error {
	if seconds <= 0 {
		return c.rdb.Persist(c.ctx, key).Err()
	}
	return c.rdb.Expire(c.ctx, key, time.Duration(seconds)*time.Second).Err()
}

// DeleteKey deletes a key
func (c *Client) DeleteKey(key string) error {
	return c.rdb.Del(c.ctx, key).Err()
}

// RenameKey renames a key
func (c *Client) RenameKey(oldKey, newKey string) error {
	return c.rdb.Rename(c.ctx, oldKey, newKey).Err()
}

// String operations

// GetString gets a string value
func (c *Client) GetString(key string) (string, error) {
	return c.rdb.Get(c.ctx, key).Result()
}

// SetString sets a string value
func (c *Client) SetString(key, value string) error {
	return c.rdb.Set(c.ctx, key, value, 0).Err()
}

// List operations

// GetList returns all elements in a list
func (c *Client) GetList(key string) ([]string, error) {
	return c.rdb.LRange(c.ctx, key, 0, -1).Result()
}

// ListPush adds an element to a list
func (c *Client) ListPush(key, value string, left bool) error {
	if left {
		return c.rdb.LPush(c.ctx, key, value).Err()
	}
	return c.rdb.RPush(c.ctx, key, value).Err()
}

// ListSet sets an element at index in a list
func (c *Client) ListSet(key string, index int64, value string) error {
	return c.rdb.LSet(c.ctx, key, index, value).Err()
}

// ListRemove removes elements from a list
func (c *Client) ListRemove(key string, count int64, value string) error {
	return c.rdb.LRem(c.ctx, key, count, value).Err()
}

// Set operations

// GetSet returns all members of a set
func (c *Client) GetSet(key string) ([]string, error) {
	return c.rdb.SMembers(c.ctx, key).Result()
}

// SetAdd adds a member to a set
func (c *Client) SetAdd(key, member string) error {
	return c.rdb.SAdd(c.ctx, key, member).Err()
}

// SetRemove removes a member from a set
func (c *Client) SetRemove(key, member string) error {
	return c.rdb.SRem(c.ctx, key, member).Err()
}

// Hash operations

// GetHash returns all fields and values in a hash
func (c *Client) GetHash(key string) (map[string]string, error) {
	return c.rdb.HGetAll(c.ctx, key).Result()
}

// HashSet sets a field in a hash
func (c *Client) HashSet(key, field, value string) error {
	return c.rdb.HSet(c.ctx, key, field, value).Err()
}

// HashDelete deletes a field from a hash
func (c *Client) HashDelete(key, field string) error {
	return c.rdb.HDel(c.ctx, key, field).Err()
}

// Sorted Set operations

// GetSortedSet returns all members with scores in a sorted set
func (c *Client) GetSortedSet(key string) ([]models.ScoredValue, error) {
	result, err := c.rdb.ZRangeWithScores(c.ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var values []models.ScoredValue
	for _, z := range result {
		values = append(values, models.ScoredValue{
			Score:  z.Score,
			Member: z.Member.(string),
		})
	}
	return values, nil
}

// SortedSetAdd adds a member with score to a sorted set
func (c *Client) SortedSetAdd(key string, score float64, member string) error {
	return c.rdb.ZAdd(c.ctx, key, redis.Z{Score: score, Member: member}).Err()
}

// SortedSetRemove removes a member from a sorted set
func (c *Client) SortedSetRemove(key, member string) error {
	return c.rdb.ZRem(c.ctx, key, member).Err()
}

// Server information

// GetServerInfo returns server information
func (c *Client) GetServerInfo() (*models.ServerInfo, error) {
	info, err := c.rdb.Info(c.ctx).Result()
	if err != nil {
		return nil, err
	}

	serverInfo := &models.ServerInfo{}
	lines := strings.Split(info, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]

		switch key {
		case "redis_version":
			serverInfo.Version = value
		case "redis_mode":
			serverInfo.Mode = value
		case "os":
			serverInfo.OS = value
		case "uptime_in_seconds":
			serverInfo.Uptime, _ = strconv.ParseInt(value, 10, 64)
		case "connected_clients":
			serverInfo.ConnectedClients, _ = strconv.ParseInt(value, 10, 64)
		case "used_memory":
			serverInfo.UsedMemory, _ = strconv.ParseInt(value, 10, 64)
		case "used_memory_human":
			serverInfo.UsedMemoryHuman = value
		case "used_memory_peak":
			serverInfo.UsedMemoryPeak, _ = strconv.ParseInt(value, 10, 64)
		case "expired_keys":
			serverInfo.ExpiredKeys, _ = strconv.ParseInt(value, 10, 64)
		case "keyspace_hits":
			serverInfo.KeyspaceHits, _ = strconv.ParseInt(value, 10, 64)
		case "keyspace_misses":
			serverInfo.KeyspaceMisses, _ = strconv.ParseInt(value, 10, 64)
		}
	}

	// Get total keys count
	dbSize, err := c.rdb.DBSize(c.ctx).Result()
	if err == nil {
		serverInfo.TotalKeys = dbSize
	}

	return serverInfo, nil
}

// GetDatabaseCount returns the number of databases
func (c *Client) GetDatabaseCount() int {
	// Try to get from server config
	result, err := c.rdb.ConfigGet(c.ctx, "databases").Result()
	if err == nil && len(result) >= 2 {
		if dbStr, ok := result["databases"]; ok {
			if count, err := strconv.Atoi(dbStr); err == nil && count > 0 {
				return count
			}
		}
	}
	// Default Redis has 16 databases (0-15)
	return 16
}

// FlushDB flushes the current database
func (c *Client) FlushDB() error {
	return c.rdb.FlushDB(c.ctx).Err()
}

// GetKeyCount returns the number of keys in the current database
func (c *Client) GetKeyCount() (int64, error) {
	return c.rdb.DBSize(c.ctx).Result()
}
