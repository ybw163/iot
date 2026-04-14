package pb

// 消息类型常量
const (
	MsgTypeHeartbeat uint32 = 1 // 心跳
	MsgTypeTelemetry uint32 = 2 // 遥测
	MsgTypeAlert     uint32 = 3 // 告警
)

// 设备状态常量
const (
	StatusOffline int32 = 0 // 离线
	StatusOnline  int32 = 1 // 在线
)
