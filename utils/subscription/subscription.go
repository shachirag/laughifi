package subscription

import (
	"context"
	"fmt"
	"laughifi/database"
	"laughifi/entity"
	"laughifi/graph/model"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Manager struct {
	db            *database.DB
	subscribers   map[string]chan *model.TemplatePlayWithFriend
	subscribersMu sync.Mutex
}

func NewManager(db *database.DB) *Manager {
	return &Manager{
		db:            db,
		subscribers:   make(map[string]chan *model.TemplatePlayWithFriend),
		subscribersMu: sync.Mutex{},
	}
}

func (m *Manager) AddSubscriber(userID string, ch chan *model.TemplatePlayWithFriend) error {
	m.subscribersMu.Lock()
	defer m.subscribersMu.Unlock()

	if _, exists := m.subscribers[userID]; exists {
		return fmt.Errorf("subscriber with ID %s already exists", userID)
	}

	m.subscribers[userID] = ch
	fmt.Printf("Added subscriber: %s\n", userID)

	if err := m.addToDatabase(userID); err != nil {
		return err
	}

	return nil
}

func (m *Manager) RemoveSubscriber(userID string) error {
	m.subscribersMu.Lock()
	defer m.subscribersMu.Unlock()

	if _, exists := m.subscribers[userID]; !exists {
		return fmt.Errorf("subscriber with ID %s does not exist", userID)
	}

	delete(m.subscribers, userID)
	fmt.Printf("Removed subscriber: %s\n", userID)

	if err := m.removeFromDatabase(userID); err != nil {
		return err
	}

	return nil
}

func (m *Manager) NotifySubscriberByID(template *model.TemplatePlayWithFriend, userID string) error {
	m.subscribersMu.Lock()
	defer m.subscribersMu.Unlock()

	ch, exists := m.subscribers[userID]
	if !exists {
		return fmt.Errorf("subscriber channel for ID %s does not exist", userID)
	}

	ch <- template
	return nil
}

func (m *Manager) SubscriberExists(userID string) bool {
	m.subscribersMu.Lock()
	defer m.subscribersMu.Unlock()
	_, exists := m.subscribers[userID]
	return exists
}

func (m *Manager) addToDatabase(userID string) error {
	collection := m.db.GetCollection("subscriber")

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	subscriber := entity.SubscriptionEntity{
		Id:        primitive.NewObjectID(),
		UserId:    userObjID,
		CreatedAt: time.Now().UTC(),
	}

	_, err = collection.InsertOne(context.TODO(), subscriber)
	if err != nil {
		return fmt.Errorf("failed to add subscriber to database: %v", err)
	}

	return nil
}

func (m *Manager) removeFromDatabase(userID string) error {
	collection := m.db.GetCollection("subscriber")

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	filter := bson.M{"userId": userObjID}

	_, err = collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("failed to remove subscriber from database: %v", err)
	}

	return nil
}

// func (m *Manager) SubscriberCount() int {
// 	m.subscribersMu.Lock()
// 	defer m.subscribersMu.Unlock()
// 	return len(m.subscribers)
// }
