package monitor

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/event"
)

func (m *MongoMonitor) poolEventMetricsFunc(evt *event.PoolEvent) {
	if m.debugLog {
		m.printEvent(evt)
	}
	m.connectionPoolEventMetrics.WithLabelValues(evt.Type, evt.Reason).Inc()
	switch evt.Type {
	case event.PoolCreated:
		m.connectionPoolMetrics.WithLabelValues("idle").Set(0)
		m.connectionPoolMetrics.WithLabelValues("active").Set(0)
		if evt.PoolOptions != nil {
			m.connectionPoolMetrics.WithLabelValues("max_pool_size").Set(float64(evt.PoolOptions.MaxPoolSize))
			m.connectionPoolMetrics.WithLabelValues("min_pool_size").Set(float64(evt.PoolOptions.MinPoolSize))
		}
	case event.PoolReady:
	case event.PoolCleared:
		m.connectionPoolMetrics.WithLabelValues("idle").Set(0)
		m.connectionPoolMetrics.WithLabelValues("active").Set(0)
	case event.PoolClosedEvent:
		m.connectionPoolMetrics.WithLabelValues("idle").Set(0)
		m.connectionPoolMetrics.WithLabelValues("active").Set(0)
	case event.ConnectionCreated:
		m.connectionPoolMetrics.WithLabelValues("idle").Inc()
	case event.ConnectionReady, event.GetStarted, event.GetFailed:
	case event.GetSucceeded:
		m.connectionPoolMetrics.WithLabelValues("active").Inc()
		m.connectionPoolMetrics.WithLabelValues("idle").Dec()
	case event.ConnectionReturned:
		m.connectionPoolMetrics.WithLabelValues("idle").Inc()
		m.connectionPoolMetrics.WithLabelValues("active").Dec()
	case event.ConnectionClosed:
		if evt.Reason == event.ReasonIdle || evt.Reason == event.ReasonStale {
			m.connectionPoolMetrics.WithLabelValues("idle").Dec()
		}
	}
}

func (m *MongoMonitor) printEvent(evt *event.PoolEvent) {
	if m.logInfoFunc == nil {
		return
	}
	if evt.PoolOptions != nil {
		m.logInfoFunc(
			"poolEvent: { \"type\": \"%s\", \"address\": \"%s\", \"connectionID\": %d, \"reason\": \"%s\", \"poolOptions:\": {"+
				"\"maxPoolSize\": %d, \"minPoolSize\": %d }, \"serviceID\": %v, \"error\": %v}",
			evt.Type, evt.Address, evt.ConnectionID, evt.Reason, evt.PoolOptions.MaxPoolSize, evt.PoolOptions.MinPoolSize, evt.ServiceID, evt.Error)
		return
	}
	m.logInfoFunc(
		"poolEvent: { \"type\": \"%s\", \"address\": \"%s\", \"connectionId\": %d, \"reason\": \"%s\", \"serviceId\": %v, \"error\": %v}",
		evt.Type, evt.Address, evt.ConnectionID, evt.Reason, evt.ServiceID, evt.Error)
}

func (m *MongoMonitor) commandStartedFunc(_ context.Context, evt *event.CommandStartedEvent) {
	m.commandStore.Store(evt.RequestID, time.Now())
	if m.logInfoFunc == nil {
		return
	}
	m.logInfoFunc("[%d] Started command [%s]: %s - %v", evt.RequestID, evt.DatabaseName, evt.CommandName, evt.Command)
}

func (m *MongoMonitor) commandSucceededFunc(_ context.Context, evt *event.CommandSucceededEvent) {
	m.commandDurationBucket.WithLabelValues("command").Observe(evt.Duration.Seconds())

	totalDuration := m.getDuration(evt.RequestID)
	if totalDuration != 0 {
		pureDuration := totalDuration - evt.Duration
		m.commandDurationBucket.WithLabelValues("total").Observe(totalDuration.Seconds())
		m.commandDurationBucket.WithLabelValues("pure").Observe(pureDuration.Seconds())
	}

	if m.logInfoFunc == nil {
		return
	}
	if totalDuration == 0 {
		m.logInfoFunc("[%d][%s] Succeeded command: %s", evt.RequestID, evt.Duration.String(), evt.CommandName)
		return
	}
	m.logInfoFunc("[%d][%s][%s] Succeeded command: %s", evt.RequestID, totalDuration.String(), evt.Duration.String(), evt.CommandName)
}

func (m *MongoMonitor) commandFailedFunc(_ context.Context, evt *event.CommandFailedEvent) {
	if m.logWarnFunc == nil {
		return
	}
	m.logWarnFunc("[%d][%s] Failed command: %s - %v", evt.RequestID, evt.Duration.String(), evt.CommandName, evt.Failure)
}

func (m *MongoMonitor) getDuration(requestID int64) time.Duration {
	startTimeAny, ok := m.commandStore.Load(requestID)
	if !ok {
		return 0
	}
	startTime, ok := startTimeAny.(time.Time)
	if !ok {
		return 0
	}
	return time.Now().Sub(startTime)
}
