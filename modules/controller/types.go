package controller

type Command uint32

const (
	Command_UNKNOWN Command = iota
	Command_SAMPLE 
	Command_CHUNK
	Command_TIME
	Command_VOLUME

	Command_MAX
)


type requestHeader  struct{
	ver uint8
	
}