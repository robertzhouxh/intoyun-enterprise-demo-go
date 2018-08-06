package define

const (
	// proto
	OP_PROTO_READY  = int32(0)
	OP_PROTO_FINISH = int32(1)

	// Operation Code 0 ~ 64
	OP_AUTH            = int32(2)
	OP_AUTH_REPLY      = int32(3)
	OP_HEARTBEAT       = int32(4)
	OP_HEARTBEAT_REPLY = int32(5)

	// msg type

	ONLINE_CODE    = int32(10)
	WIFI_GPRS_META = int32(11)
	LORA_GATE_META = int32(12)
	LORA_NODE_META = int32(13)
	TCP_WS_META    = int32(14)
	WIFI_GPRS_RX   = int32(21)
	LORA_GATE_RX   = int32(22)
	LORA_NODE_RX   = int32(23)
	TCP_WS_RX      = int32(24)
)
