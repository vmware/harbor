// Code generated by mockery v2.1.0. DO NOT EDIT.

package distribution

import (
	distribution "github.com/distribution/distribution"
	mock "github.com/stretchr/testify/mock"
)

// Manifest is an autogenerated mock type for the Manifest type
type Manifest struct {
	mock.Mock
}

// Payload provides a mock function with given fields:
func (_m *Manifest) Payload() (string, []byte, error) {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 []byte
	if rf, ok := ret.Get(1).(func() []byte); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]byte)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func() error); ok {
		r2 = rf()
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// References provides a mock function with given fields:
func (_m *Manifest) References() []distribution.Descriptor {
	ret := _m.Called()

	var r0 []distribution.Descriptor
	if rf, ok := ret.Get(0).(func() []distribution.Descriptor); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]distribution.Descriptor)
		}
	}

	return r0
}
