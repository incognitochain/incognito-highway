// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import peer "github.com/libp2p/go-libp2p-core/peer"

// Publisher is an autogenerated mock type for the Publisher type
type Publisher struct {
	mock.Mock
}

// BlacklistPeer provides a mock function with given fields: _a0
func (_m *Publisher) BlacklistPeer(_a0 peer.ID) {
	_m.Called(_a0)
}

// Publish provides a mock function with given fields: topic, msg
func (_m *Publisher) Publish(topic string, msg []byte) error {
	ret := _m.Called(topic, msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte) error); ok {
		r0 = rf(topic, msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
