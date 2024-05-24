package main_test

import (
	"testing"

	"buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMarshalProto(t *testing.T) {
	toRadio := &meshtastic.ToRadio{
		PayloadVariant: &meshtastic.ToRadio_WantConfigId{
			WantConfigId: 3412875888,
		},
	}
	packet, _ := proto.Marshal(toRadio)
	assert.Equal(t, []byte{0x18, 0xf0, 0xb4, 0xb1, 0xdb, 0xc}, packet)
	expected := &meshtastic.ToRadio{}
	proto.Unmarshal(packet, expected)
	expectedId := expected.GetPayloadVariant().(*meshtastic.ToRadio_WantConfigId).WantConfigId

	assert.Equal(t, expectedId, uint32(3412875888))
}
