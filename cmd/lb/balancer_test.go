package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBalancerEmpty(t *testing.T) {
	testServersPool := []server{
		{"server1:8080", 0, true},
		{"server2:8080", 0, true},
		{"server3:8080", 0, true},
	}

	minIndex, err := min(testServersPool, func(a, b server) bool { return a.connCnt < b.connCnt })
	assert.Nil(t, err)
	assert.Equal(t, minIndex, 0, "wrong server selected")
}

func TestBalancerSimple1(t *testing.T) {
	testServersPool := []server{
		{"server1:8080", 2, true},
		{"server2:8080", 2, true},
		{"server3:8080", 1, true},
	}

	minIndex, err := min(testServersPool, func(a, b server) bool { return a.connCnt < b.connCnt })
	assert.Nil(t, err)
	assert.Equal(t, minIndex, 2, "wrong server selected")
}

func TestBalancerSimple2(t *testing.T) {
	testServersPool := []server{
		{"server1:8080", 6, true},
		{"server2:8080", 5, true},
		{"server3:8080", 5, true},
	}

	minIndex, err := min(testServersPool, func(a, b server) bool { return a.connCnt < b.connCnt })
	assert.Nil(t, err)
	assert.Equal(t, minIndex, 1, "wrong server selected")
}

func TestBalancerComplex1(t *testing.T) {
	testServersPool := []server{
		{"server1:8080", 6, true},
		{"server2:8080", 0, false},
		{"server3:8080", 5, true},
	}

	minIndex, err := min(testServersPool, func(a, b server) bool { return a.connCnt < b.connCnt })
	assert.Nil(t, err)
	assert.Equal(t, minIndex, 2, "wrong server selected")
}

func TestBalancerComplex2(t *testing.T) {
	testServersPool := []server{
		{"server1:8080", 0, false},
		{"server2:8080", 0, false},
		{"server3:8080", 85, true},
	}
	minIndex, err := min(testServersPool, func(a, b server) bool { return a.connCnt < b.connCnt })
	assert.Equal(t, err, nil, err)
	assert.Equal(t, minIndex, 2, "wrong server selected")
}

func TestBalancerError(t *testing.T) {
	testServersPool := []server{
		{"server1:8080", 0, false},
		{"server2:8080", 0, false},
		{"server3:8080", 0, false},
	}
	_, err := min(testServersPool, func(a, b server) bool { return a.connCnt < b.connCnt })
	assert.NotNil(t, err, "error test not passed")
}
