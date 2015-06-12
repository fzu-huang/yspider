package main

import (
	//"fmt"
	"github.com/PuerkitoBio/goquery"
	//"github.com/robertkrimen/otto"
	//"github.com/robertkrimen/otto/parser"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func do() {
	idletime := time.NewTicker(time.Second * 10)
	ticktime := time.NewTicker(time.Second * 1)
	idle := false
	idlecount := 0
	defer idletime.Stop()
	go func() {
		for {
			select {
			case <-ticktime.C:
				idlecount++
				if idlecount >= 10 {
					idle = true
				}
			case <-idletime.C:
				if idle {
					//fmt.Println("long time no in channel,  abandon the routine...")
					defer func() {
						exit <- true
					}()
					return
				}
			}
		}
	}()

	for {
		select {
		case node := <-curl:
			idlecount = 0
			idle = false
			Craw(node)
		}
	}
}

func Craw(node CheckNode) {
	//	Crawtcount++
	//	fmt.Println(Crawtcount)
	//	defer func() {
	//		fmt.Println("done!", Crawtcount)
	//	}()
	url := node.url
	dep := node.dep

	if strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".JPG") || strings.HasSuffix(url, ".png") {
		atomic.AddInt64(&ImgCount, 1)
		go WriteFile(url)
		return
	}

	//	if strings.HasSuffix(url, ".torrent") || strings.HasSuffix(url, ".pdf") {
	//		atomic.AddInt64(&ImgCount, 1)
	//		fmt.Println("go  write...")
	//		go WriteFile(url)
	//		return
	//	}

	if dep+1 > Depth {
		return
	}
	atomic.AddInt64(&Count, 1)

	doc, err := goquery.NewDocument(url)

	if err != nil {
		//fmt.Println(err)
		netErrCount++
		return
	}

	//find javascript
	//	jssele := doc.Find("Script")
	//	l := len(jssele.Nodes)
	//	if jssele != nil && l != 0 {
	//		for i := 0; i < l; i++ {
	//			fmt.Println(jssele.Eq(i).Text())
	//			go ParseJS(jssele.Text())
	//		}
	//		return
	//	}

	selection := doc.Find("[href],[src],[url]")
	if selection == nil || len(selection.Nodes) == 0 {
		//fmt.Println("Find no links, close session")
		return
	}
	atomic.AddInt64(&ValidCount, 1)
	//	count := ValidCount
	//fmt.Println("find start", count)
	for _, node := range selection.Nodes {
		for _, att := range node.Attr {
			if att.Key == "href" || att.Key == "src" || att.Key == "url" {
				if nexturl, ok := Check(att.Val); ok {
					//fmt.Println("DEPTH: ", dep+1, ", ", nexturl)
					CheckNotExist(nexturl, dep)
				}
			}
		}
	}
	//fmt.Println("find end", count)
	return
}

//文件的获取和写入
func WriteFile(url string) {
	atomic.AddInt64(&writingCount, 1)
	defer atomic.AddInt64(&writingCount, -1)
	res, err := http.Get(url)
	if err != nil {
		//fmt.Println("write file err!!", err)
		return
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		//fmt.Println("write file err!!", err)
		return
	}

	subpath := strings.SplitAfter(url, `/`)
	filename := `d:/red/` + subpath[len(subpath)-1]
	//filename := `./red/` + subpath[len(subpath)-1]
	filename = CheckFileExist(filename, 0)
	f, err := os.Create(filename)

	if err != nil {
		//fmt.Println("write file err!!", err)
		return
	}
	defer f.Close()
	f.Write(buf[0:])
}

//检查文件是否已存在，给重复文件名增加后缀
func CheckFileExist(filename string, count int) string {
	esf, err := os.Stat(filename)
	if err == nil && !esf.IsDir() {
		count++
		subindex := strings.LastIndex(filename, `.`)
		newname := filename[0:subindex] + `(` + strconv.Itoa(count) + `)` + filename[subindex:]
		return CheckFileExist(newname, count)
	}
	return filename
}

//检查url合法性
func Check(url string) (string, bool) {

	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url, true
	}
	if strings.HasPrefix(url, `//`) {

		return "http:" + url, true
	}
	return url, false
}

func CheckNotExist(nexturl string, dep int) bool {
	if _, ok := Rec.Visited[nexturl]; ok {
		atomic.AddInt64(&ReuseCount, 1)
		return false
	}

	Rec.Loc.Lock()

	if _, ok := Rec.Visited[nexturl]; ok {
		Rec.Loc.Unlock()
		atomic.AddInt64(&ReuseCount, 1)
		return false
	}
	Rec.Visited[nexturl] = true
	Rec.Loc.Unlock()

	curl <- CheckNode{nexturl, dep + 1}
	return true
}

//func ParseJS(jscontent string) {
//	program, err := parser.ParseFile(nil, "", jscontent, 0)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Println(program.Body)
//}
