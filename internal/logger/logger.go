/*
© Copyright IBM Corporation 2018

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package logger provides utility functions for logging purposes
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"sync"
	"time"
)

// timestampFormat matches the format used by MQ messages (includes milliseconds)
const timestampFormat string = "2006-01-02T15:04:05.000Z07:00"
const debugLevel string = "DEBUG"
const infoLevel string = "INFO"
const errorLevel string = "ERROR"

// A Logger is used to log messages to stdout
type Logger struct {
	mutex       sync.Mutex
	writer      io.Writer
	debug       bool
	json        bool
	processName string
	pid         int
	serverName  string
	host        string
	user        *user.User
}

// NewLogger creates a new logger
func NewLogger(writer io.Writer, debug bool, json bool, serverName string) (*Logger, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	return &Logger{
		mutex:       sync.Mutex{},
		writer:      writer,
		debug:       debug,
		json:        json,
		processName: os.Args[0],
		pid:         os.Getpid(),
		serverName:  serverName,
		host:        hostname,
		user:        user,
	}, nil
}

func (l *Logger) format(entry map[string]interface{}) (string, error) {
	if l.json {
		b, err := json.Marshal(entry)
		if err != nil {
			return "", err
		}
		return string(b), err
	}
	return fmt.Sprintf("%v %v\n", entry["ibm_datetime"], entry["message"]), nil
}

// log logs a message at the specified level.  The message is enriched with
// additional fields.
func (l *Logger) log(level string, msg string) {
	t := time.Now()
	entry := map[string]interface{}{
		"message":         fmt.Sprint(msg),
		"ibm_datetime":    t.Format(timestampFormat),
		"loglevel":        level,
		"host":            l.host,
		"ibm_serverName":  l.serverName,
		"ibm_processName": l.processName,
		"ibm_processId":   l.pid,
		"ibm_userName":    l.user.Username,
		"type":            "mq_containerlog",
	}
	s, err := l.format(entry)
	l.mutex.Lock()
	if err != nil {
		// TODO: Fix this
		fmt.Println(err)
	}
	if l.json {
		fmt.Fprintln(l.writer, s)
	} else {
		fmt.Fprint(l.writer, s)
	}
	l.mutex.Unlock()
}

// LogDirect logs a message directly to stdout
func (l *Logger) LogDirect(msg string) {
	fmt.Println(msg)
}

// Debug logs a line as debug
func (l *Logger) Debug(args ...interface{}) {
	if l.debug {
		l.log(debugLevel, fmt.Sprint(args...))
	}
}

// Debugf logs a line as debug using format specifiers
func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.debug {
		l.log(debugLevel, fmt.Sprintf(format, args...))
	}
}

// Print logs a message as info
func (l *Logger) Print(args ...interface{}) {
	l.log(infoLevel, fmt.Sprint(args...))
}

// Println logs a message
func (l *Logger) Println(args ...interface{}) {
	l.Print(args...)
}

// Printf logs a message as info using format specifiers
func (l *Logger) Printf(format string, args ...interface{}) {
	l.log(infoLevel, fmt.Sprintf(format, args...))
}

// PrintString logs a string as info
func (l *Logger) PrintString(msg string) {
	l.log(infoLevel, msg)
}

// Errorf logs a message as error
func (l *Logger) Error(args ...interface{}) {
	l.log(errorLevel, fmt.Sprint(args...))
}

// Errorf logs a message as error using format specifiers
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(errorLevel, fmt.Sprintf(format, args...))
}

// Fatalf logs a message as fatal using format specifiers
// TODO: Remove this
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log("FATAL", fmt.Sprintf(format, args...))
}
