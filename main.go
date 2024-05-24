package main

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"go.bug.st/serial"
	"google.golang.org/protobuf/proto"
)

func main() {
	client := &MeshtasticClient{}
	device := FindMeshtasticSerial()
	port, err := serial.Open(device, &serial.Mode{
		BaudRate: DEFAULT_BAUD_RATE,
	})
	if err != nil {
		panic(err)
	}
	err = client.Listen(port)
	if err != nil {
		panic(err)
	}
	fmt.Println("Done")
}

type MeshtasticClient struct {
	rw *bufio.ReadWriter
}

// takes a readWriterCloser that talks meshtastic protobuf
/**
The 4 byte header is constructed to both provide framing and to not look like 'normal' 7 bit ASCII.

    Byte 0: START1 (0x94)
    Byte 1: START2 (0xc3)
    Byte 2: MSB of protobuf length
    Byte 3: LSB of protobuf length

The receiver will validate length and if >512 it will assume the packet is corrupted and return to looking for START1. While looking for START1
see https://meshtastic.org/docs/development/reference/protobufs/
**/
const DEFAULT_BAUD_RATE = 115200
const START0 byte = 0x94
const START1 byte = 0xC3
const MAX_FROM_RADIO_SIZE = 512

func (c *MeshtasticClient) handlePacket(fromRadioPacket meshtastic.FromRadio) error {
	packet := fromRadioPacket.GetPacket()
	if packet == nil {
		return fmt.Errorf("packet is nil")
	}
	decodedPacket := packet.GetDecoded()
	if decodedPacket == nil {
		return fmt.Errorf("decoded packet is nil")
	}
	fmt.Printf("Received packet: %v\n", decodedPacket)
	return nil
}

func (c *MeshtasticClient) Listen(readWriter io.ReadWriteCloser) error {
	c.rw = bufio.NewReadWriter(bufio.NewReader(readWriter), bufio.NewWriter(readWriter))
	p := make([]byte, 32)
	for i := range p {
		p[i] = byte(START1)
	}
	c.rw.Write(p)
	c.rw.Flush()
	time.Sleep(100 * time.Millisecond)

	toRadio := meshtastic.ToRadio{
		PayloadVariant: &meshtastic.ToRadio_WantConfigId{
			WantConfigId: 623995978,
		},
	}
	packet, _ := proto.Marshal(&toRadio)
	c.rw.Write(AddMeshtasticHeader(packet))
	c.rw.Flush()

	for {
		header := make([]byte, 4)
		n, err := io.ReadFull(c.rw, header)
		if n != 4 || err != nil || header[0] != START0 || header[1] != START1 {
			continue
		}
		packetLength := (int(header[2]) << 8) + (int)(header[3])
		if packetLength > MAX_FROM_RADIO_SIZE || packetLength < 0 {
			continue
		}
		protoPacket := make([]byte, packetLength)
		n, err = io.ReadFull(c.rw, protoPacket)
		if n != packetLength || err != nil {
			continue
		}
		message := meshtastic.FromRadio{}
		err = proto.Unmarshal(protoPacket, &message)
		if err != nil {
			continue
		}
		c.handlePacket(message)
	}
}

// byte[] returned by proto.Marshal()
type MeshtasticProtoBuffPacket = []byte

// raw byte[] send over the wire
type MeshtasticTransportPacket = []byte

func AddMeshtasticHeader(packet MeshtasticProtoBuffPacket) MeshtasticTransportPacket {
	packetLength := len(packet)
	header := []byte{START0, START1, byte(packetLength>>8) & 0xff, byte(packetLength) & 0xff}
	return append(header, packet...)
}

func FindMeshtasticSerial() (port string) {
	ports, err := serial.GetPortsList()
	if err != nil {
		panic(err)
	}
	for _, _port := range ports {
		port = _port
	}
	return
}
