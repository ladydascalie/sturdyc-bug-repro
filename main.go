package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/redis/go-redis/v9"
	"github.com/viccon/sturdyc"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	var keys []string
	for range 4 {
		keys = append(keys, ulid.Make().String())
	}

	c := NewCache("lb", client, WithTTL(10*time.Second))

	// try twice to access each key.
	for range 2 {
		for _, k := range keys {
			lb, err := sturdyc.GetOrFetch(context.Background(), c, k, func(ctx context.Context) (LeaderboardRecord, error) {
				return getLeaderboard(false)
			})
			fmt.Printf("lb: %#+v\n", lb)
			if err != nil {
				fmt.Printf("err: %v\n", err)
			}
			fmt.Println("-----------------------------")
			time.Sleep(250 * time.Millisecond)
		}
	}

	// and try a key that will fail.
	lb, err := sturdyc.GetOrFetch(context.Background(), c, "", func(ctx context.Context) (LeaderboardRecord, error) {
		return getLeaderboard(true)
	})
	fmt.Printf("lb: %#+v\n", lb)
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	fmt.Println("-----------------------------")
	time.Sleep(time.Millisecond)
}

var ErrLeaderboardNotFound = errors.New("leaderboard not found")

func getLeaderboard(failed bool) (LeaderboardRecord, error) {
	if failed {
		return LeaderboardRecord{}, ErrLeaderboardNotFound
	}
	// if rand.IntN(3)%2 == 0 {
	// 	return LeaderboardRecord{}, sql.ErrNoRows
	// } else {
	return LeaderboardRecord{
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		DeletedAt:              nil,
		Key:                    "some-key",
		DirectionMethod:        Descending,
		Name:                   "some-name",
		Type:                   Generic,
		ID:                     1,
		GameID:                 1,
		EnableGameAPIWrites:    false,
		OverwriteScoreOnSubmit: false,
		HasMetadata:            false,
		ULID:                   NullULID{ULID: ulid.Make(), Valid: true},
	}, nil
	// }
}

type (
	LeaderboardRecord struct {
		CreatedAt              time.Time  `json:"created_at"`
		UpdatedAt              time.Time  `json:"updated_at"`
		DeletedAt              *time.Time `gorm:"index" json:"-"`
		Key                    string     `json:"key"`
		DirectionMethod        Direction  `json:"direction_method"`
		Name                   string     `json:"name"`
		Type                   Type       `json:"type"`
		ID                     uint64     `gorm:"primarykey" json:"id"`
		GameID                 uint64     `gorm:"index" json:"game_id"`
		EnableGameAPIWrites    bool       `json:"enable_game_api_writes"`
		OverwriteScoreOnSubmit bool       `json:"overwrite_score_on_submit"`
		HasMetadata            bool       `json:"has_metadata"`
		ULID                   NullULID   `gorm:"column:ulid" json:"ulid"`
	}
)
