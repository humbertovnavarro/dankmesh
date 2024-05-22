package main

import (
	"bufio"
	"fmt"
	"io"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"go.bug.st/serial"
	"google.golang.org/protobuf/proto"
)

func main() {
	client := &MeshtasticClient{}
	client.ListenSerial("/dev/ttyUSB0", 32400)
}

type MeshtasticClient struct {
	readWriter *bufio.ReadWriter
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

const START0 = 0x94
const START1 = 0xc3

func (c *MeshtasticClient) Listen(readWriter io.ReadWriteCloser) error {
	bufferedReadWriter := bufio.NewReadWriter(bufio.NewReader(readWriter), bufio.NewWriter(readWriter))
	c.readWriter = bufferedReadWriter

	for {
		packet, err := c.adminPacket(12345, make([]byte, 0))
		c.SendPacket(packet)
		if err != nil {
			return err
		}

		header := make([]byte, 2)
		io.ReadFull(c.readWriter, header)
		fmt.Printf("%x,%x\n", header[0], header[1])
		if header[0] == START0 && header[1] == START1 {
			fmt.Println("======HEADER=======")
		}
	}
}

func (c *MeshtasticClient) ListenSerial(portPath string, baud int) error {
	if baud == 0 {
		baud = 32400
	}
	port, err := serial.Open(portPath, &serial.Mode{
		BaudRate: baud,
	})
	if err != nil {
		panic(err)
	}
	return c.Listen(port)
}

// Adds the appropriate header to the provided packet
func (c *MeshtasticClient) SendPacket(protobufPacket []byte) (err error) {
	packageLength := len(string(protobufPacket))
	header := []byte{START0, START1, byte(packageLength>>8) & 0xff, byte(packageLength) & 0xff}
	radioPacket := append(header, protobufPacket...)
	c.readWriter.Write(radioPacket)
	if err != nil {
		return err
	}
	return
}

// Constructs an admin packet on node with provided payload for administrative tasks (config, status, etc)
func (r *MeshtasticClient) adminPacket(nodeNum uint32, payload []byte) (packetOut []byte, err error) {
	radioMessage := meshtastic.ToRadio{
		PayloadVariant: &meshtastic.ToRadio_Packet{
			Packet: &meshtastic.MeshPacket{
				To:      nodeNum,
				WantAck: true,
				PayloadVariant: &meshtastic.MeshPacket_Decoded{
					Decoded: &meshtastic.Data{
						Payload:      payload,
						Portnum:      meshtastic.PortNum_ADMIN_APP,
						WantResponse: true,
					},
				},
			},
		},
	}
	packetOut, err = proto.Marshal(&radioMessage)
	if err != nil {
		return nil, err
	}
	return
}
