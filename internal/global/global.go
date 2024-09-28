package global

const (
	HeaderContentTypeAppJson = "application/json"

	EndpointHealth          = "/healthz"
	EndpointInfo            = "/info"
	EndpointRequestTransfer = "/transfer"
	EndpointRequestObject   = "/request/object"
	EndpointRequestDatabase = "/request/database"
)

const (
	LogCommonFieldKeyService   = "service"
	LogCommonFieldKeyInstance  = "instance"
	LogCommonFieldKeyComponent = "component"

	LogCommonExtraFieldKeyError         = "error"
	LogCommonExtraFieldKeyObject        = "object"
	LogCommonExtraFieldKeyBackendObject = "backend_object"
	LogCommonExtraFieldKeyRequestCount  = "request_count"
	LogCommonExtraFieldKeyRequestId     = "request_id"
	LogCommonExtraFieldKeyRequestList   = "request_list"
	LogCommonExtraFieldKeyThreads       = "threads"
	LogCommonExtraFieldKeyPoolLen       = "pool_len"

	LogFieldValueNone = "none"
)

var (
	ServerState ServerReadyT

	LogCommonFields = map[string]any{
		LogCommonFieldKeyService:   "bot",
		LogCommonFieldKeyInstance:  LogFieldValueNone,
		LogCommonFieldKeyComponent: LogFieldValueNone,
	}

	LogCommonExtraFields = map[string]any{
		LogCommonExtraFieldKeyError:         LogFieldValueNone,
		LogCommonExtraFieldKeyObject:        LogFieldValueNone,
		LogCommonExtraFieldKeyBackendObject: LogFieldValueNone,
		LogCommonExtraFieldKeyRequestCount:  LogFieldValueNone,
		LogCommonExtraFieldKeyRequestId:     LogFieldValueNone,
		LogCommonExtraFieldKeyRequestList:   LogFieldValueNone,
		LogCommonExtraFieldKeyThreads:       LogFieldValueNone,
		LogCommonExtraFieldKeyPoolLen:       LogFieldValueNone,
	}
)
