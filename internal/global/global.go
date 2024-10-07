package global

const (
	HeaderContentTypeAppJson = "application/json"

	EndpointHealthz         = "/healthz"
	EndpointInfo            = "/info"
	EndpointRequestTransfer = "/transfer"
	EndpointRequestObject   = "/request/object"
	EndpointRequestDatabase = "/request/database"
)

const (
	LogFieldKeyCommonService   = "service"
	LogFieldKeyCommonInstance  = "instance"
	LogFieldKeyCommonComponent = "component"

	LogFieldKeyExtraError              = "error"
	LogFieldKeyExtraObject             = "object"
	LogFieldKeyExtraBackendObject      = "backend_object"
	LogFieldKeyExtraRequestId          = "request_id"
	LogFieldKeyExtraRequestList        = "request_list"
	LogFieldKeyExtraActiveRequestCount = "active_request_count"
	LogFieldKeyExtraActiveThreadCount  = "active_thread_count"
	LogFieldKeyExtraCurrentPoolLength  = "current_pool_length"

	LogFieldValueDefault                 = "none"
	LogFieldValueService                 = "bot"
	LogFieldValueComponentAPIService     = "APIService"
	LogFieldValueComponentObjectWorker   = "ObjectWorker"
	LogFieldValueComponentDatabaseWorker = "DatabaseWorker"
	LogFieldValueComponentHashringWorker = "HashringWorker"
)

var (
	ServerState ServerReadyT
)

func GetLogCommonFields() map[string]any {
	return map[string]any{
		LogFieldKeyCommonService:   "bot",
		LogFieldKeyCommonInstance:  LogFieldValueDefault,
		LogFieldKeyCommonComponent: LogFieldValueDefault,
	}
}

func GetLogExtraFieldsAPI() map[string]any {
	return map[string]any{
		LogFieldKeyExtraError:  LogFieldValueDefault,
		LogFieldKeyExtraObject: LogFieldValueDefault,
	}
}

func GetLogExtraFieldsObjectWorker() map[string]any {
	return map[string]any{
		LogFieldKeyExtraError:              LogFieldValueDefault,
		LogFieldKeyExtraObject:             LogFieldValueDefault,
		LogFieldKeyExtraBackendObject:      LogFieldValueDefault,
		LogFieldKeyExtraActiveRequestCount: LogFieldValueDefault,
		LogFieldKeyExtraActiveThreadCount:  LogFieldValueDefault,
		LogFieldKeyExtraCurrentPoolLength:  LogFieldValueDefault,
	}
}

func GetLogExtraFieldsDatabaseWorker() map[string]any {
	return map[string]any{
		LogFieldKeyExtraError:              LogFieldValueDefault,
		LogFieldKeyExtraRequestId:          LogFieldValueDefault,
		LogFieldKeyExtraRequestList:        LogFieldValueDefault,
		LogFieldKeyExtraActiveRequestCount: LogFieldValueDefault,
		LogFieldKeyExtraActiveThreadCount:  LogFieldValueDefault,
		LogFieldKeyExtraCurrentPoolLength:  LogFieldValueDefault,
	}
}

func GetLogExtraFieldsHashringWorker() map[string]any {
	return map[string]any{
		LogFieldKeyExtraError:       LogFieldValueDefault,
		LogFieldKeyExtraRequestId:   LogFieldValueDefault,
		LogFieldKeyExtraRequestList: LogFieldValueDefault,
	}
}
