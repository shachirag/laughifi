package subscription

import (
	"laughifi/graph/model"
	"sync"
)

type Manager struct {
	subscribers map[string]chan *model.TemplatePlayWithFriend
	mu          sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		subscribers: make(map[string]chan *model.TemplatePlayWithFriend),
	}
}

func (m *Manager) AddSubscriber(id string, ch chan *model.TemplatePlayWithFriend) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers[id] = ch
}

func (m *Manager) RemoveSubscriber(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscribers, id)
}

func (m *Manager) NotifySubscribers(template *model.TemplatePlayWithFriend) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ch := range m.subscribers {
		ch <- template
	}
}

func (m *Manager) SubscriberCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.subscribers)
}
