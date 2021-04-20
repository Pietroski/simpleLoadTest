package main

import (
	"bufio"
	"fmt"
	"github.com/valyala/fasthttp"
	"net/http"
	"os"
	"runtime"
	"simpleLoadTest/controllers"
	"sync"
	"time"
)

var (
	MachineCores  = runtime.NumCPU()
	MaxGoroutines = runtime.GOMAXPROCS(MachineCores)
)

const (
	LOADER        = 48
	ExecutionTime = 10
	URL           = "https://jsonplaceholder.typicode.com/todos/1"
	// URL = "https://www.imovelweb.com.br/"
	//URL = "https://gatasproibidas.com/"
)

var (
	wg    sync.WaitGroup
	rc    sync.WaitGroup
	prntr sync.WaitGroup
	mtx   sync.Mutex

	rqs       int64
	AvgRespTm int64
	minTestTm int64 = 1000
	maxRespTm int64

	errorCounter int64

	clear map[string]func()

	tc = controllers.TerminalCleaner

	globalTime time.Time

	statusCode int
)

func init() {
	runtime.GOMAXPROCS(MachineCores)
	clear = *tc.SetCleaner()
}

func main() {

	//prntr.Add(1)
	//go func() {
	//	defer prntr.Done()
	//	for time.Since(globalTime).Seconds() <= ExecutionTime {
	//		countDown := time.Since(globalTime).Seconds()
	//		rps := int64(float64(rqs) / countDown)
	//		PrintToTerminal(statusCode, rqs, AvgRespTm, errorCounter, rps, countDown)
	//		tc.Clean(&clear)
	//	}
	//}()

	globalTime = time.Now()
	timer := time.Now()
	for time.Since(timer).Seconds() <= ExecutionTime {
		Shooter(LOADER)
	}
	Elapsed := time.Since(timer).Milliseconds()

	rps := float64(rqs) / float64(Elapsed/1000)
	rpt := rps / float64(LOADER)

	prntr.Wait()
	tc.Clean(&clear)

	fmt.Println()
	fmt.Printf("\nTest finished in -> %v milliseconds\n", Elapsed)
	fmt.Printf("\nTotal number of requests -> %v\n", rqs)
	fmt.Printf("\nAverage requests per unit of time -> %v\n", rps)
	fmt.Printf("\nAverage requests per Thread per unit of time -> %v\n", rpt)

	fmt.Printf("\nMinimum response time -> %v\n", minTestTm)
	fmt.Printf("\nAverage response time -> %v\n", AvgRespTm)
	fmt.Printf("\nMaximum response time -> %v\n", maxRespTm)
	fmt.Println()

	fmt.Printf("\nNumber of internal errors -> %v\n", errorCounter)
	fmt.Println(runtime.NumGoroutine(), runtime.NumCPU())
}

func Shooter(n int) {
	receiver := make(chan int64, n)
	errorReceiver := make(chan int64, n)

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(c, ec chan int64) {
			GetHTTP(c, ec)
			wg.Done()
		}(receiver, errorReceiver)
	}
	wg.Wait()

	rc.Add(2)
	go func() {
		defer rc.Done()
		for r := range receiver {
			rqs += r
		}
	}()
	close(receiver)
	go func() {
		defer rc.Done()
		for r := range errorReceiver {
			errorCounter += r
		}
	}()
	close(errorReceiver)
	rc.Wait()
}

func Get(c, ec chan int64) {
	// Acquire a request instance
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(URL)

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
	c <- 1

	mtx.Lock()
	if duration < minTestTm {
		minTestTm = duration
	}
	if duration > maxRespTm {
		maxRespTm = duration
	}
	AvgRespTm = (AvgRespTm + duration) / 2
	mtx.Unlock()

	//fmt.Printf("\nRequest took -> %v milliseconds\n", duration)
	//fmt.Println("Status code ->", resp.StatusCode)
	statusCode = resp.StatusCode()
}

func GetHTTP(c, ec chan int64) {
	start := time.Now()
	resp, err := http.Get(URL)
	if err != nil {
		fmt.Printf("Client get failed: %s\n", err)
		ec <- 1
		return
	}
	duration := time.Since(start).Milliseconds()
	c <- 1

	mtx.Lock()
	if duration < minTestTm {
		minTestTm = duration
	}
	if duration > maxRespTm {
		maxRespTm = duration
	}
	AvgRespTm = (AvgRespTm + duration) / 2

	//fmt.Printf("\nRequest took -> %v milliseconds\n", duration)
	//fmt.Println("Status code ->", resp.StatusCode)
	statusCode = resp.StatusCode

	mtx.Unlock()
}

func PrintToTerminal(sts int, rqs, AvgRespTm, err, rps int64, countDown float64) {
	w := bufio.NewWriter(os.Stdout)
	str := fmt.Sprintf(
		"Results -> sts: %v, rqs: %v avgrt: %v errs: %v rps: %v countdown: %v",
		sts, rqs, AvgRespTm, err, rps, countDown)
	fmt.Fprint(w, str)
	w.Flush()

	time.Sleep(1 * time.Second)
}
