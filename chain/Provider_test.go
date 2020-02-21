// Code generated by mockery v1.0.0. DO NOT EDIT.

package chain_test

import chain "highway/chain"
import context "context"
import mock "github.com/stretchr/testify/mock"

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

// GetBlockByHash provides a mock function with given fields: ctx, req, hashes
func (_m *Provider) GetBlockByHash(ctx context.Context, req chain.GetBlockByHashRequest, hashes [][]byte) ([][]byte, error) {
	ret := _m.Called(ctx, req, hashes)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func(context.Context, chain.GetBlockByHashRequest, [][]byte) [][]byte); ok {
		r0 = rf(ctx, req, hashes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, chain.GetBlockByHashRequest, [][]byte) error); ok {
		r1 = rf(ctx, req, hashes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBlockByHeight provides a mock function with given fields: ctx, req, heights
func (_m *Provider) GetBlockByHeight(ctx context.Context, req chain.GetBlockByHeightRequest, heights []uint64) ([][]byte, error) {
	ret := _m.Called(ctx, req, heights)

	var r0 [][]byte
	if rf, ok := ret.Get(0).(func(context.Context, chain.GetBlockByHeightRequest, []uint64) [][]byte); ok {
		r0 = rf(ctx, req, heights)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, chain.GetBlockByHeightRequest, []uint64) error); ok {
		r1 = rf(ctx, req, heights)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetBlockByHeight provides a mock function with given fields: ctx, req, heights, blocks
func (_m *Provider) SetBlockByHeight(ctx context.Context, req chain.GetBlockByHeightRequest, heights []uint64, blocks [][]byte) error {
	ret := _m.Called(ctx, req, heights, blocks)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, chain.GetBlockByHeightRequest, []uint64, [][]byte) error); ok {
		r0 = rf(ctx, req, heights, blocks)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}