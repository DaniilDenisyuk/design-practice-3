package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/DaniilDenisyuk/design-practice-3/httptools"
	"github.com/DaniilDenisyuk/design-practice-3/signal"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	port       = flag.Int("port", 8090, "load balancer port")
	timeoutSec = flag.Int("timeout-sec", 3, "request timeout time in seconds")
	https      = flag.Bool("https", false, "whether backends support HTTPs")

	traceEnabled = flag.Bool("trace", false, "whether to include tracing information into responses")
)

type server struct {
	url       string
	connCnt   int
	isHealthy bool
}

var (
	timeout     = time.Duration(*timeoutSec) * time.Second
	serversPool = []server{
		{"server1:8080", 0, true},
		{"server2:8080", 0, true},
		{"server3:8080", 0, true},
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

func forward(dst *server, rw http.ResponseWriter, r *http.Request) error {
	ctx, _ := context.WithTimeout(r.Context(), timeout)
	url := dst.url
	fwdRequest := r.Clone(ctx)
	fwdRequest.RequestURI = ""
	fwdRequest.URL.Host = url
	fwdRequest.URL.Scheme = scheme()
	fwdRequest.Host = url
	dst.connCnt += 1
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
		dst.connCnt -= 1
		return nil
	} else {
		log.Printf("Failed to get response from %s: %s", dst.url, err)
		rw.WriteHeader(http.StatusServiceUnavailable)
		dst.connCnt -= 1
		return err
	}
}

func min(serversPool []server, condition func(a, b server) bool) (int, error) {
	serverIndex := 0
	for index, server := range serversPool {
		if server.isHealthy {
			if condition(server, serversPool[serverIndex]) {
				serverIndex = index
			} else if !serversPool[serverIndex].isHealthy {
				serverIndex = index
			}
		}
	}
	if !serversPool[serverIndex].isHealthy {
		return 0, errors.New("No healthy servers left")
	}
	return serverIndex, nil
}

func main() {
	flag.Parse()
	// TODO: Використовуйте дані про стан сервреа, щоб підтримувати список тих серверів, яким можна відправляти ззапит.
	for index, _ := range serversPool {
		server := &serversPool[index]
		go func() {
			for range time.Tick(10 * time.Second) {
				isHealthy := health(server.url)
				if isHealthy != true {
					server.isHealthy = isHealthy
				}
				log.Printf("%s: healthy? %t, connections: %d", server.url, isHealthy, server.connCnt)
			}
		}()
	}

	frontend := httptools.CreateServer(*port, http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		serverIndex, err := min(serversPool, func(a, b server) bool { return a.connCnt < b.connCnt })
		if err != nil {
			log.Println(err)
			rw.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		server := &serversPool[serverIndex]
		forward(server, rw, r)
	}))

	log.Println("Starting load balancer...")
	log.Printf("Tracing support enabled: %t", *traceEnabled)
	frontend.Start()
	signal.WaitForTerminationSignal()
}
