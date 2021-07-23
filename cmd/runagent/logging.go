/*
© Copyright IBM Corporation 2020, 2021

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
package main

import (
	"context"
	"strconv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
//	"os/exec"
	"strings"
	"sync"	
	"time"
//	"github.com/ibm-messaging/mq-container-mft/cmd/mqcont/utilities/command"
	"github.com/ibm-messaging/mq-container-mft/cmd/mqcont/pkg/logger"
	"github.com/antchfx/xmlquery"
)

/*
 * This file contains source code for logging of events from agent
 * container.
 */
var eventLog *logger.Logger

var collectDiagOnFail = false

func logTerminationf(format string, args ...interface{}) {
	logTermination(fmt.Sprintf(format, args...))
}

func logTermination(args ...interface{}) {
	msg := fmt.Sprint(args...)
	// Write the message to the termination log.  This is not the default place
	// that Kubernetes will look for termination information.
	eventLog.Debugf("Writing termination message: %v", msg)
	err := ioutil.WriteFile("/run/termination-log", []byte(msg), 0660)
	if err != nil {
		eventLog.Debug(err)
	}
	eventLog.Error(msg)
}

func getLogFormat() string {
	// We only support basic type - i.e. whatever the format that is logged in 
	// agent output0.log file.
	return "basic"
}

// formatBasic formats a log message parsed from JSON, as "basic" text
func formatBasic(obj map[string]interface{}) string {
	xmlString := fmt.Sprintf("%s", obj["message"]);
	if strings.Contains(xmlString, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"){
		// If this is a transfer log, then format it.
		if strings.Contains(xmlString, "<transaction version") {
			return formatTransferXML(xmlString)
		} else if strings.Contains(xmlString, "<monitorLog version") {
			return ""
			//return formatMonitorXML(xmlString)
		} else {
			return ""
		}
	} else {
		if strings.HasPrefix(xmlString, "#") {
			return ""
		} else {
			return fmt.Sprintf("%s\n", obj["message"])
		}
	}
}

// Format time into given format
func getFormattedTime(timeValue string) string {
	format :="02/01/2006 15:04:05.000"
	layout := "2021-06-22T11:11:37.138Z"
	t, err := time.Parse(layout, timeValue)
	if err != nil {
		return fmt.Sprintf("%s", t.Format(format))
	} else {
		return timeValue
	}
}

// Parse monitor XML messages into a simple readable text.
func formatMonitorXML(xmlString string) string {
	tokens := strings.Split(xmlString, "!")
	var xmlData string
	var supplementText string
	
	if len(tokens) == 2 {
		xmlData = tokens[1]
	} else if len(tokens) == 3 {
		xmlData = tokens[2]
	} else {
		xmlData = ""
	}
	
	doc, erro := xmlquery.Parse(strings.NewReader(xmlData))
	var parsedData string
	if erro != nil {
		eventLog.Printf("Failed to parse XML %v\n", erro);
	} else {
		statusCodeText := xmlquery.FindOne(doc, "//monitorLog/status/@resultCode").InnerText()
		statusCode, err := strconv.Atoi(statusCodeText)
		if err == nil {
			if statusCode != 0 {
				supplements := xmlquery.Find(doc, "//monitorLog/status/supplement")
				for _, supplement := range supplements {
					supplementText += supplement.InnerText() + " "
				}
				parsedData = fmt.Sprintf("[%s] %s\n", 
				getFormattedTime(xmlquery.FindOne(doc, "//monitorLog/action/@time").InnerText()),
				supplementText)
			} else {
				parsedData = fmt.Sprintf("[%s] Monitor %s %s\n", 
				getFormattedTime(xmlquery.FindOne(doc, "//monitorLog/action/@time").InnerText()),
				xmlquery.FindOne(doc, "//monitorLog/@monitorName").InnerText(),
				xmlquery.FindOne(doc, "//monitorLog/action").InnerText())
			}
		}
	}
	
	return parsedData
}

