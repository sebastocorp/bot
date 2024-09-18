package global

import "sync"

type ServerReadyT struct {
	mu       sync.Mutex
	api      bool
	object   bool
	database bool
	hasring  bool
}

func (s *ServerReadyT) IsReady() bool {
	s.mu.Lock()
	result := s.api && s.object && s.database && s.hasring
	s.mu.Unlock()
	return result
}

func (s *ServerReadyT) IsAPIReady() bool {
	s.mu.Lock()
	result := s.api
	s.mu.Unlock()
	return result
}

func (s *ServerReadyT) IsObjectReady() bool {
	s.mu.Lock()
	result := s.object
	s.mu.Unlock()
	return result
}

func (s *ServerReadyT) IsDatabaseReady() bool {
	s.mu.Lock()
	result := s.database
	s.mu.Unlock()
	return result
}

func (s *ServerReadyT) IsHashringReady() bool {
	s.mu.Lock()
	result := s.hasring
	s.mu.Unlock()
	return result
}

func (s *ServerReadyT) SetAPIReady() {
	s.mu.Lock()
	s.api = true
	s.mu.Unlock()
}

func (s *ServerReadyT) SetObjectReady() {
	s.mu.Lock()
	s.object = true
	s.mu.Unlock()
}

func (s *ServerReadyT) SetDatabaseReady() {
	s.mu.Lock()
	s.database = true
	s.mu.Unlock()
}

func (s *ServerReadyT) SetHashringReady() {
	s.mu.Lock()
	s.hasring = true
	s.mu.Unlock()
}
