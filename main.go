package main

import (
	"fmt"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/humbertovnavarro/dankmesh/pkg/stream"
)

func main() {
	s := stream.MeshtasticTCPStream{
		Endpoint: "192.168.1.59",
	}
	s.Open()
	nodeInfo := meshtastic.ToRadio{PayloadVariant: &meshtastic.ToRadio_WantConfigId{WantConfigId: 42}}
	s.ToRadio() <- &nodeInfo
	for {
		if s.Closed {
			break
		}
		packet := <-s.FromRadio()
		fmt.Println(packet)
	}
}