// Parse the transfer XML message and return simple text
func formatTransferXML(xmlMessage string) string {
	var parsedData string
    // Replace all &quot; with single quote
    strings.ReplaceAll(xmlMessage, "&quot;", "'")
    // Create an parsed XML document
    doc, err := xmlquery.Parse(strings.NewReader(xmlMessage))
    if err != nil {
        return parsedData
    }

    // Get required elements from Xml message
    transaction := xmlquery.FindOne(doc, "//transaction")
    if transaction != nil {
        transferId := transaction.SelectAttr("ID")
        if action := transaction.SelectElement("action"); action != nil {
            if strings.EqualFold(action.InnerText(),"completed") {
                var supplementMsg string
                status := transaction.SelectElement("status")
                if status != nil {
                    supplementMsg = status.SelectElement("supplement").InnerText()
                    parsedData = fmt.Sprintf("[%s] TransferID: %s Status: %s\n \tSupplement: %s\n",
                        action.SelectAttr("time"),
                        strings.ToUpper(transferId),
                        action.InnerText(),
                        supplementMsg )
                }

                destAgent := transaction.SelectElement("destinationAgent")
                var actualStartTimeText string = ""
                statistics := transaction.SelectElement("statistics")
                if statistics != nil {
                    actualStartTime := statistics.SelectElement("actualStartTime")
                    if actualStartTime != nil {
                        actualStartTimeText = actualStartTime.InnerText()
                    }
                }
                var retryCount string
                var numFileFailures string
                var numFileWarnings string
                if statistics != nil {
                    if statistics.SelectElement("retryCount") != nil {
                        retryCount = statistics.SelectElement("retryCount").InnerText()
                    }
                    if statistics.SelectElement("numFileFailures") != nil {
                        numFileFailures = statistics.SelectElement("numFileFailures").InnerText()
                    }
                    if statistics.SelectElement("numFileWarnings") != nil {
                        numFileWarnings = statistics.SelectElement("numFileWarnings").InnerText()
                    }
                }
                parsedData += fmt.Sprintf("\tDestination Agent: %s\n\tStart time: %s\n\tCompletion Time: %s\n\tRetry Count: %s\n\tFailures:%s\n\tWarnings:%s\n",
                            destAgent.SelectAttr("agent"),
                            actualStartTimeText,
                            action.SelectAttr("time"),
                            retryCount,
                            numFileFailures,
                            numFileWarnings)
            } else if strings.EqualFold(action.InnerText(),"progress") {
                destAgent := transaction.SelectElement("destinationAgent")
                parsedData += fmt.Sprintf("[%s] %s Status: %s Destination: %s \n", action.SelectAttr("time"),
                        strings.ToUpper(transferId),
                        action.InnerText(),
                        destAgent.SelectAttr("agent"))
                transferSet := transaction.SelectElement("transferSet")
                parsedData += fmt.Sprintf("\tTotal items in transfer request: %s\n\tBytes sent: %s\n", transferSet.SelectAttr("total"),transferSet.SelectAttr("bytesSent"))
                items := transferSet.SelectElements("item")
                for i := 0 ; i < len(items); i++ {
                    status := items[i].SelectElement("status")
                    resultCode := status.SelectAttr("resultCode")
                    var sourceName string
                    queueSource := items[i].SelectElement("source/queue")
                    if queueSource != nil {
                        sourceName = queueSource.InnerText()
                    } else {
                        fileName := items[i].SelectElement("source/file")
                        if fileName != nil {
                            sourceName = fileName.InnerText()
                        }
                    }

                    var destinationName string
                    queueDest := items[i].SelectElement("destination/queue")
                    if queueDest != nil {
                        destinationName = queueDest.InnerText()
                    } else {
                        fileName := items[i].SelectElement("destination/file")
                        if fileName != nil {
                           destinationName = fileName.InnerText()
                        }
                    }

                    parsedData += fmt.Sprintf("\tItem # %d\n\t\tSource: %s\n\t\tDestination: %s\n",i+1, sourceName, destinationName)
                    if resultCode != "0" {
                        supplement := status.SelectElement("supplement")
                        if supplement != nil {
                            parsedData += fmt.Sprintf("\t\tResult code %s Supplement %s\n", resultCode, supplement.InnerText())
                        } else {
                            parsedData += fmt.Sprintf("\t\tResult code %s\n", resultCode)
                        }
                    } else {
                        parsedData += fmt.Sprintf("\t\tResult code %s\n", resultCode)
                    }
                }
            } else if strings.EqualFold(action.InnerText(),"started") {
                destAgent := transaction.SelectElement("destinationAgent")
                destinationAgentName := destAgent.SelectAttr("agent")
                parsedData += fmt.Sprintf("[%s] TransferID: %s Status: %s Destination: %s\n",
                        action.SelectAttr("time"),
                        strings.ToUpper(transferId),
                        action.InnerText(),
                        destinationAgentName)
            }
        } // Action
    } // Transaction != null
	
	return parsedData
}


// mirrorAgentEventLogs starts a goroutine to mirror the contents of the agent logs
func mirrorAgentEventLogs(ctx context.Context, wg *sync.WaitGroup, logFilePath string, fromStart bool, mf mirrorFunc) (chan error, error) {
	// Use the current format agent output log format.
	return mirrorLog(ctx, wg, logFilePath, fromStart, mf)
}

func getDebug() bool {
	// No debug logs supported for now. Simply return false.
	return false
}

// Setup logger to capture events.
func configureLogger(name string) (mirrorFunc, error) {
	var err error
	f := getLogFormat()
	d := getDebug()
	switch f {
	case "basic":
		eventLog, err = logger.NewLogger(os.Stderr, d, false, name)
		if err != nil {
			return nil, err
		}
		return func(msg string) bool {
			// Parse the JSON message, and print a simplified version
			obj, err := processLogMessage(msg)
			if err != nil {
				eventLog.Printf("Failed to unmarshall JSON - %v", err)
			} else {
				fmt.Printf(formatBasic(obj))
			}
			return true
		}, nil
	default:
		eventLog, err = logger.NewLogger(os.Stdout, d, false, name)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("invalid value for LOG_FORMAT: %v", f)
	}
}
// Process log messages from agent's output0.log and others.
func processLogMessage(msg string) (map[string]interface{}, error) {
	var obj map[string]interface{}
	// Replace all double quotes with an escape character so that JSON marshalling
	// does not run into problems.
	escapedMsg := strings.Replace(msg, "\"","\\\"", -1)
	// Make a JSON message that contains only one attribute - a single line from
	// output0.log file. Also replace all single quote with double quotes
	jsonMsg := "{\"message\":\"" + escapedMsg + "\"}"
	err := json.Unmarshal([]byte(jsonMsg), &obj)
	return obj, err
}

