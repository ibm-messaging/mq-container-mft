/*
© Copyright IBM Corporation 2018, 2019

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
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

// timestampFormat matches the format used by MQ messages (includes milliseconds)
const timestampFormat string = "2006-01-02T15:04:05.000Z07:00"
const debugLevel string = "DEBUG"
const infoLevel string = "INFO"
const errorLevel string = "ERROR"

// A Logger is used to log messages to stdout
type Logger struct {
	mutex           sync.Mutex
	writer          io.Writer
	debug           bool
	json            bool
	processName     string
	pid             string
	serverName      string
	host            string
	userName        string
	logUrl          string
	logKey          string
	logServerType   int16
	logPubsDisabled bool
}

// NewLogger creates a new logger
func NewLogger(writer io.Writer, debug bool, json bool, serverName string, dnaUrl string, dnaKey string, logServType int16) (*Logger, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	// This can fail because the container's running as a random UID which
	// is not known by the OS.  We don't want this to break the logging
	// entirely, so just use a blank user name.
	user, err := user.Current()
	userName := ""
	if err == nil {
		userName = user.Username
	}
	return &Logger{
		mutex:           sync.Mutex{},
		writer:          writer,
		debug:           debug,
		json:            json,
		processName:     os.Args[0],
		pid:             strconv.Itoa(os.Getpid()),
		serverName:      serverName,
		host:            hostname,
		userName:        userName,
		logUrl:          dnaUrl,
		logKey:          dnaKey,
		logServerType:   logServType,
		logPubsDisabled: false,
	}, nil
}

func (l *Logger) format(entry map[string]interface{}) (string, error) {
	//	if l.json {
	//		b, err := json.Marshal(entry)
	//		if err != nil {
	//			return "", err
	//		}
	//		return string(b), err
	//	}
	return fmt.Sprintf("%v\n", entry["message"]), nil
}

// log logs a message at the specified level.  The message is enriched with
// additional fields.
func (l *Logger) log(level string, msg string) {
	entry := map[string]interface{}{
		"message": fmt.Sprint(msg),
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

/*
  Function to publish transfer log to a server
*/
func (l *Logger) PushToLogToServer(msg string) {
	// Don't do anything if log publication has been disabled
	// This can happen if an error occured when an attempt was
	// made to publish logs to server earlier.
	if l.logPubsDisabled == true {
		return
	}

	// Return if this is not a valid JSON
	if !gjson.Valid(msg) {
		return
	}

	// Simply return if The JSON does not contain eventDescription.
	if !gjson.Get(msg, "eventDescription").Exists() {
		return
	}

	if l.logServerType == 1 {
		l.PushToLogToLogDNA(msg)
	} else if l.logServerType == 2 {
		l.PushLogToELK(msg)
	}
}

/*
  Function to publish transfer log to logDNA server
*/
func (l *Logger) PushToLogToLogDNA(msg string) {
	logDNALevel := getLogLevel(msg)
	// Build the payload to publish to server
	lineText := gjson.Get(msg, "transferId").String() + " " + gjson.Get(msg, "eventDescription").String()
	msgToPush := "{\"lines\":[{"
	msgToPush += "\"app\":\"" + l.serverName + "\","
	msgToPush += "\"level\":\"" + logDNALevel + "\","
	msgToPush += "\"line\":\"" + lineText + "\","
	msgToPush += "\"meta\":" + msg + "}]}"
	responseBody := bytes.NewBufferString(msgToPush)
	logDNAUrl := l.logUrl + "?hostname=" + l.host + "&now=" + strconv.FormatInt(int64(time.Now().Nanosecond()), 10)
	reqPOST, errRes := http.NewRequest("POST", logDNAUrl, responseBody)
	if errRes != nil {
		// There was an error creating HTTP request. So return
		utils.PrintLog(fmt.Sprintf("An error occured while creating HTTP request to %s. The error is: %v\n", logDNAUrl, errRes))
		l.logPubsDisabled = true
		return
	}

	// All is well, so set the required headers for logDNA
	reqPOST.Header.Set("Accept", "application/json")
	reqPOST.Header.Set("Content-Type", "application/json")
	reqPOST.Header.Set("apikey", l.logKey)

	logDNAClient := &http.Client{}
	respDNA, errDNA := logDNAClient.Do(reqPOST)
	if errDNA != nil {
		utils.PrintLog(fmt.Sprintf("An error occured while publishing transfer logs to %s. The error is: %v\n", logDNAUrl, errDNA))
		l.logPubsDisabled = true
		return
	}
	defer respDNA.Body.Close()

	//Read the response body
	_, err := ioutil.ReadAll(respDNA.Body)
	if err != nil {
		utils.PrintLog(fmt.Sprintf("An error occurred while reading response from server %s. The error is: %v\n", logDNAUrl, err))
		l.logPubsDisabled = true
		return
	}
}

