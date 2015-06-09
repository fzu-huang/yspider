// tryquery_norecursion project main.go
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

//var m sync.WaitGroup

type Record struct {
	Visited map[string]bool
	Loc     *sync.RWMutex
}

type CheckNode struct {
	url string
	dep int
}

var Crawtcount int
var Rec Record
var curl chan CheckNode
var exit chan bool
var Count, ValidCount, ReuseCount, ImgCount int64

const (
	Depth = 5
)

func main() {
	Crawtcount = 0
	leng := 30
	curl = make(chan CheckNode, leng)
	exit = make(chan bool, leng)
	//m = sync.WaitGroup{}
	Count = 0
	ValidCount = 0
	ReuseCount = 0
	ImgCount = 0
	Rec = Record{
		Visited: make(map[string]bool),
		Loc:     new(sync.RWMutex),
	}
	fmt.Println("Hello World!")

	runtime.GOMAXPROCS(runtime.NumCPU())

	//url := "http://alpha.wallhaven.cc/random"
	//url := "http://www.xinhuanet.com"
	url := "http://tieba.baidu.com"
	curl <- CheckNode{url, 1}
	for i := 0; i < leng; i++ {
		//m.Add(1)
		go do()
	}

	//go func() {
	ticker := time.NewTicker(time.Second * 5)
	idleticker := time.NewTicker(time.Second * 20)
	defer idleticker.Stop()
	defer ticker.Stop()
	var tmpvisitcount int64
	tmpvisitcount = 0
	for {
		select {
		case <-ticker.C:
			fmt.Println("RUNNING ROUTINES COUNT: ", runtime.NumGoroutine(), ", ReVis Count:", ReuseCount, ", ImgCount:", ImgCount, ", All Visited Count: ", Count, ", Valid url Count:", ValidCount)
			runtime.GC()
			//fmt.Println(runtime.MemProfileRate)
		case <-exit:

			//fmt.Println("a routine is stucked. run a new one.....")
			go do()

		case <-idleticker.C:
			if tmpvisitcount < Count {
				tmpvisitcount = Count
			} else {
				fmt.Println("long time no visit url,  maybe it is stuck... mission end..")
				return
			}
		}
	}
	//}()

	//m.Wait()

}
