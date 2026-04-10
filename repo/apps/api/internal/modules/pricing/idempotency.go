package pricing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CheckIdempotencyKey(
	ctx context.Context,
	pool *pgxpool.Pool,
	actorID, route, key, requestHash string,
) (bool, *IdempotencyResponse, error) {
	var responseCode int
	var responseBody []byte
	var storedHash string

	err := pool.QueryRow(ctx,
		`SELECT request_hash, response_code, response_body_json
		 FROM idempotency_keys
		 WHERE actor_id = $1 AND route = $2 AND key = $3 AND response_code IS NOT NULL`,
		actorID, route, key,
	).Scan(&storedHash, &responseCode, &responseBody)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("check idempotency key: %w", err)
	}

	if storedHash != requestHash {
		return true, nil, fmt.Errorf("idempotency key reused with different request body")
	}

	return true, &IdempotencyResponse{
		ResponseCode: responseCode,
		ResponseBody: json.RawMessage(responseBody),
	}, nil
}

func LockIdempotencyKey(
	ctx context.Context,
	pool *pgxpool.Pool,
	actorID, route, key, requestHash string,
) error {
	id := uuid.New().String()
	_, err := pool.Exec(ctx,
		`INSERT INTO idempotency_keys (id, actor_id, key, route, request_hash, locked_at, created_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), NOW() + INTERVAL '24 hours')
		 ON CONFLICT (actor_id, route, key) DO NOTHING`,
		id, actorID, key, route, requestHash,
	)
	if err != nil {
		return fmt.Errorf("lock idempotency key: %w", err)
	}
	return nil
}

func CompleteIdempotencyKey(
	ctx context.Context,
	pool *pgxpool.Pool,
	actorID, route, key string,
	responseCode int,
	responseBody []byte,
) error {
	_, err := pool.Exec(ctx,
		`UPDATE idempotency_keys
		 SET response_code = $1, response_body_json = $2
		 WHERE actor_id = $3 AND route = $4 AND key = $5`,
		responseCode, responseBody, actorID, route, key,
	)
	if err != nil {
		return fmt.Errorf("complete idempotency key: %w", err)
	}
	return nil
}
