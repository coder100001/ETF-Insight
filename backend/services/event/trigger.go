package event

import (
	"context"
	"sync"
	"time"

	"etf-insight/services/etf"
	"etf-insight/utils"

	"gorm.io/gorm"
)

// EventType 事件类型
type EventType string

const (
	// EventETFHoldingsUpdated ETF持仓数据更新事件
	EventETFHoldingsUpdated EventType = "etf_holdings_updated"
	// EventETFDataUpdated ETF行情数据更新事件
	EventETFDataUpdated EventType = "etf_data_updated"
	// EventCacheInvalidation 缓存失效事件
	EventCacheInvalidation EventType = "cache_invalidation"
)

// Event 事件结构
type Event struct {
	Type      EventType              `json:"type"`
	Source    string                 `json:"source"` // 触发源：api, sync, manual, etc.
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// EventHandler 事件处理器接口
type EventHandler interface {
	Handle(ctx context.Context, event *Event) error
	CanHandle(eventType EventType) bool
}

// EventBus 事件总线
type EventBus struct {
	handlers map[EventType][]EventHandler
	db       *gorm.DB
	cache    *etf.OverlapCacheService
	mu       sync.RWMutex
}

// NewEventBus 创建事件总线
func NewEventBus(db *gorm.DB) *EventBus {
	cache := etf.NewOverlapCacheService(db)
	bus := &EventBus{
		handlers: make(map[EventType][]EventHandler),
		db:       db,
		cache:    cache,
	}

	bus.registerDefaultHandlers()
	return bus
}

// registerDefaultHandlers 注册默认的事件处理器
func (bus *EventBus) registerDefaultHandlers() {
	bus.RegisterHandler(&CacheInvalidationHandler{cache: bus.cache})
}

// RegisterHandler 注册事件处理器
func (bus *EventBus) RegisterHandler(handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	eventTypes := []EventType{
		EventETFHoldingsUpdated,
		EventETFDataUpdated,
		EventCacheInvalidation,
	}

	for _, eventType := range eventTypes {
		if handler.CanHandle(eventType) {
			bus.handlers[eventType] = append(bus.handlers[eventType], handler)
		}
	}
}

// Publish 发布事件
func (bus *EventBus) Publish(ctx context.Context, event *Event) error {
	if event == nil {
		return nil
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	utils.Info("Event published",
		"type", string(event.Type),
		"source", event.Source,
		"timestamp", event.Timestamp,
	)

	bus.mu.RLock()
	handlers := bus.handlers[event.Type]
	bus.mu.RUnlock()

	var wg sync.WaitGroup
	errChan := make(chan error, len(handlers))

	for _, handler := range handlers {
		wg.Add(1)
		go func(h EventHandler) {
			defer wg.Done()
			if err := h.Handle(ctx, event); err != nil {
				errChan <- err
				utils.Error("Event handler failed", err,
					"handler", h.CanHandle(event.Type),
					"event_type", string(event.Type),
				)
			}
		}(handler)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return errors[0]
	}

	return nil
}

// PublishETFHoldingsUpdated 发布ETF持仓更新事件
func (bus *EventBus) PublishETFHoldingsUpdated(ctx context.Context, symbol string, source string) error {
	return bus.Publish(ctx, &Event{
		Type:   EventETFHoldingsUpdated,
		Source: source,
		Payload: map[string]interface{}{
			"symbol": symbol,
		},
	})
}

// PublishETFDataUpdated 发布ETF数据更新事件
func (bus *EventBus) PublishETFDataUpdated(ctx context.Context, symbol string, source string) error {
	return bus.Publish(ctx, &Event{
		Type:   EventETFDataUpdated,
		Source: source,
		Payload: map[string]interface{}{
			"symbol": symbol,
		},
	})
}

// CacheInvalidationHandler 缓存失效处理器
type CacheInvalidationHandler struct {
	cache *etf.OverlapCacheService
}

// Handle 处理缓存失效事件
func (h *CacheInvalidationHandler) Handle(ctx context.Context, event *Event) error {
	switch event.Type {
	case EventETFHoldingsUpdated:
		symbol, ok := event.Payload["symbol"].(string)
		if !ok || symbol == "" {
			return nil
		}

		if err := h.cache.InvalidateOverlapCache(ctx, symbol); err != nil {
			utils.Warn("Failed to invalidate cache for ETF", "symbol", symbol, "error", err)
		} else {
			utils.Info("Cache invalidated due to ETF holdings update", "symbol", symbol)
		}

	case EventETFDataUpdated:
		symbol, _ := event.Payload["symbol"].(string)
		if symbol == "" {
			if err := h.cache.InvalidateAllOverlapCache(ctx); err != nil {
				utils.Warn("Failed to invalidate all cache", "error", err)
			} else {
				utils.Info("All cache invalidated due to ETF data update")
			}
		} else {
			if err := h.cache.InvalidateOverlapCache(ctx, symbol); err != nil {
				utils.Warn("Failed to invalidate cache for ETF", "symbol", symbol, "error", err)
			}
		}
	}

	return nil
}

// CanHandle 判断是否可以处理该类型事件
func (h *CacheInvalidationHandler) CanHandle(eventType EventType) bool {
	return eventType == EventETFHoldingsUpdated || eventType == EventETFDataUpdated
}

// GlobalEventBus 全局事件总线实例
var GlobalEventBus *EventBus

// InitEventBus 初始化全局事件总线
func InitEventBus(db *gorm.DB) {
	GlobalEventBus = NewEventBus(db)
	utils.Info("Event bus initialized with default handlers")
}
