package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	joiner "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valyala/gorpc"
)

// ActionStruct struct
type ActionStruct struct {
	Action struct {
		Action *string `json:"action"`
	}
	mu sync.Mutex
}

func main() {
	as := &ActionStruct{}
	e := echo.New()
	e.Use(middleware.Logger())
	e.Router().Add(http.MethodGet, "/", as.Handle)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: e,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("starting server error: %v", err)
		}
	}()

	c := &gorpc.Client{Addr: "server:9090"}
	c.Start()

	var oldJSON string

	for {
		resp, err := c.CallTimeout("", time.Second)
		if err != nil {
			continue
		}
		if resp != nil {
			if oldJSON != resp.(string) {
				as.mu.Lock()
				_ = joiner.Unmarshal([]byte(resp.(string)), &as.Action)
				as.mu.Unlock()
				oldJSON = resp.(string)
				e.Logger.Printf("%v", as.Action)
			}
		}
	}
}

// Handle gives action
func (as *ActionStruct) Handle(c echo.Context) error {
	for as.Action.Action == nil {
		time.Sleep(time.Millisecond)
	}
	defer func() { as.mu.Lock(); as.Action.Action = nil; as.mu.Unlock() }()
	return c.String(http.StatusOK, *as.Action.Action)
}
