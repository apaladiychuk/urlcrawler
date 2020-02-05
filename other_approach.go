package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func mainO() {
	startTime := time.Now()

	var filePath = flag.String("f", "urls.txt", "Enter the full path to your text file of URLs. Default: input.txt")
	var threads = flag.Int("g", 200, "Number of goroutines. Default:100")

	urlsChan := make(chan string, 100)
	urlsDataChan := make(chan string, 100)

	go func(filePath string) {
		file, err := os.Open(filePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var url string
		for scanner.Scan() {
			url = scanner.Text()
			if url != "" {
				urlsChan <- url
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		close(urlsChan)
		_ = file.Close()

	}(*filePath)

	var wg sync.WaitGroup
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go scan(&wg, urlsChan, urlsDataChan)
	}

	f, err := os.Create("output.csv")
	if err != nil {
		fmt.Println(err)
		return
	}

	var wgWrite sync.WaitGroup
	wgWrite.Add(1)
	go func(f *os.File, wg *sync.WaitGroup) {
		for urlData := range urlsDataChan {
			_, _ = f.WriteString(urlData + "\n")
		}
		wg.Done()
	}(f, &wgWrite)

	wg.Wait()
	close(urlsDataChan)

	wgWrite.Wait()

	_ = f.Close()

	fmt.Printf("%d", time.Now().Sub(startTime).Milliseconds())
	fmt.Scanf("h")

}

func scan(wg *sync.WaitGroup, urlsChan chan string, urlsDataChan chan string) {
	defer wg.Done()
	for url := range urlsChan {
		urlsDataChan <- scanUrl(url)
	}
}

func scanUrl(url string) string {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Get(url)
	if err != nil {
		return url + ";;"
	}

	body, _ := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()
	return url + ";" + strconv.Itoa(resp.StatusCode) + ";" + strconv.Itoa(len(body))
}
