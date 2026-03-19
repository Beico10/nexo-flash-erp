package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/nexoone/nexo-one/internal/ai"
)

type AIRepo struct {
	mu          sync.RWMutex
	suggestions map[string]*ai.Suggestion
}

func NewAIRepo() *AIRepo {
	return &AIRepo{suggestions: make(map[string]*ai.Suggestion)}
}

func (r *AIRepo) CreatePending(_ context.Context, s *ai.Suggestion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	r.suggestions[s.ID] = s
	return nil
}

func (r *AIRepo) GetPending(_ context.Context, tenantID string) ([]*ai.Suggestion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*ai.Suggestion
	for _, s := range r.suggestions {
		if s.TenantID == tenantID && s.Status == "pending" {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *AIRepo) Approve(_ context.Context, suggestionID, approvedByUserID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.suggestions[suggestionID]
	if !ok {
		return fmt.Errorf("sugestao %s nao encontrada", suggestionID)
	}
	s.Status = "approved"
	return nil
}

func (r *AIRepo) Reject(_ context.Context, suggestionID, userID, reason string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.suggestions[suggestionID]
	if !ok {
		return fmt.Errorf("sugestao %s nao encontrada", suggestionID)
	}
	s.Status = "rejected"
	return nil
}
