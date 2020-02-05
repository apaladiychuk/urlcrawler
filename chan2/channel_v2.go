package chan2

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const numberOfThreads = 200

// start function
func MainChannel() {
	startTime := time.Now().Unix()
	var result [][]string
	var cnt = 0
	var outC = make(chan []string, 100)
	var inC = make(chan []string, 100)
	var wg = &sync.WaitGroup{}
	// array of runner
	var runner = make([]Parcer, numberOfThreads)
	for i := 0; i < numberOfThreads; i++ {
		runner[i] = new(outC, wg)
		go runner[i].scrapeUrl(inC)
	}

	file, err := os.Open("majestic_million.csv")
	if err != nil {
		return
	}
	defer file.Close()

	r := csv.NewReader(file)
	if _, err := r.Read(); err != nil {
		log.Fatalln(err)
	}

	// thread for saving data
	// possible to  write in  file in this thread
	go func() {
		cmp := 0
		for l := range outC {
			if len(l) == 0 {
				cmp++
				if cmp == numberOfThreads {
					break
				}
			} else {
				result = append(result, l)
			}
		}
		// finish writer thread
		wg.Done()
	}()

	// for waiting writer thread
	wg.Add(1)
	// main loop
	for {
		row, err := r.Read()
		if err == io.EOF {
			//Break if end of file
			break
		}
		cnt++
		if cnt >= 1000 {
			break
		}
		wg.Add(1)
		inC <- row
	}
	// finalize parsers threads
	close(inC)
	wg.Wait()
	// write result
	fileO, err := os.Create("result.csv")
	defer fileO.Close()
	writer := csv.NewWriter(fileO)
	defer writer.Flush()
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range result {
		if err := writer.Write(r); err != nil {
		}
	}
	fmt.Println("-----------------------------------")
	fmt.Printf(" total time %d \n", time.Now().Unix()-startTime)
	fmt.Printf(" total valid cnt  %d \n", len(result))
}

type Parcer struct {
	c      chan []string
	client *http.Client
	outC   chan []string
	wg     *sync.WaitGroup
}

func (p *Parcer) scrapeUrl(c chan []string) {
	defer p.wg.Done()
	for url := range c {
		if resp, err := p.client.Get("http://" + url[2]); err != nil {
			//p.outC <- fmt.Sprintf("[REQUEST ERR ] %v " , err.Error())
			fmt.Printf("[ERR] %v\n", err)
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			l := len(body)
			code := resp.StatusCode
			resp.Body.Close()
			body = nil
			url = append(url, fmt.Sprint("%d", code), fmt.Sprintf("%d", l))
			p.outC <- url
		}
	}
}

func new(o chan []string, w *sync.WaitGroup) Parcer {
	return Parcer{
		c: make(chan []string, 5),
		client: &http.Client{
			// Parameters from requirement
			Timeout: 10 * time.Second,
		},
		outC: o,
		wg:   w,
	}
}

func (p *Parcer) Finish() {
	v := []string{}
	p.c <- v
}
