package protocol

const VERSION uint8 = 1

type Type uint8

const (
	PT_UNKNOWN Type = iota
	PT_Ping
	PT_Pong
	PT_SpeakerInfo         // 设备广播
	PT_ServerInfo          // 服务器响应广播
	PT_SpeakerLeave        // 设备下线
	PT_ServerLeave         // 服务器下线
	PT_ServerMutexRequest  // 服务器单例请求
	PT_ServerMutexResponse // 服务器单例响应
	PT_Control             // 控制设备
	PT_SpeakerDataPush     // 向设备发送数据
	PT_SpeakerDataResult   // 设备响应结果
	PT_SpeakerStat         // 设备发送状态

	PT_ReceiveDataRequest  Type = 101 + iota // 接收数据
	PT_ReceiveDataResponse                   // 响应结果
)

type Packer interface {
	Pack() (p *Package, err error)
}

type Unpacker interface {
	UnPack(p *Package) (err error)
}
