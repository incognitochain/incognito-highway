package common

type CommitteePublicKey struct {
	IncPubKey    []byte
	MiningPubKey map[string][]byte
}

type MiningPublicKey map[string][]byte

var (
	SelfID string = "aaaaa"
	//Remove soon
	CommitteeGenesis        map[string]byte
	MiningKeyByCommitteeKey map[string]string
)

const (
	BEACONID      byte = 255
	NumberOfShard      = 2
	CommitteeSize      = 4
)
