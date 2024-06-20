package monitor

// NewMetricsOptions is the options for NewMetricsMonitor
// it can be used to set the name of metrics or the log function
// defaultOptions will be used if field not set
type NewMetricsOptions struct {
	PoolEventMetricsCounterName string    // PoolEventMetricsCounterName name of event metric. default: mongo_pool_event_total
	ConnectionMetricsGaugeName  string    // ConnectionMetricsGaugeName name of connection metric. default: mongo_pool_connection
	CommandDurationBucketName   string    // CommandDurationBucketName name of command duration bucket metric. default: mongo_command_duration_bucket
	CommandDurationBucket       []float64 // CommandDurationBucket bucket control of command duration.
	DebugLog                    bool      // DebugLog will print all event/command log to logInfoFunc()

	LogInfoFunc func(format string, args ...any) // LogInfoFunc
	LogWarnFunc func(format string, args ...any) // LogWarnFunc
}

var DefaultMetricsOptions = &NewMetricsOptions{
	PoolEventMetricsCounterName: "mongo_pool_event_total",
	ConnectionMetricsGaugeName:  "mongo_pool_connection",
	CommandDurationBucketName:   "mongo_command_duration",
	CommandDurationBucket:       []float64{0.0001, 0.001, 0.01, 0.05, 0.1, 0.5, 0.8, 1, 2, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 200, 300, 400, 500, 600},
	DebugLog:                    false,
	LogInfoFunc:                 nil,
	LogWarnFunc:                 nil,
}

func mergeOptions(src *NewMetricsOptions, target *NewMetricsOptions) *NewMetricsOptions {
	opt := &NewMetricsOptions{
		PoolEventMetricsCounterName: src.PoolEventMetricsCounterName,
		ConnectionMetricsGaugeName:  src.ConnectionMetricsGaugeName,
		CommandDurationBucketName:   src.CommandDurationBucketName,
		CommandDurationBucket:       src.CommandDurationBucket,
		DebugLog:                    src.DebugLog,
		LogInfoFunc:                 src.LogInfoFunc,
		LogWarnFunc:                 src.LogWarnFunc,
	}
	if len(target.PoolEventMetricsCounterName) > 0 {
		opt.PoolEventMetricsCounterName = target.PoolEventMetricsCounterName
	}
	if len(target.ConnectionMetricsGaugeName) > 0 {
		opt.ConnectionMetricsGaugeName = target.ConnectionMetricsGaugeName
	}
	if len(target.CommandDurationBucketName) > 0 {
		opt.CommandDurationBucketName = target.CommandDurationBucketName
	}
	if len(target.CommandDurationBucket) > 0 {
		opt.CommandDurationBucket = target.CommandDurationBucket
	}
	if target.LogInfoFunc != nil {
		opt.LogInfoFunc = target.LogInfoFunc
	}
	if target.LogWarnFunc != nil {
		opt.LogWarnFunc = target.LogWarnFunc
	}
	if target.DebugLog {
		opt.DebugLog = target.DebugLog
	}
	return opt
}
