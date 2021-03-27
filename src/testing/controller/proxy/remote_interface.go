// Code generated by mockery v2.1.0. DO NOT EDIT.

package proxy

import (
	io "io"

	distribution "github.com/distribution/distribution"

	mock "github.com/stretchr/testify/mock"
)

// RemoteInterface is an autogenerated mock type for the RemoteInterface type
type RemoteInterface struct {
	mock.Mock
}

// BlobReader provides a mock function with given fields: repo, dig
func (_m *RemoteInterface) BlobReader(repo string, dig string) (int64, io.ReadCloser, error) {
	ret := _m.Called(repo, dig)

	var r0 int64
	if rf, ok := ret.Get(0).(func(string, string) int64); ok {
		r0 = rf(repo, dig)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 io.ReadCloser
	if rf, ok := ret.Get(1).(func(string, string) io.ReadCloser); ok {
		r1 = rf(repo, dig)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(io.ReadCloser)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, string) error); ok {
		r2 = rf(repo, dig)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Manifest provides a mock function with given fields: repo, ref
func (_m *RemoteInterface) Manifest(repo string, ref string) (distribution.Manifest, string, error) {
	ret := _m.Called(repo, ref)

	var r0 distribution.Manifest
	if rf, ok := ret.Get(0).(func(string, string) distribution.Manifest); ok {
		r0 = rf(repo, ref)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(distribution.Manifest)
		}
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(string, string) string); ok {
		r1 = rf(repo, ref)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, string) error); ok {
		r2 = rf(repo, ref)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ManifestExist provides a mock function with given fields: repo, ref
func (_m *RemoteInterface) ManifestExist(repo string, ref string) (bool, *distribution.Descriptor, error) {
	ret := _m.Called(repo, ref)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(repo, ref)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 *distribution.Descriptor
	if rf, ok := ret.Get(1).(func(string, string) *distribution.Descriptor); ok {
		r1 = rf(repo, ref)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*distribution.Descriptor)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, string) error); ok {
		r2 = rf(repo, ref)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
