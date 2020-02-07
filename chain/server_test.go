package chain

import (
	"highway/common"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			heights := convertToSpecificHeights(tc.specific, tc.from, tc.to, tc.heights)
			assert.Equal(t, tc.expectedHeights, heights)
		})
	}
}
