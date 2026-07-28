package controller

import "github.com/zwcway/castserver-go/common/audio"

func DefaultChannel() audio.Channel {
	return audio.ChannelNone
}

func DefaultRate() audio.Rate {
	return audio.AudioRate_44100
}

func DefaultBits() audio.Bits {
	return audio.Bits_S16LE
}
