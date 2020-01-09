package chain

import (
	"highway/chaindata"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChoosePeerFromGroup(t *testing.T) {
	buildGroup := func(height uint64) []chaindata.PeerWithBlk {
		g := []chaindata.PeerWithBlk{}
		for i := uint64(0); i < 10+height*100; i++ {
			g = append(g, chaindata.PeerWithBlk{
				Height: height,
			})
		}
		return g
	}

	testCases := []struct {
		desc          string
		groups        [][]chaindata.PeerWithBlk
		expectedGroup uint64
	}{
		{
			desc:          "Choose from 1st group",
			groups:        [][]chaindata.PeerWithBlk{buildGroup(0), buildGroup(1), buildGroup(2), buildGroup(3)},
			expectedGroup: 0,
		},
		{
			desc:          "Choose from 2nd group",
			groups:        [][]chaindata.PeerWithBlk{[]chaindata.PeerWithBlk{}, buildGroup(1), buildGroup(2), buildGroup(3)},
			expectedGroup: 1,
		},
		{
			desc:          "Choose from 3rd group",
			groups:        [][]chaindata.PeerWithBlk{[]chaindata.PeerWithBlk{}, []chaindata.PeerWithBlk{}, buildGroup(2), buildGroup(3)},
			expectedGroup: 2,
		},
		{
			desc:          "Choose from 4th group",
			groups:        [][]chaindata.PeerWithBlk{[]chaindata.PeerWithBlk{}, []chaindata.PeerWithBlk{}, []chaindata.PeerWithBlk{}, buildGroup(3)},
			expectedGroup: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			p, err := choosePeerFromGroup(tc.groups)
			if assert.Nil(t, err) {
				assert.Equal(t, tc.expectedGroup, p.Height)
			}
		})
	}
}

func TestCapBlocks(t *testing.T) {
	testCases := []struct {
		desc            string
		specific        bool
		from            uint64
		to              uint64
		heights         []uint64
		expectedTo      uint64
		expectedHeights []uint64
	}{
		{
			desc:       "Range, not exceeded cap",
			specific:   false,
			from:       5,
			to:         85,
			expectedTo: 85,
		},
		{
			desc:       "Range, exceeded cap",
			specific:   false,
			from:       15,
			to:         155,
			expectedTo: 115,
		},
		{
			desc:            "Specific, not exceeded cap",
			specific:        true,
			heights:         []uint64{5, 7, 15, 22, 33, 99, 150, 1555},
			expectedHeights: []uint64{5, 7, 15, 22, 33, 99, 150, 1555},
		},
		{
			desc:     "Specific, exceeded cap",
			specific: true,
			heights: func() []uint64 {
				h := []uint64{}
				for i := 0; i < 333; i++ {
					h = append(h, uint64(i))
				}
				return h
			}(),
			expectedHeights: func() []uint64 {
				h := []uint64{}
				for i := 0; i < 100; i++ {
					h = append(h, uint64(i))
				}
				return h
			}(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			to, heights := capBlocksPerRequest(tc.specific, tc.from, tc.to, tc.heights)
			if tc.specific {
				assert.Equal(t, tc.expectedHeights, heights)
			} else {
				assert.Equal(t, tc.expectedTo, to)
			}
		})
	}
}
