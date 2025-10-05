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
)

const (
	RMQTopicMessage        = "goim_publish_message"
	RMQConsumeGroupMessage = "cg_publish_message"
)

const (
	UserIconURI = "default_icon/user_default_icon.png"
)
