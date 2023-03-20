package config

// TODO 增加配置项的通用校验规则

var ConfigStruct = []CfgSection{
	{"", []CfgKey{
		{&ServerNetMTU, "mtu", "1500", "", nil},
		{&RuntimeThreads, "max thread", "100", "", nil},
	}},
	{"audio", []CfgKey{
		{&SupportAudioBits, "support bits", "u8/u16le/u24le/s32le/u32le/fltle", "", parseBits},
		{&SupportAudioRates, "support rates", "44100/48000/96000/192000", "", parseRates},
		{&AudioBuferSize, "buffer size", "512", "", nil},
	}},
	{"detect", []CfgKey{
		{&ServerListen, "listen", "4414", "", nil},
		{&SpeakerOfflineTimeout, "offline timeout", "5", "", nil},
		{&SpeakerOfflineCheckInterval, "offline check interval", "5", "", nil},
	}},
	{"speaker", []CfgKey{
		{&SpeakerDir, "save dir", "speakers/", "", nil},
		{&ReadBufferSize, "receive buffer", "1024", "", nil},
		{&SendRoutinesMax, "send thread max", "2", "", nil},
		{&SendQueueSize, "send queue size", "16", "", nil},
		{&ReadQueueSize, "read queue size", "512", "", nil},
	}},
	{"http", []CfgKey{
		{&HTTPListen, "listen", "4415", "", nil},
		{&HTTPRoot, "root", "web/public", "", parsePath},
	}},
	{"receive", []CfgKey{
		{&ReceiveListen, "listen", "4416", "", nil},
		{&ReceiveTempDir, "tempdir", "", "", parseTempDir},
		{&EnableDLNA, "dlna", "false", "", nil},
	}},
	{"dlna", []CfgKey{
		{&DLNAListen, "listen", "4416", "", nil},
		{&DLNANotifyInterval, "notify interval", "30", "", nil},
		{&DLNAAllowIps, "allow ips", "", "", nil},
		{&DLNADenyIps, "deny ips", "", "", nil},
	}},
}
