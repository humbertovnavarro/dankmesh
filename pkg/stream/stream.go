package stream

import "buf.build/gen/go/meshtastic/protobufs/protocolbuffers/go/meshtastic"

type MeshtasticStream interface {
	FromRadioChannel() *meshtastic.FromRadio
	ToRadioChannel() *meshtastic.ToRadio
	Open() error
	Close() error
}
