package stream

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"google.golang.org/protobuf/proto"
)

type MeshtasticTCPStream struct {
	Endpoint  string
	fromRadio chan *meshtastic.FromRadio
	toRadio   chan *meshtastic.ToRadio
	Closed    bool
}

func (s *MeshtasticTCPStream) FromRadio() chan *meshtastic.FromRadio {
	return s.fromRadio
}

func (s *MeshtasticTCPStream) ToRadio() chan *meshtastic.ToRadio {
	return s.toRadio
}

func (s *MeshtasticTCPStream) Open() {
	s.fromRadio = make(chan *meshtastic.FromRadio)
	s.toRadio = make(chan *meshtastic.ToRadio)
	fromRadioEndpoint := fmt.Sprintf("http://%s/api/v1/fromradio?all=false", s.Endpoint)
	toRadioEndpoint := fmt.Sprintf("http://%s/api/v1/toradio", s.Endpoint)
	go func() {
		for {
			if s.Closed {
				break
			}
			request, err := http.NewRequest("GET", fromRadioEndpoint, nil)
			request.Header.Add("Accept", "application/x-protobuf")
			if err != nil {
				continue
			}
			resp, err := http.DefaultClient.Do(request)
			if err != nil {
				continue
			}
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			fromRadioPacket := &meshtastic.FromRadio{}
			err = proto.Unmarshal(body, fromRadioPacket)
			if err != nil {
				continue
			}
			s.fromRadio <- fromRadioPacket
			time.Sleep(50 * time.Millisecond)
		}
	}()
	go func() {
		for {
			if s.Closed {
				break
			}
			packet := <-s.toRadio
			packetBytes, err := proto.Marshal(packet)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			req, err := http.NewRequest("PUT", toRadioEndpoint, bytes.NewReader(packetBytes))
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			_, err = http.DefaultClient.Do(req)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
		}
	}()
}

func (s *MeshtasticTCPStream) Close() error {
	s.Closed = true
	return nil
}
