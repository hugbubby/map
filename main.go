package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func clog(msg ...interface{}) {
	log.Println("map: ", msg)
}

func runcmd(cmd_s string, done chan struct{}) {
	cmd := exec.Command("/bin/bash", "-c", cmd_s)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		clog(err)
	}
	done <- struct{}{}
}

var threads int

func main() {
	var replacement, delimiter string
	flag.StringVar(&replacement, "r", "_MAP_", "The string to replace in your command with the array's members.")
	flag.StringVar(&delimiter, "d", " ", "The delimiter for your array.")
	flag.IntVar(&threads, "t", 1, "The number of threads to use for cmd replacement; default 1. Currently unimplemented.")
	flag.Parse()

	delimiter = strings.ReplaceAll(delimiter, "\\n", "\n")
	delimiter = strings.ReplaceAll(delimiter, "\\t", "\t")
	delimiter = strings.ReplaceAll(delimiter, "\\r", "\r")

	if flag.NArg() != 2 {
		fmt.Println("Incorrect number of arguments.")
		os.Exit(1)
	}

	array := strings.Split(flag.Arg(0), delimiter)
	cmd_s := flag.Arg(1)

    //Small Queue.
	done := make(chan struct{}, threads)
    numJobs := len(array)
    jobsFinished := 0
    for i := 0; i < threads && len(array) > 0; i++ {
        cmd_replaced := strings.ReplaceAll(cmd_s, replacement, array[0])
        go runcmd(cmd_replaced, done)
        array = array[1:]
    }
	for len(array) > 0 {
        <-done
        jobsFinished++
        cmd_replaced := strings.ReplaceAll(cmd_s, replacement, array[0])
        go runcmd(cmd_replaced, done)
        array = array[1:]
	}
    for jobsFinished < numJobs {
        <-done
        jobsFinished++
    }
}
