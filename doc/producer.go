package main

import (
	"expvar"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

func add(i *int, d, min int) int {
	*i += d
	if *i < min {
		*i = min
	}
	return *i
}

var (
	memMap   = expvar.NewMap("mem")
	cpuMap   = expvar.NewMap("cpu")
	threads  = expvar.NewInt("threads")
	memHeap  expvar.Int
	memSys   expvar.Int
	memStack expvar.Int
	cpuUTime expvar.Int
	cpuSTime expvar.Int
)

func main() {
	sock, err := net.Listen("tcp", ":8123")
	if err != nil {
		log.Fatal("Error listening on port 8123")
		os.Exit(1)
	}
	go http.Serve(sock, nil)

	t := time.NewTicker(time.Millisecond * 250)
	var utime, stime, cpu int

	memMap.Set("heap", &memHeap)
	memMap.Set("sys", &memSys)
	memMap.Set("stack", &memStack)
	cpuMap.Set("uTime", &cpuUTime)
	cpuMap.Set("sTime", &cpuSTime)

	for range t.C {
		memHeap.Set(int64(10000 + rand.Intn(2000)))
		memSys.Set(int64(20000 + rand.Intn(1000)))
		memStack.Set(int64(3000 + rand.Intn(500)))
		cpuUTime.Set(int64(add(&utime, 100+rand.Intn(100), 0)))
		cpuSTime.Set(int64(add(&stime, 100+rand.Intn(200), 0)))
		threads.Set(int64(add(&cpu, rand.Intn(10)-4, 1)))
	}
}