// Generate a logDNA type level using the transfer log
func getLogLevel(msg string) string {
	// Use INFO as default
	level := "INFO"

	// Mark level as error is eventDescription contains words like error
	// and fail
	eventDescription := gjson.Get(msg, "eventDescription").String()
	if strings.Contains(eventDescription, "error") ||
		strings.Contains(eventDescription, "fail") {
		level = "ERROR"
	}

	// Mark resynchronize messages as warnings.
	if strings.Contains(eventDescription, "BFGTL0015") ||
		strings.Contains(eventDescription, "BFGTL0017") ||
		strings.Contains(eventDescription, "BFGTL0019") ||
		strings.Contains(eventDescription, "BFGTL0021") ||
		strings.Contains(eventDescription, "BFGTL0032") ||
		strings.Contains(eventDescription, "BFGTL0033") ||
		strings.Contains(eventDescription, "BFGTL0034") ||
		strings.Contains(eventDescription, "BFGTL0035") {
		level = "WARN"
	}

	// Check progress information and mark level accordingly
	if gjson.Get(msg, "progressInformation").Exists() {
		progInfo := gjson.Get(msg, "progressInformation")
		if gjson.Get(progInfo.String(), "failed").Int() > 0 {
			level = "ERROR"
		}
		if gjson.Get(progInfo.String(), "warnings").Int() > 0 {
			level = "WARN"
		}
	}

	// Process transfer completed message
	if gjson.Get(msg, "transferCompleted").Exists() {
		progInfo := gjson.Get(msg, "transferCompleted")
		if gjson.Get(progInfo.String(), "failures").Int() > 0 {
			level = "ERROR"
		}
		if gjson.Get(progInfo.String(), "warnings").Int() > 0 {
			level = "WARN"
		}
		if gjson.Get(progInfo.String(), "resultCode").Int() > 0 {
			level = "ERROR"
		}
	}

	return level
}

/**
  Push data a ELK server
*/
func (l *Logger) PushLogToELK(msg string) {
	//elkURL := "localhost:9200/ibmmqmft/tlog"

	// Return if this is not a valid JSON
	if !gjson.Valid(msg) {
		return
	}

	// Simply return if The JSON does not contain eventDescription.
	if !gjson.Get(msg, "eventDescription").Exists() {
		return
	}

	elkLevel := getLogLevel(msg)
	// Build the payload to publish to server
	//lineText := gjson.Get(msg, "transferId").String() + " " + gjson.Get(msg, "eventDescription").String()
	msgToPush := "{\"transferLog\":{"
	msgToPush += "\"hostName\":\"" + l.host + "\","
	msgToPush += "\"level\":\"" + elkLevel + "\","
	msgToPush += "\"metaData\":" + msg + "}]}"
	responseBody := bytes.NewBufferString(msgToPush)
	logUrl := l.logUrl + "/ibmmqmft/" + l.serverName
	reqPOST, errRes := http.NewRequest("POST", strings.ToLower(logUrl), responseBody)
	if errRes != nil {
		// There was an error creating HTTP request. So return
		utils.PrintLog(fmt.Sprintf("An error occured while creating HTTP request to %s. The error is: %v\n", logUrl, errRes))
		l.logPubsDisabled = true
		return
	}

	// All is well, so set the required headers for ELK
	reqPOST.Header.Set("Content-Type", "application/json")
	logHTTPClient := &http.Client{}
	respPost, errPost := logHTTPClient.Do(reqPOST)
	if errPost != nil {
		utils.PrintLog(fmt.Sprintf("An error occured while publishing transfer logs to %s. The error is: %v\n", logUrl, errPost))
		l.logPubsDisabled = true
		return
	}
	defer respPost.Body.Close()

	//Read the response body
	_, err := ioutil.ReadAll(respPost.Body)
	if err != nil {
		utils.PrintLog(fmt.Sprintf("An error occurred while reading response from server %s. The error is: %v\n", logUrl, err))
		l.logPubsDisabled = true
		return
	}
}
