package event

import (
	"context"
	"testing"
	"time"

	"etf-insight/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	db.AutoMigrate(
		&models.Asset{},
		&models.ETFHolding{},
		&models.ETFOverlapCache{},
		&models.CacheInvalidationLog{},
	)

	return db
}

func TestNewEventBus(t *testing.T) {
	db := setupTestDB(t)
	bus := NewEventBus(db)

	assert.NotNil(t, bus)
	assert.NotNil(t, bus.handlers)
}

func TestEventBus_Publish(t *testing.T) {
	db := setupTestDB(t)
	bus := NewEventBus(db)

	event := &Event{
		Type:      EventETFHoldingsUpdated,
		Source:    "test",
		Timestamp: time.Now(),
		Payload: map[string]any{
			"symbol": "QQQ",
		},
	}

	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
}

func TestEventBus_PublishWithNilEvent(t *testing.T) {
	db := setupTestDB(t)
	bus := NewEventBus(db)

	err := bus.Publish(context.Background(), nil)
	assert.NoError(t, err)
}

func TestEventBus_PublishWithTimestampAutoFill(t *testing.T) {
	db := setupTestDB(t)
	bus := NewEventBus(db)

	beforePublish := time.Now()

	event := &Event{
		Type:   EventETFHoldingsUpdated,
		Source: "test",
		Payload: map[string]any{
			"symbol": "VOO",
		},
	}

	err := bus.Publish(context.Background(), event)
	assert.NoError(t, err)
	assert.True(t, !event.Timestamp.IsZero())
	assert.True(t, event.Timestamp.After(beforePublish) || event.Timestamp.Equal(beforePublish))
}

func TestCacheInvalidationHandler_CanHandle(t *testing.T) {
	handler := &CacheInvalidationHandler{}

	assert.True(t, handler.CanHandle(EventETFHoldingsUpdated))
	assert.True(t, handler.CanHandle(EventETFDataUpdated))
	assert.False(t, handler.CanHandle("unknown_event"))
}

func TestEventBus_RegisterHandler(t *testing.T) {
	db := setupTestDB(t)
	bus := NewEventBus(db)

	initialCount := len(bus.handlers[EventETFHoldingsUpdated])

	customHandler := &CustomTestHandler{}
	bus.RegisterHandler(customHandler)

	assert.Equal(t, initialCount+1, len(bus.handlers[EventETFHoldingsUpdated]))
}

func TestGlobalEventBus(t *testing.T) {
	db := setupTestDB(t)
	InitEventBus(db)

	assert.NotNil(t, GlobalEventBus)
}

func TestEventTypeConstants(t *testing.T) {
	assert.Equal(t, EventType("etf_holdings_updated"), EventETFHoldingsUpdated)
	assert.Equal(t, EventType("etf_data_updated"), EventETFDataUpdated)
	assert.Equal(t, EventType("cache_invalidation"), EventCacheInvalidation)
}

func TestEventBus_MultipleHandlers(t *testing.T) {
	db := setupTestDB(t)
	bus := NewEventBus(db)

	handler1 := &CustomTestHandler{}
	handler2 := &AnotherTestHandler{}

	bus.RegisterHandler(handler1)
	bus.RegisterHandler(handler2)

	count := len(bus.handlers[EventETFHoldingsUpdated])
	assert.GreaterOrEqual(t, count, 2)
}

type CustomTestHandler struct{}

func (h *CustomTestHandler) Handle(ctx context.Context, event *Event) error {
	return nil
}

func (h *CustomTestHandler) CanHandle(eventType EventType) bool {
	return eventType == EventETFHoldingsUpdated
}

type AnotherTestHandler struct{}

func (h *AnotherTestHandler) Handle(ctx context.Context, event *Event) error {
	return nil
}

func (h *AnotherTestHandler) CanHandle(eventType EventType) bool {
	return eventType == EventETFHoldingsUpdated || eventType == EventETFDataUpdated
}
