package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	flag "github.com/spf13/pflag"
)
var quiet bool

func clog(msg string) {
	log.Println("map: ", msg)
}

func runcmd(cmd_s string, done chan struct{}) {
	cmd := exec.Command("/bin/bash", "-c", cmd_s)
	if !quiet {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
	}
	if err := cmd.Run(); err != nil {
		clog(err.Error())
	}
	done <- struct{}{}
}

func main() {
	var wordlists []string
	var threads int
	var replacename, delimiter string

	flag.StringArrayVarP(&wordlists, "wordlist", "w", []string{}, "The wordlist to run commands from. Can be used multiple times to specify multiple wordlists.")
	flag.StringVarP(&replacename, "replacename", "r", "GMAP", "The string to replace in your command with the array's members.")
	flag.StringVarP(&delimiter, "delimiter", "d", " ", "The delimiter for your array.")
	flag.IntVarP(&threads, "threads", "t", 1, "The number of threads to use for cmd replacement; default 1.")
	flag.BoolVarP(&quiet, "quiet", "q", false, "Whether or not to quiet output.")

	flag.Parse()

	delimiter = strings.ReplaceAll(delimiter, "\\n", "\n")
	delimiter = strings.ReplaceAll(delimiter, "\\t", "\t")
	delimiter = strings.ReplaceAll(delimiter, "\\r", "\r")

	if flag.NArg() == 0 {
		fmt.Println("Must specify entire comamnd to run in the single argument given to his program!")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if flag.NArg() > 1 {
		fmt.Println("Must specify only one argument - the command to map over!")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if len(wordlists) < 1 {
		clog("Must specify at least one wordlist to map over the command!")
		flag.PrintDefaults()
		os.Exit(1)
	}
	array := strings.Split(wordlists[0], delimiter)
	cmd_s := flag.Arg(0)

	//Small Queue.
	done := make(chan struct{}, threads)
	numJobs := len(array)
	jobsFinished := 0
	for i := 0; i < threads && len(array) > 0; i++ {
		cmd_replaced := strings.ReplaceAll(cmd_s, replacename, array[0])
		go runcmd(cmd_replaced, done)
		array = array[1:]
	}
	for len(array) > 0 {
		<-done
		jobsFinished++
		cmd_replaced := strings.ReplaceAll(cmd_s, replacename, array[0])
		go runcmd(cmd_replaced, done)
		array = array[1:]
	}
	for jobsFinished < numJobs {
		<-done
		jobsFinished++
	}
}
