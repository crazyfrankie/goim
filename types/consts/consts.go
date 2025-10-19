package consts

const (
	JWTSignAlgo   = "JWT_SIGN_ALGO"
	JWTSecretKey  = "JWT_SECRET_KEY"
	JWTPublicKey  = "JWT_PUBLIC_KEY"
	MinIOAK       = "MINIO_AK"
	MinIOSK       = "MINIO_SK"
	MinIOEndpoint = "MINIO_ENDPOINT"
	StorageBucket = "STORAGE_BUCKET"
	MQTypeKey     = "MQ_TYPE"
	RMQAccessKey  = "RMQ_ACCESS_KEY"
	RMQSecretKey  = "RMQ_SECRET_KEY"
	MQServer      = "MQ_SERVER"
	DiscoveryType = "DISCOVERY_TYPE"
)

const (
	RMQTopicMessage        = "goim_push_message"
	RMQConsumeGroupMessage = "cg_push_message"
)

const (
	UserIconURI = "default_icon/user_default_icon.png"
)

const (
	UserServiceName    = "goim-rpc-user"
	AuthServiceName    = "goim-rpc-auth"
	MessageServiceName = "goim-rpc-msg"
	PushServiceName    = "goim-rpc-push"

	UserApiName    = "goim-api-user"
	MessageApiName = "goim-api-msg"

	MsgGatewayName = "goim-msggateway"

	UserServiceVer    = "v0.0.1"
	MessageServiceVer = "v0.0.1"
	AuthServiceVer    = "v0.0.1"
	PushServiceVer    = "v0.0.1"
	UserApiVer        = "v0.0.1"
	MessageApiVer     = "v0.0.1"
	MsgGatewayVer     = "v0.0.1"
)
