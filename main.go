package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var wg sync.WaitGroup

var mtx sync.Mutex
var rqs int64

func main() {

	timer := time.Now()
	for rqs < 30_000 {
		//Shooter(500)
		select {
		case <-time.After(1 * time.Millisecond):
			Shooter(500)
		}
	}
	Elapsed := time.Since(timer).Milliseconds()

	fmt.Printf("\nTest finished in -> %v milliseconds\n", Elapsed)

	rps := float64(rqs)/float64(Elapsed/1000)
	fmt.Printf("\nAverage requests per unit of time -> %v\n", rps)

	fmt.Println(runtime.NumGoroutine(), runtime.NumCPU())
}

func Shooter(n int) {
	atomic.AddInt64(&rqs, int64(n))
	fmt.Println(atomic.LoadInt64(&rqs))

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			Get()
			wg.Done()
		}()
	}
	wg.Wait()
	//fmt.Println(runtime.NumGoroutine())
}

func Get() {
	url := "https://jsonplaceholder.typicode.com/todos/2"

	// Acquire a request instance
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)

	// Acquire a response instance
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	start := time.Now()
	err := fasthttp.Do(req, resp)
	if err != nil {
		fmt.Printf("Client get failed: %s\n", err)
		return
	}
	duration := time.Since(start).Milliseconds()
	fmt.Printf("\nRequest took -> %v milliseconds\n", duration)

	fmt.Println("Status code ->", resp.StatusCode())
}
