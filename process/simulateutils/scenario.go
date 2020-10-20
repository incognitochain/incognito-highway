package simulateutils

import (
	"encoding/json"
	"fmt"
	"sync"
)

type TriggerByBlockInput struct {
	Heights     []uint64 `json:"Heights"`
	TotalBranch int      `json:"Branchs"`
	ChoseBranch int      `json:"Chose"`
}

type TriggerByTxIDInput struct {
	TxIDs       []string `json:"TxIDs"`
	TotalBranch int      `json:"Branchs"`
	ChoseBranch int      `json:"Chose"`
}

type Scenario struct {
	Scenes map[string][]Scene
	Lock   *sync.RWMutex
}

func NewScenario() *Scenario {
	return &Scenario{
		Scenes: map[string][]Scene{},
		Lock:   &sync.RWMutex{},
	}
}

type ScenarioBasic struct {
	Scenes map[int]*SimpleTrigger
	Lock   sync.RWMutex
}

func NewScenarioBasic(supportShard []byte) *ScenarioBasic {
	res := &ScenarioBasic{}
	res.Scenes = make(map[int]*SimpleTrigger)
	for _, cID := range supportShard {
		res.Scenes[int(cID)] = NewSimpleTrigger()
	}
	return res
}

type Scene struct {
	From     uint64              `json:"From"`
	To       uint64              `json:"To"`
	PubGroup map[string][]string `json:"PublishGroup"`
}

func (s *Scenario) GetPubGroup(chainKey string, blkHeight uint64) map[string][]string {
	s.Lock.RLock()
	scenes, ok := s.Scenes[chainKey]
	s.Lock.RUnlock()
	i := 0
	defer func(s *Scenario, chainKey string, i int) {
		go func(s *Scenario) {
			s.Lock.Lock()
			if _, ok := s.Scenes[chainKey]; ok {
				s.Scenes[chainKey] = s.Scenes[chainKey][i:]
			}
			s.Lock.Unlock()
		}(s)
	}(s, chainKey, i)
	if !ok {
		return nil
	}
	if len(scenes) == 0 {
		return nil
	}
	for i = 0; i < len(scenes); i++ {
		if scenes[i].From > blkHeight {
			return nil
		}
		if scenes[i].To >= blkHeight {
			return scenes[i].PubGroup
		}
	}
	return nil
}

func (s *Scenario) SetScenes(scenes map[string]Scene) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	for ck, scene := range scenes {
		if _, ok := s.Scenes[ck]; ok {
			s.Scenes[ck] = append(s.Scenes[ck], scene)
		} else {
			s.Scenes[ck] = []Scene{scene}
		}
	}
	return nil
}

func (s *ScenarioBasic) SetScenes(sType SceneType, scene []byte, cID int) error {
	s.Lock.RLock()
	simpleScene, ok := s.Scenes[cID]
	s.Lock.RUnlock()
	if !ok {
		return fmt.Errorf("Set fork scenes error: Do not support committee ID %v", cID)
	}
	switch sType {
	case TRIGGER_BY_BLOCKHEIGHT:
		newS := new(TriggerByBlockInput)
		err := json.Unmarshal(scene, newS)
		if err != nil {
			return err
		}
		simpleScene.Lock()
		bs := BasicForkScene{
			TotalForkBlock: uint64(newS.TotalBranch),
			NextBlock:      uint64(newS.ChoseBranch),
		}
		for _, h := range newS.Heights {
			simpleScene.ByBlock[h] = bs
		}
		simpleScene.Unlock()
	case TRIGGER_BY_TXID:
		newS := new(TriggerByTxIDInput)
		err := json.Unmarshal(scene, newS)
		if err != nil {
			return err
		}
		simpleScene.Lock()
		bs := BasicForkScene{
			TotalForkBlock: uint64(newS.TotalBranch),
			NextBlock:      uint64(newS.ChoseBranch),
		}
		for _, txID := range newS.TxIDs {
			simpleScene.ByTxID[txID] = bs
		}
		simpleScene.Unlock()
	}
	return nil
}
