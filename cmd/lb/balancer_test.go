package main

import (
	"container/heap"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"testing"
)

func TestBalancerVeryComplex1(t *testing.T) {

	testServersPool := PriorityQueue{
		{"server1:8080", 0, true, 1},
		{"server2:8080", 0, true, 2},
		{"server3:8080", 0, true, 3},
	}
	heap.Init(&testServersPool)
	_, minServer := min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 17)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 15)
	atomic.AddInt64(&minServer.priority, -5)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 12)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 12)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 1)
	_, minServer = min(&testServersPool)

	// This is example of trouble
	println(testServersPool[0].priority, testServersPool[1].priority, testServersPool[2].priority)
	assert.Equal(t, minServer.priority, int64(17), "wrong server selected")
}
func TestBalancerVeryComplex2(t *testing.T) {

	testServersPool := PriorityQueue{
		{"server1:8080", 0, true, 1},
		{"server2:8080", 0, true, 2},
		{"server3:8080", 0, true, 3},
	}
	heap.Init(&testServersPool)
	_, minServer := min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 17)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 15)
	atomic.AddInt64(&minServer.priority, -5)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 12)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, 12)
	_, minServer = min(&testServersPool)
	atomic.AddInt64(&minServer.priority, -10)
	_, minServer = min(&testServersPool)
	_, minServer = min(&testServersPool)

	// This is example of trouble
	println(testServersPool[0].priority, testServersPool[1].priority, testServersPool[2].priority)
	assert.Equal(t, minServer.priority, int64(2), "wrong server selected")
}

func TestBalancerError(t *testing.T) {
	testServersPool := PriorityQueue{
		{"server1:8080", 0, false, 1},
		{"server2:8080", 0, false, 2},
		{"server3:8080", 0, false, 3},
	}
	heap.Init(&testServersPool)
	err, _ := min(&testServersPool)
	assert.Error(t, err, testServersPool.Len())
}

func TestBalancerSimple(t *testing.T) {
	testServersPool := PriorityQueue{
		{"server1:8080", 1, true, 1},
		{"server2:8080", 2, true, 2},
		{"server3:8080", 2, true, 3},
	}
	heap.Init(&testServersPool)
	_, minServer := min(&testServersPool)
	assert.Equal(t, minServer.priority, int64(1), "wrong server selected")
}

func TestBalancerComplex(t *testing.T) {
	testServersPool := PriorityQueue{
		{"server1:8080", 0, false, 1},
		{"server2:8080", 0, false, 2},
		{"server3:8080", 150, true, 3},
	}
	heap.Init(&testServersPool)
	_, minServer := min(&testServersPool)
	assert.Equal(t, minServer.url, "server3:8080", "wrong server selected")
	assert.Equal(t, testServersPool.Len(), 1, "wrong server selected")
}
