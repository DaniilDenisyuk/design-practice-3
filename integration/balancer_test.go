package integration

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

const baseAddress = "http://balancer:8090"

var (
	client = http.Client{
		Timeout: 3 * time.Second,
	}
	srvConns = map[string]int{
		"server1:8080": 0,
		"server2:8080": 0,
		"server3:8080": 0,
	}
)

func makeRequest(client http.Client, request *http.Request, connCounter map[string]int) {
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	connCounter[resp.Header.Get("lb-from")] += 1
}

func TestBalancer(t *testing.T) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/some-data", baseAddress), nil)
	//I didn't understand purpose of benchmark test here when we have constant delay so i did't wrote this
	//Explanation: 10 req/sec when request processing with delay for 1 sec means average load 3 request on server/sec
	//or 6 req on server for this test
	cnt := 0
	for range time.Tick(100 * time.Millisecond) {
		go makeRequest(client, req, srvConns)
		cnt++
		if cnt > 20 {
			break
		}
	}
	time.Sleep(5 * time.Second)
	assert.True(t, srvConns["server1:8080"] >= 6 && srvConns["server2:8080"] >= 6 && srvConns["server3:8080"] >= 6)
}
