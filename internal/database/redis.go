package database

import (
	"context"
	"time"

	"Afaisal94/first-app/internal/config"
	"Afaisal94/first-app/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// ConnectRedis membuka koneksi ke Redis dari config, melakukan ping untuk
// memastikan koneksi hidup, lalu mengembalikan *redis.Client.
// Client dikembalikan apa adanya (bukan interface) agar caller bisa memakai
// seluruh perintah Redis dan menutupnya lewat Close() saat shutdown.
func ConnectRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),   // host:port, mis. localhost:6379
		Password: cfg.RedisPassword, // kosong berarti tanpa auth
		DB:       0,                 // database default
	})

	// Batasi waktu ping agar startup tidak menggantung bila Redis down.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Default.Fatalf("failed to connect redis: %v", err)
	}

	logger.Default.Infof("redis connected: %s", cfg.RedisAddr())
	return client
}
