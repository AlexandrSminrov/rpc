package main

import (
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/valyala/gorpc"
)

// ResponseStruct struct
type ResponseStruct struct {
	NewJSON []byte
	mu      sync.Mutex
}

func main() {
	rs := &ResponseStruct{}
	s := &gorpc.Server{
		Addr:    ":9090",
		Handler: rs.Handle,
	}

	ch1 := make(chan byte, 1)
	ch2 := make(chan byte, 1)
	ch1 <- 1
	ch2 <- 1

	go func() {
		for {
			select {
			case <-ch1:
				go rs.Worker(ch1, "https://novasite.su/test1.php")
			case <-ch2:
				go rs.Worker(ch2, "https://novasite.su/test2.php")
			}
		}
	}()

	log.Println("start server...")

	if err := s.Serve(); err != nil {
		log.Fatalf("Cannot start rpc server: %s", err)
	}
}

// Handle gives data
func (rs *ResponseStruct) Handle(_ string, _ interface{}) interface{} {
	return string(rs.NewJSON)
}

// Worker makes requests to services
func (rs *ResponseStruct) Worker(ch chan byte, url string) {
	defer func() { ch <- byte(1) }()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	time.Sleep(time.Duration(rand.Intn(3)) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Errorf("Reqest error %v", err.Error())
		return
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Errorf("error close body: %v", err)
		}
	}()

	if resp.StatusCode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		if string(rs.NewJSON) != string(bodyBytes) && bodyBytes != nil {
			log.Println(string(bodyBytes))
			rs.mu.Lock()
			rs.NewJSON = bodyBytes
			rs.mu.Unlock()
		}

	}
}
