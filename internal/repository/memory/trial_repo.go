package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexoone/nexo-one/internal/trial"
)

// TrialRepo implementa trial.TrialRepository in-memory.
type TrialRepo struct {
	mu       sync.RWMutex
	controls map[string]*trial.TrialControl // keyed by phoneHash
	codes    map[string]codeEntry           // phoneHash -> code
	attempts map[string]int                 // phoneHash -> count
}

type codeEntry struct {
	code      string
	expiresAt time.Time
}

func NewTrialRepo() *TrialRepo {
	return &TrialRepo{
		controls: make(map[string]*trial.TrialControl),
		codes:    make(map[string]codeEntry),
		attempts: make(map[string]int),
	}
}

func (r *TrialRepo) GetByPhoneHash(_ context.Context, hash string) (*trial.TrialControl, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tc, ok := r.controls[hash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return tc, nil
}

func (r *TrialRepo) GetByDeviceHash(_ context.Context, hash string, since time.Time) ([]*trial.TrialControl, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*trial.TrialControl
	for _, tc := range r.controls {
		if tc.DeviceHash == hash && tc.CreatedAt.After(since) {
			result = append(result, tc)
		}
	}
	return result, nil
}

func (r *TrialRepo) Create(_ context.Context, tc *trial.TrialControl) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if tc.ID == "" {
		tc.ID = fmt.Sprintf("trial-%d", time.Now().UnixNano())
	}
	tc.CreatedAt = time.Now()
	r.controls[tc.PhoneHash] = tc
	return nil
}

func (r *TrialRepo) Update(_ context.Context, tc *trial.TrialControl) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.controls[tc.PhoneHash] = tc
	return nil
}

func (r *TrialRepo) SaveCode(_ context.Context, phoneHash, code string, ttl time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.codes[phoneHash] = codeEntry{code: code, expiresAt: time.Now().Add(ttl)}
	return nil
}

func (r *TrialRepo) GetCode(_ context.Context, phoneHash string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entry, ok := r.codes[phoneHash]
	if !ok || time.Now().After(entry.expiresAt) {
		return "", fmt.Errorf("code expired or not found")
	}
	return entry.code, nil
}

func (r *TrialRepo) IncrementAttempts(_ context.Context, phoneHash string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attempts[phoneHash]++
	return r.attempts[phoneHash], nil
}
