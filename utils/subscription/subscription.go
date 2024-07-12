package subscription

import (
	"fmt"
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
	fmt.Printf("Added subscriber: %s\n", id)
}

func (m *Manager) RemoveSubscriber(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscribers, id)
	fmt.Printf("Removed subscriber: %s\n", id)
}

func (m *Manager) NotifySubscriberByID(template *model.TemplatePlayWithFriend, id string) error {
	if !m.SubscriberExists(id) {
		return fmt.Errorf("subscriber with ID %s does not exist", id)
	}

	fmt.Printf("Notifying subscriber with template ID: %s, user ID: %s\n", template.ID, id)
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Println("39")
	ch, exists := m.subscribers[id]
	if !exists {
		fmt.Println("41")
		fmt.Println("42", id)
		return fmt.Errorf("subscriber with ID %s does not exist", id)
	}

	fmt.Println(ch)

	fmt.Printf("Notifying subscriber by ID: %s\n", id)
	ch <- template

	return nil
}

func (m *Manager) SubscriberCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.subscribers)
}

func (m *Manager) SubscriberExists(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.subscribers[id]
	return exists
}

// func (m *Manager) NotifySubscribers(template *model.TemplatePlayWithFriend) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	for id, ch := range m.subscribers {
// 		fmt.Printf("Notifying subscriber: %s\n", id)
// 		ch <- template
// 	}
// }
