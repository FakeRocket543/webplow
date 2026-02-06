package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Token struct {
	Key       string    `json:"key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Store struct {
	path   string
	tokens map[string]*Token // key -> Token
	mu     sync.RWMutex
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path, tokens: make(map[string]*Token)}
	if data, err := os.ReadFile(path); err == nil {
		var list []*Token
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, t := range list {
			s.tokens[t.Key] = t
		}
	}
	return s, nil
}

func (s *Store) Valid(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if t, ok := s.tokens[key]; ok {
		return t.Name, true
	}
	return "", false
}

func (s *Store) Add(name string) (*Token, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	t := &Token{
		Key:       hex.EncodeToString(b),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.tokens[t.Key] = t
	s.mu.Unlock()
	return t, s.save()
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	delete(s.tokens, key)
	s.mu.Unlock()
	return s.save()
}

func (s *Store) List() []*Token {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Token, 0, len(s.tokens))
	for _, t := range s.tokens {
		list = append(list, t)
	}
	return list
}

func (s *Store) save() error {
	s.mu.RLock()
	list := make([]*Token, 0, len(s.tokens))
	for _, t := range s.tokens {
		list = append(list, t)
	}
	s.mu.RUnlock()
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0600)
}
