package simulateutils

import (
	"sync"
)

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
