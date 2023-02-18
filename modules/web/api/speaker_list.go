package api

import "github.com/fasthttp/websocket"

type ResponseSpeakerList struct {
	Id   string `jp:"id"`
	Name string `jp:"name"`
}

func SpeakerList(c *websocket.Conn, req any) {

}
