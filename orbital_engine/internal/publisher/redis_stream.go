package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// AdjacencyEntry represents network topology state
type AdjacencyEntry struct {
	Timestamp time.Time  `json:"timestamp"`
	Satellite int        `json:"satellite_id"`
	Adjacency []int      `json:"adjacent_satellites"`
	Position  [3]float64 `json:"position"`
	Velocity  [3]float64 `json:"velocity"`
}

// StateSnapshot is the complete network state
type StateSnapshot struct {
	Timestamp time.Time
	Entries   map[int]AdjacencyEntry
}

// RedisPublisher manages publishing network topology to Redis Stream
type RedisPublisher struct {
	client        *redis.Client
	streamKey     string
	snapshotKey   string
	mu            sync.RWMutex
	lastPublished time.Time
	publishRate   time.Duration
}

// NewRedisPublisher creates a new Redis publisher
func NewRedisPublisher(redisAddr, streamKey, snapshotKey string) (*RedisPublisher, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisPublisher{
		client:        client,
		streamKey:     streamKey,
		snapshotKey:   snapshotKey,
		publishRate:   100 * time.Millisecond, // 10 Hz publish rate
		lastPublished: time.Now(),
	}, nil
}

// PublishAdjacencyMatrix publishes the network topology to Redis
func (rp *RedisPublisher) PublishAdjacencyMatrix(snapshot *StateSnapshot) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// Rate limit to 10 Hz
	if time.Since(rp.lastPublished) < rp.publishRate {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Serialize snapshot
	data, err := json.Marshal(snapshot.Entries)
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}

	// Publish as Redis Stream entry for history
	entry := map[string]interface{}{
		"data":      string(data),
		"timestamp": snapshot.Timestamp.Unix(),
	}

	if err := rp.client.XAdd(ctx, &redis.XAddArgs{
		Stream: rp.streamKey,
		Values: entry,
	}).Err(); err != nil {
		log.Printf("Error publishing to stream: %v", err)
		// Don't fail entirely if stream publish fails
	}

	// Also publish to a key for direct access (overwrite previous)
	if err := rp.client.Set(ctx, rp.snapshotKey, string(data), 0).Err(); err != nil {
		return fmt.Errorf("publish snapshot: %w", err)
	}

	// Publish to pub/sub channel for real-time subscribers
	if err := rp.client.Publish(ctx, "topology-updates", string(data)).Err(); err != nil {
		log.Printf("Error publishing to pubsub: %v", err)
	}

	rp.lastPublished = time.Now()
	return nil
}

// SubscribeToAdjacencyUpdates returns a channel of topology snapshots
func (rp *RedisPublisher) SubscribeToAdjacencyUpdates(ctx context.Context) <-chan *StateSnapshot {
	updates := make(chan *StateSnapshot, 1)

	go func() {
		defer close(updates)
		pubsub := rp.client.Subscribe(ctx, "topology-updates")
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if msg == nil {
					return
				}

				var entries map[int]AdjacencyEntry
				if err := json.Unmarshal([]byte(msg.Payload), &entries); err != nil {
					log.Printf("Error unmarshaling snapshot: %v", err)
					continue
				}

				snapshot := &StateSnapshot{
					Timestamp: time.Now(),
					Entries:   entries,
				}

				select {
				case updates <- snapshot:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return updates
}

// Close closes the Redis connection
func (rp *RedisPublisher) Close() error {
	return rp.client.Close()
}

// Health checks Redis connectivity
func (rp *RedisPublisher) Health(ctx context.Context) error {
	return rp.client.Ping(ctx).Err()
}
