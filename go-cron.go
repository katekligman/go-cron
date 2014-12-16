package main

import "os"
import "os/exec"
import "strings"
import "sync"
import "os/signal"
import "syscall"
import "github.com/robfig/cron"
import "log"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

//var last_err error
var last_err LastRun

type LastRun struct {
	Exit_status int
	//stdout      bytes.Buffer
	//stderr      bytes.Buffer
	Stdout string
	Stderr string
}

func execute(command string, args []string) {

	log.Println("executing:", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	//last_err.stdout.Reset()
	//last_err.stderr.Reset()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start: %v")
	}
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			last_err.Exit_status = 0

			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				last_err.Exit_status = status.ExitStatus()
				log.Printf("Exit Status: %d", last_err.Exit_status)
			}
			last_err.Stderr = stderr.String()
			last_err.Stdout = stdout.String()
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}
}

func create() (cr *cron.Cron, wgr *sync.WaitGroup) {
	var schedule string = os.Args[1]
	var command string = os.Args[2]
	var args []string = os.Args[3:len(os.Args)]

	wg := &sync.WaitGroup{}

	c := cron.New()
	log.Println("new cron:", schedule)

	c.AddFunc(schedule, func() {
		wg.Add(1)
		execute(command, args)
		wg.Done()
	})

	return c, wg
}

func handler(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(last_err)
	resp := string(b[:len(b)])
	if last_err.Exit_status != 0 {
		http.Error(w, resp, http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, resp)
}

func http_server(c *cron.Cron, wg *sync.WaitGroup) {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":18080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func start(c *cron.Cron, wg *sync.WaitGroup) {
	c.Start()
}

func stop(c *cron.Cron, wg *sync.WaitGroup) {
	log.Println("Stopping")
	c.Stop()
	log.Println("Waiting")
	wg.Wait()
	log.Println("Exiting")
	os.Exit(0)
}

func main() {

	c, wg := create()

	go start(c, wg)
	go http_server(c, wg)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	println(<-ch)
	stop(c, wg)
}