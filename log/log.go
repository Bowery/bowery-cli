// Copyright 2013-2014 Bowery, Inc.
// Package log provides routines to log and print debug messages.
package log

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Bowery/cli/api"
	"github.com/Bowery/cli/db"

	redigo "github.com/garyburd/redigo/redis"
)

var (
	debug     = os.Getenv("DEBUG")
	env       = os.Getenv("ENV")
	host      = os.Getenv("HOST")
	logWriter *LogWriter
)

func init() {
	// Attempt to get developer.
	devId := "0"
	dev, _ := db.GetDeveloper()
	if dev != nil && dev.Developer != nil {
		devId = dev.Developer.ID
	}

	logWriter = NewLogWriter(devId, api.RedisPath)
}

// Debug prints the given arguments if the ENV var is set to development.
func Debug(args ...interface{}) {
	if debug == "cli" {
		Fprint(os.Stderr, "cyan", "DEBUG: ")
		Fprintln(os.Stderr, "", args...)
	}
}

// Print prints arguments with the given attributes, to stdout.
func Print(attrs string, args ...interface{}) {
	go logWriter.Write(args)
	Fprint(os.Stdout, attrs, args...)
}

// Fprint prints arguments with the given attributes, to a writer.
func Fprint(w io.Writer, attrs string, args ...interface{}) {
	attrList := strings.Split(attrs, " ")
	for _ = range attrList {
		args = append(args, noAttr)
	}

	fmt.Fprint(w, getColor(attrList[0]))
	if len(attrList) > 1 {
		fmt.Fprint(w, getAttr(attrList[1]))
	}

	go logWriter.Write(args)
	fmt.Fprint(w, args...)
}

// Println prints arguments with the given attributes, to stdout.
func Println(attrs string, args ...interface{}) {
	Fprintln(os.Stdout, attrs, args...)
}

// Fprintln prints arguments with the given attributes, to a writer.
func Fprintln(w io.Writer, attrs string, args ...interface{}) {
	attrList := strings.Split(attrs, " ")
	for _ = range attrList {
		args = append(args, noAttr)
	}

	fmt.Fprint(w, getColor(attrList[0]))
	if len(attrList) > 1 {
		fmt.Fprint(w, getAttr(attrList[1]))
	}

	go logWriter.Write(args)
	fmt.Fprintln(w, args...)
}

type LogWriter struct {
	id   string       // unique identifier
	pool *redigo.Pool // redis conn pool
}

func NewLogWriter(id, addr string) *LogWriter {
	return &LogWriter{
		id: id,
		pool: &redigo.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redigo.Conn, error) {
				return redigo.Dial("tcp", addr)
			},
		},
	}
}

func (lw *LogWriter) Write(data interface{}) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return err
	}
	conn := lw.pool.Get()
	defer conn.Close()
	return conn.Send("PUBLISH", "console:"+lw.id, string(buf.Bytes()))
}

func (lw *LogWriter) Close() error {
	return lw.pool.Close()
}