package datahandler

import (
	"context"
	"highway/p2p"
	"highway/process/topic"
	"testing"

	libp2p "github.com/incognitochain/go-libp2p-pubsub"
	crypto "github.com/libp2p/go-libp2p-crypto"
)

func TestBlkCrossCommitteeHandler_HandleDataFromTopic(t *testing.T) {
	pri, _, _ := crypto.GenerateKeyPair(crypto.ECDSA, 2048)
	proxyHost := p2p.NewHost("0", "0.0.0.0", 8000, pri)
	FloodMachine, err := libp2p.NewFloodSub(context.Background(), proxyHost.Host)
	if err != nil {
		t.Error(err)
	}
	topic.Handler.Init("aaaa")
	topic.Handler.UpdateSupportShards([]byte{255, 0, 1, 2, 3, 4, 5, 6, 7})
	type fields struct {
		FromNode bool
		PubSub   *libp2p.PubSub
	}
	type args struct {
		topicReceived string
		dataReceived  libp2p.Message
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Inside case",
			fields: fields{
				FromNode: true,
				PubSub:   FloodMachine,
			},
			args: args{
				topicReceived: "blkshdtobcn--aaaa-nodepub",
				dataReceived:  libp2p.Message{},
			},
			wantErr: false,
		},
		{
			name: "Outside case",
			fields: fields{
				FromNode: false,
				PubSub:   FloodMachine,
			},
			args: args{
				topicReceived: "blkshdtobcn--",
				dataReceived:  libp2p.Message{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &BlkCrossCommitteeHandler{
				FromNode: tt.fields.FromNode,
				PubSub:   tt.fields.PubSub,
			}
			if err := handler.HandleDataFromTopic(tt.args.topicReceived, tt.args.dataReceived); (err != nil) != tt.wantErr {
				t.Errorf("BlkCrossCommitteeHandler.HandleDataFromTopic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
