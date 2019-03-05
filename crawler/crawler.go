package main

import (
	"fmt"
	"sync"
)

// Fetcher interface
type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// AlreadyFetchedError is returned when a url has already been
// fetched or is in progress of being fetched
type AlreadyFetchedError struct {
	url string
}

func (e AlreadyFetchedError) Error() string {
	return fmt.Sprintf("Already fetched %v", e.url)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, c chan<- fakeResult, wg *sync.WaitGroup) {
	defer wg.Done()
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)

	if _, ok := err.(*AlreadyFetchedError); ok {
		return
	} else if err != nil {
		fmt.Println(err)
		return
	}

	c <- fakeResult{body, urls}

	fmt.Printf("found: %s %q\n", url, body)
	for _, u := range urls {
		wg.Add(1)
		go Crawl(u, depth-1, fetcher, c, wg)
	}
	return
}

func main() {
	var c = make(chan fakeResult)

	var f = myFetcher{
		cache: make(map[string]bool),
		mux:   &sync.Mutex{},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go Crawl("https://golang.org/", 4, f, c, &wg)

	// close c when all crawlers are done
	go func() {
		wg.Wait()
		close(c)
	}()

	for r := range c {
		fmt.Println(r)
	}
}

type myFetcher struct {
	cache map[string]bool
	mux   *sync.Mutex
}

type fakeResult struct {
	body string
	urls []string
}

func (f myFetcher) Fetch(url string) (string, []string, error) {
	f.mux.Lock()
	if _, ok := f.cache[url]; ok {
		f.mux.Unlock()
		return "", nil, &AlreadyFetchedError{url}
	}
	// insert into cache so nobody else gets this url
	f.cache[url] = true
	f.mux.Unlock()

	if res, ok := rawData[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var rawData = map[string]*fakeResult{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
