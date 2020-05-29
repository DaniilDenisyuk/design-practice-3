package main

import (
	"container/heap"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/DaniilDenisyuk/design-practice-3/httptools"
	"github.com/DaniilDenisyuk/design-practice-3/signal"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

// An Server is something we manage in a priority queue.
type Server struct {
	url       string // The value of the item; arbitrary.
	priority  int64  // The priority of the item in the queue.
	isHealthy bool
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Server

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	iServer := x.(*Server)
	iServer.index = n
	*pq = append(*pq, iServer)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	iServer := old[n-1]
	return iServer
}

func (pq *PriorityQueue) Remove() {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(iServer *Server, url string, priority int64) {
	iServer.url = url
	iServer.priority = priority
	heap.Fix(pq, iServer.index)
}

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.

var (
	port       = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https      = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

var (
	timeout     = time.Duration(*timeoutSec) * time.Second
	serversPool = PriorityQueue{
		{"server1:8080", 0, true, 0},
		{"server2:8080", 0, true, 1},
		{"server3:8080", 0, true, 2},
	}
)

func scheme() string {
	if *https {
		return "https"
	}
	return "http"
}

func health(dst string) bool {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	req, _ := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s://%s/health", scheme(), dst), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	if resp.StatusCode != http.StatusOK {
		return false
	}
	return true
}

func forward(url string, rw http.ResponseWriter, r *http.Request) error {
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = url
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = url
	resp, err := http.DefaultClient.Do(fwdRequest)
	if err == nil {
		for k, values := range resp.Header {
			for _, value := range values {
				rw.Header().Add(k, value)
			}
		}
		if *traceEnabled {
			rw.Header().Set("lb-from", url)
		}
		log.Println("fwd", resp.StatusCode, resp.Request.URL)
		rw.WriteHeader(resp.StatusCode)
		defer resp.Body.Close()
		_, err := io.Copy(rw, resp.Body)
		if err != nil {
			log.Printf("Failed to write response: %s", err)
		}
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", url, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
}

func min(pq *PriorityQueue) (error, *Server) {
	if pq.Len() == 0 {
		return errors.New("no servers in queue"), nil
	}
	var err error = nil
	server := heap.Pop(pq).(*Server)
	if !server.isHealthy {
		pq.Remove()
		err, server = min(pq)
	}
	return err, server
}

func main() {
	flag.Parse()
	for index, _ := range serversPool {
		server := serversPool[index]
		go func() {
			for range time.Tick(10 * time.Second) {
				isHealthy := health(server.url)
				if !isHealthy {
					server.isHealthy = isHealthy
				} else if server.isHealthy == false { //add server to queue if it becomes healthy
					heap.Push(&serversPool, server)
				}
				log.Printf("%s: healthy? %t, connections: %d", server.url, isHealthy, server.priority)
			}
		}()
	}
	heap.Init(&serversPool)

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		err, server := min(&serversPool)
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		atomic.AddInt64(&server.priority, 1)
		forward(server.url, rw, r)
		atomic.AddInt64(&server.priority, -1)
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}
