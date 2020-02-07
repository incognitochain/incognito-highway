package chain_test

import (
	context "context"
	"highway/chain"
	"highway/common"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBlockByHeightFiltered(t *testing.T) {
	p1 := &Provider{}
	d1 := [][]byte{nil, []byte{2}, nil, []byte{4}, []byte{5}, nil}
	p1.On("GetBlockByHeight", mock.Anything, mock.Anything, mock.Anything).Return(d1, nil)
	p1.On("SetBlockByHeight", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	p2 := &Provider{}
	d2 := [][]byte{[]byte{1}, []byte{3}, nil}
	p2.On("GetBlockByHeight", mock.Anything, mock.Anything, mock.Anything).Return(d2, nil)
	p2.On("SetBlockByHeight", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	p3 := &Provider{}
	d3 := [][]byte{nil}
	p3.On("GetBlockByHeight", mock.Anything, mock.Anything, mock.Anything).Return(d3, nil)
	providers := []chain.Provider{p1, p2, p3}
	s := chain.Server{Providers: providers}
	req := chain.RequestByHeight{}
	heights := []uint64{1, 2, 3, 4, 5, 6}
	data := s.GetBlockByHeight(context.Background(), req, heights)
	expectedData := [][]byte{[]byte{1}, []byte{2}, []byte{3}, []byte{4}, []byte{5}, nil}
	assert.Equal(t, expectedData, data)
}

func TestConvertToSpecific(t *testing.T) {
	maxBlocksPerRequest := common.MaxBlocksPerRequest
	common.MaxBlocksPerRequest = 5
	defer func() {
		common.MaxBlocksPerRequest = maxBlocksPerRequest
	}()

	testCases := []struct {
		desc            string
		specific        bool
		from            uint64
		to              uint64
		heights         []uint64
		expectedHeights []uint64
	}{
		{
			desc:            "Specific capped",
			specific:        true,
			heights:         []uint64{1, 2, 3, 4, 5, 6, 7, 8},
			expectedHeights: []uint64{1, 2, 3, 4, 5},
		},
		{
			desc:            "Specific, no change",
			specific:        true,
			heights:         []uint64{5, 6, 7, 8},
			expectedHeights: []uint64{5, 6, 7, 8},
		},
		{
			desc:            "Range converted into specific",
			specific:        false,
			from:            123,
			to:              456,
			expectedHeights: []uint64{123, 124, 125, 126, 127},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			heights := chain.ConvertToSpecificHeights(tc.specific, tc.from, tc.to, tc.heights)
			assert.Equal(t, tc.expectedHeights, heights)
		})
	}
}
