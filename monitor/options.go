package monitor

// NewMetricsOptions is the options for NewMetricsMonitor
// it can be used to set the name of metrics or the log function
// defaultOptions will be used if field not set
type NewMetricsOptions struct {
	PoolEventMetricsCounterName string    // PoolEventMetricsCounterName name of event metric. default: mongo_pool_event_total
	ConnectionMetricsGaugeName  string    // ConnectionMetricsGaugeName name of connection metric. default: mongo_pool_connection
	CommandDurationBucketName   string    // CommandDurationBucketName name of command duration bucket metric. default: mongo_command_duration_bucket
	commandDurationBucket       []float64 // commandDurationBucket

	LogInfoFunc func(format string, args ...any) // LogInfoFunc
	LogWarnFunc func(format string, args ...any) // LogWarnFunc
}

func mergeOptions(src *NewMetricsOptions, target *NewMetricsOptions) *NewMetricsOptions {
	opt := &NewMetricsOptions{
		PoolEventMetricsCounterName: src.PoolEventMetricsCounterName,
		ConnectionMetricsGaugeName:  src.ConnectionMetricsGaugeName,
		CommandDurationBucketName:   src.CommandDurationBucketName,
		commandDurationBucket:       src.commandDurationBucket,
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
	if len(target.commandDurationBucket) > 0 {
		opt.commandDurationBucket = target.commandDurationBucket
	}
	if target.LogInfoFunc != nil {
		opt.LogInfoFunc = target.LogInfoFunc
	}
	if target.LogWarnFunc != nil {
		opt.LogWarnFunc = target.LogWarnFunc
	}
	return opt
}
