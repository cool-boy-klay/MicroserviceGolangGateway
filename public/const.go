package public

const (
	ValidatorKey  = "ValidatorKey"
	TranslatorKey = "TranslatorKey"
	AdminSessionInfoKey = "AdminSessionInfoKey"

	LoadTypeHTTP = 0
	LoadTypeTCP = 1
	LoadTypeGRPC = 2

	HTTPRuleTypePrefixURL = 0
	HTTPRuleTypeDomain = 1

	RedisFlowDayKey = "flow_count_day"
	RedisFlowHourKey = "flow_count_hour"

	FlowTotal = "flow_total"
	FlowServicePrefix = "flow_service_"
	FlowAppPrefix = "flow_app_"



	JwtSignKey = "my_sign_key"
	JWtExpiresAt = 60*60 //1h
)
var (
	LoadTypeMap = map[int]string{
		LoadTypeHTTP: "HTTP",
		LoadTypeTCP:  "TCP",
		LoadTypeGRPC: "GRPC",
	}
)