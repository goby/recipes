package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/juju/ratelimit"
)

type Server struct {
	buckets []int
	count   *int64
	limiter *ratelimit.Bucket
}

// recept req from client, suppose all request need retry.
// return retry time
func (s *Server) req() int {
	// do something suppose
	time.Sleep(time.Duration(rand.Intn(10) * 1e6))

	retry := 1
	atomic.AddInt64(s.count, 1)
	s.limiter.TakeAvailable(1)
	available := int(s.limiter.Available())
	for _, bucket := range s.buckets {
		if available >= bucket {
			break
		}
		retry++
	}

	return retry
}

func (s *Server) run() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	prev := atomic.LoadInt64(s.count)
	for {
		select {
		case <-ticker.C:
			curr := atomic.LoadInt64(s.count)
			fmt.Printf("%-39v: %d\n", time.Now(), curr-prev)
			prev = curr
		}
	}
}

type Client struct {
	retry  time.Duration
	server *Server
}

func (c *Client) run() {
	for {
		r := c.server.req()
		time.Sleep(time.Duration(r * 1e9))
	}
}

func main() {
	fmt.Println("starting...")
	clientCount := 30000
	// config
	capacity := 2048
	rate := capacity
	buckets := []int{1024, 512, 256, 128, 128, 64, 32, 10}

	var count int64
	s := Server{
		buckets: buckets,
		count:   &count,
		limiter: ratelimit.NewBucketWithRate(float64(rate), int64(capacity)),
	}

	// print qps
	go s.run()

	for i := 0; i < clientCount; i++ {
		c := Client{server: &s}
		go c.run()
	}

	select {}
}
