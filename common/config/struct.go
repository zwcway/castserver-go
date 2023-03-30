package config

// TODO 增加配置项的通用校验规则

var ConfigStruct = []CfgSection{
	{"", []CfgKey{
		{&ServerNetMTU, "mtu", "", nil},
		{&RuntimeThreads, "max thread", "", nil},
	}},
	{"audio", []CfgKey{
		{&SupportAudioBits, "support bits", "", parseBits},
		{&SupportAudioRates, "support rates", "", parseRates},
		{&AudioBuferMSDuration, "buffer duration", "", nil},
	}},
	{"detect", []CfgKey{
		{&ServerListen, "listen", "", nil},
		{&SpeakerOfflineTimeout, "offline timeout", "", nil},
		{&SpeakerOfflineCheckInterval, "offline check interval", "", nil},
	}},
	{"speaker", []CfgKey{
		{&DB, "database", "", parseDatabase},
		{&ReadBufferSize, "receive buffer", "", nil},
		{&SendRoutinesMax, "send thread max", "", nil},
		{&SendQueueSize, "send queue size", "", nil},
		{&ReadQueueSize, "read queue size", "", nil},
	}},
	{"http", []CfgKey{
		{&HTTPListen, "listen", "", nil},
		{&HTTPRoot, "root", "", parsePath},
		{&WSClientMAX, "client max", "", nil},
	}},
	{"receive", []CfgKey{
		{&ReceiveListen, "listen", "", nil},
		{&ReceiveTempDir, "tempdir", "", parseTempDir},
	}},
	{"dlna", []CfgKey{
		{&EnableDLNA, "enable", "", nil},
		{&DLNAListen, "listen", "", nil},
		{&DLNANotifyInterval, "notify interval", "", nil},
		{&DLNAAllowIps, "allow ips", "", nil},
		{&DLNADenyIps, "deny ips", "", nil},
	}},
}
