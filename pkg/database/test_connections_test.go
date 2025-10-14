package database

import (
	"testing"
)

func TestPostgresConnection(t *testing.T) {
	config := PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "webuser",
		Password: "webpass123",
		DBName:   "taskboard_db",
		SSLMode:  "disable",
	}

	db, err := NewPostgresDB(config)
	if err != nil {
		t.Skipf("Skipping test - PostgreSQL not available: %v", err)
		return
	}
	defer ClosePostgresDB(db)

	t.Log("✅ PostgreSQL connection successful")
}

func TestRedisConnection(t *testing.T) {
	config := RedisConfig{
		Host:     "localhost",
		Port:     6379,
		Password: "",
		DB:       0,
	}

	client, err := NewRedisClient(config)
	if err != nil {
		t.Skipf("Skipping test - Redis not available: %v", err)
		return
	}
	defer CloseRedisClient(client)

	t.Log("✅ Redis connection successful")
}
