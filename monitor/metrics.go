package monitor

import (
	"context"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/event"
)

type MongoMonitor struct {
	connectionPoolEventMetrics *prometheus.CounterVec
	connectionPoolMetrics      *prometheus.GaugeVec
	commandDurationBucket      *prometheus.HistogramVec
	poolMonitorFuncChain       []func(evt *event.PoolEvent)
	debugLog                   bool
	logInfoFunc                func(format string, args ...any)
	logWarnFunc                func(format string, args ...any)
	commandStore               sync.Map
	monitorSucceededFuncChain  []func(context.Context, *event.CommandSucceededEvent)
	monitorStartedFuncChain    []func(context.Context, *event.CommandStartedEvent)
	monitorFailedFuncChain     []func(context.Context, *event.CommandFailedEvent)
}

// NewMongoMonitor create a new mongo monitor, it can generator monitor func for mongo-go-driver monitor
func NewMongoMonitor(newMetricsOpt *NewMetricsOptions) *MongoMonitor {
	opt := mergeOptions(DefaultMetricsOptions, newMetricsOpt)
	monitor := &MongoMonitor{
		connectionPoolEventMetrics: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: opt.PoolEventMetricsCounterName, Help: "MongoDB connection pool event counter"}, []string{"type", "reason"}),
		connectionPoolMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: opt.ConnectionMetricsGaugeName, Help: "MongoDB connection pool state gauge"}, []string{"type"}),
		commandDurationBucket: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: opt.CommandDurationBucketName, Help: "MongoDB every command duration histogram", Buckets: opt.CommandDurationBucket}, []string{"type"}),
	}
	monitor.logInfoFunc = opt.LogInfoFunc
	monitor.logWarnFunc = opt.LogWarnFunc
	monitor.debugLog = opt.DebugLog
	monitor.poolMonitorFuncChain = append(monitor.poolMonitorFuncChain, monitor.poolEventMetricsFunc)
	monitor.monitorSucceededFuncChain = append(monitor.monitorSucceededFuncChain, monitor.commandSucceededFunc)
	monitor.monitorStartedFuncChain = append(monitor.monitorStartedFuncChain, monitor.commandStartedFunc)
	monitor.monitorFailedFuncChain = append(monitor.monitorFailedFuncChain, monitor.commandFailedFunc)
	return monitor
}

// AddPoolMonitorFunc add a poolMonitorFunc to the poolMonitorFuncChain.
// when mongo pool monitor func is triggered, custom func will also be executed after default metrics func.
func (m *MongoMonitor) AddPoolMonitorFunc(poolEvent func(evt *event.PoolEvent)) {
	m.poolMonitorFuncChain = append(m.poolMonitorFuncChain, poolEvent)
}

// AddCommandMonitorSucceededFunc add commandSucceededFunc to the monitorSucceededFuncChain.
// when mongo command succeeded monitor func is triggered, custom func will also be executed after default metrics func.
func (m *MongoMonitor) AddCommandMonitorSucceededFunc(commandFunc func(context.Context, *event.CommandSucceededEvent)) {
	m.monitorSucceededFuncChain = append(m.monitorSucceededFuncChain, commandFunc)
}

// AddCommandMonitorStartedFunc add commandStartedFunc to the monitorStartedFuncChain.
// when mongo command started monitor func is triggered, custom func will also be executed after default metrics func.
func (m *MongoMonitor) AddCommandMonitorStartedFunc(commandFunc func(context.Context, *event.CommandStartedEvent)) {
	m.monitorStartedFuncChain = append(m.monitorStartedFuncChain, commandFunc)
}

// AddCommandMonitorFailedFunc add commandFailedFunc to the monitorFailedFuncChain.
// when mongo command failed monitor func is triggered, custom func will also be executed after default metrics func.
func (m *MongoMonitor) AddCommandMonitorFailedFunc(commandFunc func(context.Context, *event.CommandFailedEvent)) {
	m.monitorFailedFuncChain = append(m.monitorFailedFuncChain, commandFunc)
}

// GetPoolMonitor return a poolMonitor for mongo-go-driver monitor
func (m *MongoMonitor) GetPoolMonitor() *event.PoolMonitor {
	return &event.PoolMonitor{
		Event: m.poolEventMonitor,
	}
}

// RegistryMetrics register all metrics to prometheus registry.
func (m *MongoMonitor) RegistryMetrics(registry *prometheus.Registry) error {
	if err := registry.Register(m.connectionPoolEventMetrics); err != nil {
		return fmt.Errorf("register connectionPoolEventMetrics failed: %w", err)
	}
	if err := registry.Register(m.connectionPoolMetrics); err != nil {
		return fmt.Errorf("register connectionPoolMetrics failed: %w", err)
	}
	if err := registry.Register(m.commandDurationBucket); err != nil {
		return fmt.Errorf("register CommandDurationBucket failed: %w", err)
	}
	return nil
}

func (m *MongoMonitor) poolEventMonitor(evt *event.PoolEvent) {
	for _, monitor := range m.poolMonitorFuncChain {
		monitor(evt)
	}
}

// GetCommandMonitor return a commandMonitor for mongo-go-driver monitor
func (m *MongoMonitor) GetCommandMonitor() *event.CommandMonitor {
	return &event.CommandMonitor{
		Started:   m.commandStarted,
		Succeeded: m.commandSucceeded,
		Failed:    m.commandFailed,
	}
}

func (m *MongoMonitor) commandStarted(ctx context.Context, evt *event.CommandStartedEvent) {
	for _, monitor := range m.monitorStartedFuncChain {
		monitor(ctx, evt)
	}
}
func (m *MongoMonitor) commandSucceeded(ctx context.Context, succeededEvent *event.CommandSucceededEvent) {
	for _, monitor := range m.monitorSucceededFuncChain {
		monitor(ctx, succeededEvent)
	}
}
func (m *MongoMonitor) commandFailed(ctx context.Context, failedEvent *event.CommandFailedEvent) {
	for _, monitor := range m.monitorFailedFuncChain {
		monitor(ctx, failedEvent)
	}
}
