package locker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

const SeatLockTTL = 5 * time.Minute

type SeatLocker interface {
	Lock(ctx context.Context, showtimeID string, seatNumbers []string) (token string, locked bool, err error)
	Unlock(ctx context.Context, showtimeID string, seatNumbers []string, token string) error
	IsOwner(ctx context.Context, showtimeID string, seatNumbers []string, token string) (bool, error)
}

type BookingSeatLocker struct {
	client *redis.Client
}

func NewBookingSeatLocker(client *redis.Client) *BookingSeatLocker {
	return &BookingSeatLocker{client: client}
}

func (l *BookingSeatLocker) Lock(
	ctx context.Context,
	showtimeID string,
	seatNumbers []string,
) (string, bool, error) {
	token, err := newLockToken()
	if err != nil {
		return "", false, fmt.Errorf("create seat lock token: %w", err)
	}
	orderedSeatNumbers := append([]string(nil), seatNumbers...)
	sort.Strings(orderedSeatNumbers)
	lockedSeats := make([]string, 0, len(orderedSeatNumbers))
	for _, seatNumber := range orderedSeatNumbers {
		locked, err := l.client.SetNX(
			ctx,
			seatLockKey(showtimeID, seatNumber),
			token,
			SeatLockTTL,
		).Result()
		if err != nil {
			l.unlockAfterFailedAcquisition(showtimeID, lockedSeats, token)
			return "", false, fmt.Errorf("lock seat %s: %w", seatNumber, err)
		}
		if !locked {
			l.unlockAfterFailedAcquisition(showtimeID, lockedSeats, token)
			return "", false, nil
		}
		lockedSeats = append(lockedSeats, seatNumber)
	}

	return token, true, nil
}

func (l *BookingSeatLocker) IsOwner(
	ctx context.Context,
	showtimeID string,
	seatNumbers []string,
	token string,
) (bool, error) {
	keys := make([]string, len(seatNumbers))
	for i, seatNumber := range seatNumbers {
		keys[i] = seatLockKey(showtimeID, seatNumber)
	}
	if len(keys) == 0 {
		return false, nil
	}

	const checkOwnedLocks = `
for _, key in ipairs(KEYS) do
  if redis.call("GET", key) ~= ARGV[1] then
    return 0
  end
end
return 1`

	owned, err := l.client.Eval(ctx, checkOwnedLocks, keys, token).Int()
	if err != nil {
		return false, fmt.Errorf("check seat lock ownership: %w", err)
	}
	return owned == 1, nil
}

func (l *BookingSeatLocker) Unlock(
	ctx context.Context,
	showtimeID string,
	seatNumbers []string,
	token string,
) error {
	keys := make([]string, len(seatNumbers))
	for i, seatNumber := range seatNumbers {
		keys[i] = seatLockKey(showtimeID, seatNumber)
	}
	if len(keys) == 0 {
		return nil
	}

	const releaseOwnedLocks = `
for _, key in ipairs(KEYS) do
  if redis.call("GET", key) == ARGV[1] then
    redis.call("DEL", key)
  end
end
return 1`

	if err := l.client.Eval(ctx, releaseOwnedLocks, keys, token).Err(); err != nil {
		return fmt.Errorf("unlock seats: %w", err)
	}
	return nil
}

func (l *BookingSeatLocker) unlockAfterFailedAcquisition(
	showtimeID string,
	seatNumbers []string,
	token string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = l.Unlock(ctx, showtimeID, seatNumbers, token)
}

func seatLockKey(showtimeID, seatNumber string) string {
	return fmt.Sprintf("seat-lock:%s:%s", showtimeID, seatNumber)
}

func newLockToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
