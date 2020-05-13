package simulateutils

import (
	"highway/common"
	"sync"
)

type CommitteeTable struct {
	PubKeyBySID map[byte][]string
	SIDByPubKey map[string]byte
	lock        *sync.RWMutex
}

func NewCommitteeTable() *CommitteeTable {
	return &CommitteeTable{
		PubKeyBySID: map[byte][]string{},
		SIDByPubKey: map[string]byte{},
		lock:        &sync.RWMutex{},
	}
}

func (table *CommitteeTable) AddPubKey(pubKey string, SID byte) {
	table.lock.Lock()
	defer table.lock.Unlock()
	if sID, ok := table.SIDByPubKey[pubKey]; ok {
		if sID == SID {
			return
		}
		delete(table.SIDByPubKey, pubKey)
		table.PubKeyBySID[sID] = common.DeleteStringInList(pubKey, table.PubKeyBySID[sID])
	}
	table.SIDByPubKey[pubKey] = SID
	table.PubKeyBySID[SID] = append(table.PubKeyBySID[SID], pubKey)
}

func (table *CommitteeTable) GetKeysByKey(pubKey string) []string {
	table.lock.RLock()
	defer table.lock.RUnlock()
	if sID, ok := table.SIDByPubKey[pubKey]; ok {
		return table.PubKeyBySID[sID]
	}
	return []string{}
}
