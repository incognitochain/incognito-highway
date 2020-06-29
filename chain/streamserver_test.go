package chain

import (
	"highway/common"
	"highway/proto"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func sendBlk(blkChan chan common.ExpectedBlk, numOfBlock int) {
	for i := 0; i < numOfBlock; i++ {
		blkChan <- common.ExpectedBlk{
			Height: uint64(i),
			Data:   []byte{byte(i)},
		}
	}
	close(blkChan)
}

func TestSendWithTimeout(t *testing.T) {
	type args struct {
		blkChan    chan common.ExpectedBlk
		timeout    time.Duration
		send       func(*proto.BlockData) error
		numOfBlock int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		wantInt uint
	}{
		{
			name: "Happy case",
			args: args{
				blkChan: make(chan common.ExpectedBlk),
				timeout: 5 * time.Second,
				send: func(bd *proto.BlockData) error {
					time.Sleep(1 * time.Second)
					return nil
				},
				numOfBlock: 3,
			},
			wantErr: false,
			wantInt: 3,
		},
		{
			name: "error case",
			args: args{
				blkChan: make(chan common.ExpectedBlk),
				timeout: 5 * time.Second,
				send: func(bd *proto.BlockData) error {
					return errors.New("Client close stream")
				},
				numOfBlock: 5,
			},
			wantErr: true,
			wantInt: 0,
		},
		{
			name: "Timeout case",
			args: args{
				blkChan: make(chan common.ExpectedBlk),
				timeout: 5 * time.Second,
				send: func(bd *proto.BlockData) error {
					time.Sleep(2 * time.Second)
					return nil
				},
				numOfBlock: 5,
			},
			wantErr: true,
			wantInt: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go sendBlk(tt.args.blkChan, tt.args.numOfBlock)
			if got, err := SendWithTimeout(tt.args.blkChan, tt.args.timeout, tt.args.send); ((err != nil) != tt.wantErr) || (got != tt.wantInt) {
				t.Errorf("SendWithTimeout() error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("SendWithTimeout() numblks = %v, want %v", got, tt.wantInt)
				for range tt.args.blkChan {
				}
			}
		})
	}
}
