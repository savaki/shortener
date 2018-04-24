package main

import (
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/tj/assert"
)

func TestConcurrency(t *testing.T) {
	http.Get("http://localhost:3000/def") // prime the lru cache

	begin := time.Now()
	concurrency := 500
	wg := &sync.WaitGroup{}
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req, err := http.NewRequest(http.MethodGet, "http://localhost:3000/def", nil)
			assert.Nil(t, err)

			_, err = http.DefaultTransport.RoundTrip(req)
			assert.Nil(t, err)
		}()
	}
	wg.Wait()

	fmt.Println(time.Now().Sub(begin).Round(time.Millisecond))
}
