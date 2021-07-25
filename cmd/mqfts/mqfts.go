package main
/*
************************************************************************
* This file contains the source code for IBM MQ Managed File Transfer
* Log Capture parse utility
*
************************************************************************
* © Copyright IBM Corporation 2021, 2021

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
************************************************************************
*/

import (
    "bufio"
    "bytes"
    "fmt"
    "github.com/antchfx/xmlquery"
    "github.com/icza/backscanner"
    flag "github.com/spf13/pflag"
    "io"
    "log"
    "os"
    "strings"
    "time"

    "github.com/tidwall/gjson"
    "github.com/ibm-messaging/mq-container-mft/cmd/utils"
)

// Map to cache transfer ids already processed
var transferIdMap map[string]string

var displayCount int
var displayTransferType int
var counter int
const transferSUCCESSFUL = 1
const transferPARTIALSUCCESS = 2
const transferFAILED = 3
const transferSTARTED = 4
const transferINPROGRESS = 5

// Main entry point
func main() {
    var failedTransfers int
    var successTransfers int
    var partSuccessTransfers int
    var startedTransfers int
    var inProgressTransfers int
    var logFilePath string
    var transferId string

    fmt.Printf("IBM MQ Managed File Transfer Status Utility\n")

    flag.StringVar(&logFilePath, "lf", "", "Capture log file path")
    flag.Lookup("lf").NoOptDefVal = ""

    flag.StringVar(&transferId, "id", "", "Transfer ID")
    flag.Lookup("id").NoOptDefVal = ""

    flag.IntVar(&successTransfers, "sf", -1, "Display successful transfers")
    flag.Lookup("sf").NoOptDefVal = "-1"

    flag.IntVar(&partSuccessTransfers, "ps", -1, "Display partially successful transfers")
    flag.Lookup("ps").NoOptDefVal = "-1"

    flag.IntVar(&failedTransfers, "fl", -1, "Display failed transfers")
    flag.Lookup("fl").NoOptDefVal = "-1"

    flag.IntVar(&startedTransfers, "st", -1, "Display 'started' transfers")
    flag.Lookup("st").NoOptDefVal = "-1"

    flag.IntVar(&inProgressTransfers, "ip", -1, "Display 'In Progress' transfers")
    flag.Lookup("ip").NoOptDefVal = "-1"

    flag.Usage = func() {
        displayHelp()
        return
    }

    // Parse the provided command line
    flag.Parse()

    // Display usage if we have some unknown parameters
    if len(flag.Args()) > 0 {
        flag.Usage()
        return
    }

    // Get capture log file path
    var outputLogFilePath = getLogPath(logFilePath)
    fmt.Printf("\nDisplaying transfer details from %s\n\n", outputLogFilePath)

    if isFlagPassed("sf") {
        displayCount = successTransfers
        displayTransferType = transferSUCCESSFUL
    }

    if isFlagPassed("ps") {
        displayCount = partSuccessTransfers
        displayTransferType = transferPARTIALSUCCESS
    }

    if isFlagPassed("fl") {
        displayCount = failedTransfers
        displayTransferType = transferFAILED
    }

    if isFlagPassed("st") {
        displayCount = startedTransfers
        displayTransferType = transferSTARTED
    }

    if isFlagPassed("ip") {
        displayCount = inProgressTransfers
        displayTransferType = transferINPROGRESS
    }

    // Initialize map for storing transfer ids that have been processed.
    transferIdMap = make(map[string]string, 1)
    if transferId != "" {
        parseAndDisplayTransfer (outputLogFilePath, transferId)
    } else {
        // Display transfer status
        displayTransferStatus(outputLogFilePath)
    }
}

func isFlagPassed(name string) bool {
    found := false
    flag.Visit(func(f *flag.Flag) {
        if f.Name == name {
            found = true
        }
    })
    return found
}

func displayHelpSample() {
    dispUsage := "\nUsage:\n\n"
    dispUsage += "  Specify either:\n"
    dispUsage += "      --lf <capture log file>\n\n"
    dispUsage += "	  OR\n\n"
    dispUsage += "      Set the following environment variables\n"
    dispUsage += "         MFT_CAPTURE_LOG_PATH=<path of MFT log capture file>\n\n"
    dispUsage += "	  OR\n\n"
    dispUsage += "      Set the following environment variables\n"
    dispUsage += "         MFT_AGENT_NAME=<name of your agent>\n"
    dispUsage += "         MFT_COORDINATION_QM=<coordination queue manager name>\n"
    dispUsage += "         BFG_DATA=<MFT data directory>\n\n"
    dispUsage += "    Example:\n\n"
    dispUsage += "      mqfts --lf=/var/mqm/mqft/logs/QM/agents/SRC/logs/capture0.log\n\n"
    dispUsage += "         OR \n\n"
    dispUsage += "      export MFT_CAPTURE_LOG_PATH=/var/mqm/mqft/logs/QM/agents/SRC/logs/capture0.log\n\n"
    dispUsage += "         OR \n\n"
    dispUsage += "      export MFT_AGENT_NAME=SRC\n"
    dispUsage += "      export MFT_COORDINATION_QM=QM1\n"
    dispUsage += "      export BFG_DATA=/var/mqm\n\n"
    dispUsage += "  Examples:\n"
    dispUsage += "  Display list of transfers a capture log file\n"
    dispUsage += "    mqfts --lf=/var/mqm/mqft/capture0.log\n\n"
    dispUsage += "  Display details of a transfer from a capture log file\n"
    dispUsage += "    mqfts --lf=/var/mqm/mqft/capture0.log --id=414d51204d46544841514d20202020205947c35e2105470f\n"
    fmt.Println(dispUsage)
}
func displayHelp() {
    dispUsage := "\nOptions:\n"
    dispUsage += "\t mqfts       Display status of transfers present in log file\n"
    dispUsage += "\t mqfts <--lf>=<absolute path of capture log file>\n"
    dispUsage += "\t mqfts <--id>=<Transfer ID> Display details of single transfer. Specify * for all transfers\n"
    dispUsage += "\t mqfts <--fl>=<n> Display recent <n> failed transfers\n"
    dispUsage += "\t mqfts <--sf>=<n> Display recent <n> successful transfers\n"
    dispUsage += "\t mqfts <--ps>=<n> Display recent <n> partially successful transfers\n"
    dispUsage += "\t mqfts <--st>=<n> Display recent <n> transfers in 'started' state\n"
    dispUsage += "\t mqfts <--ip>=<n> Display recent <n> 'In Progress' transfers\n"
    fmt.Println(dispUsage)
    displayHelpSample()
}

// Get absolute captureX log file path
func getLogPath(logFilePath string) string {
    var outputLogFilePath string

    logCapturePath, logCapturePathSet := os.LookupEnv("MFT_CAPTURE_LOG_PATH")
    if logCapturePathSet {
        outputLogFilePath = logCapturePath
    } else if isFlagPassed ("lf") {
        outputLogFilePath = logFilePath
    } else {
        var bfgDataPath string
        var agentConfig string
        var agentNameEnv string
        var e error
        var coordinationQMgr string

        logCapturePath, logCapturePathSet := os.LookupEnv("MFT_CAPTURE_LOG_PATH")
        if logCapturePathSet {
            outputLogFilePath = logCapturePath
        } else {
            bfgConfigFilePath, bfgConfigFilePathSet := os.LookupEnv("MFT_AGENT_CONFIG_FILE")
            if bfgConfigFilePathSet {
                // Read agent configuration data from JSON file.
                agentConfig, e = utils.ReadConfigurationDataFromFile(bfgConfigFilePath)
                // Exit if we had any error when reading configuration file
                if e != nil {
                    fmt.Print(e)
                    os.Exit(1)
                }
                coordinationQMgr = gjson.Get(agentConfig, "coordinationQMgr.name").String()
            } else {
                coordinationQMgrLocal, coordinationQMgrSet := os.LookupEnv("MFT_COORDINATION_QM")
                if !coordinationQMgrSet {
                    fmt.Println("Failed to determine coordination queue manager name. Set 'MFT_COORDINATION_QM' environment variable with coordination queue manager")
                    displayHelpSample()
                    os.Exit(1)
                } else {
                    coordinationQMgr = coordinationQMgrLocal
                }
            }

            // Read agent name from environment variable
            agentNameEnvLocal, agentNameEnvSet := os.LookupEnv("MFT_AGENT_NAME")
            if !agentNameEnvSet {
                fmt.Println("Failed to determine agent name from environment. Ensure 'MFT_AGENT_NAME' environment variable set to agent name")
                displayHelpSample()
                os.Exit(1)
            } else {
                agentNameEnv = agentNameEnvLocal
            }

            // Get path from environment variable
            bfgConfigMountPath, bfgConfigMountPathSet := os.LookupEnv("BFG_DATA")
            if !bfgConfigMountPathSet {
                fmt.Println("Failed to determine agent configuration directory name from environment. Ensure BFG_DATA environment variable set to IBM MQ Managed File Transfer data directory")
                displayHelpSample()
                os.Exit(1)
            } else {
                if len(bfgConfigMountPath) > 0 {
                    bfgDataPath = bfgConfigMountPath
                }
            }

            // Build capture log file path
            outputLogFilePath = bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentNameEnv + "/logs/capture0.log"
        }
    }

    return outputLogFilePath
}

/**
 * Parse the transfer XML and display details
 * @param captureLogFileName - Capture log filename
 */
func displayTransferStatus(logFile string){
    _, err := os.Stat(logFile)
    if err != nil {
        if strings.Contains(err.Error(), "no such file or directory") {
            fmt.Println("No transfer logs available")
        } else {
            fmt.Println(err)
        }
        return
    }

    // Read the capture0.log file from the agent logs directory
    outputLogFile, err := os.Open(logFile)
    if err != nil {
        fmt.Print(err)
        return
    }

    // defer the closing of our jsonFile so that we can parse it later on
    defer func(outputLogFile *os.File) {
        err := outputLogFile.Close()
        if err != nil {
            fmt.Printf("An error occurred while closing file %s. The error is: %v\n",outputLogFile.Name(),  err)
        }
    }(outputLogFile)
    fi, err := outputLogFile.Stat()
    if err != nil {
        fmt.Printf("Error when finding capture0.log file. %v\n", err)
        return
    }

    scanner := backscanner.New(outputLogFile, int(fi.Size()))
    topicSystemFTELog := [] byte ("SYSTEM.FTE/Log/")
    // Variable to print header only once
    firstTime := true

    for {
        line, _, err := scanner.LineBytes()
        if err != nil {
            if err == io.EOF {
                    fmt.Printf("\nEnd of file reached.\n")
            } else {
                fmt.Println("Error:", err)
            }
            break
        }

        // Consider only those lines in the capture0.log file that contain SYSTEM.FTE/Log/ string for parsing
        if bytes.Contains(line, topicSystemFTELog) {
            // Split the line on '!' and then parse to get the latest status
            tokens := strings.Split(string(line), "!")
            if len(tokens) > 1 {
                if firstTime {
                    fmt.Println(" Transfer ID                                     \tStatus")
                    fmt.Println("-------------------------------------------------\t------------------")
                    firstTime = false
                }
                parseAndDisplay(tokens[len(tokens) - 1], displayTransferType)
                if displayCount > 0 {
                    if counter == displayCount {
                        // Displayed required number of records. Exit
                        break
                    }
                }
            } // Number of tokens more
        } // If line contains SYSTEM.FTE/Log
    } // For loop
}

// Retrieves the transfer id from XML
func getTransferId(xmlMessage string) string {
    var transferId string

    // Create an parsed XML document
    doc, err := xmlquery.Parse(strings.NewReader(xmlMessage))
    if err != nil {
        panic(err)
    }

    // Get required elements
    transaction := xmlquery.FindOne(doc, "//transaction")
    if transaction != nil {
        transferId = transaction.SelectAttr("ID")
    }
    return transferId
}

/**
 * Parse the transfer XML and display details of the given transfer
 * @param captureLogFileName - Capture log filename
 * @param transferId - ID of the transfer whose details to be displayed
 */
func parseAndDisplayTransfer(captureLogFileName string, transferId string) {
    file, err := os.Open(captureLogFileName)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer func(file *os.File) {
        err := file.Close()
        if err != nil {
            fmt.Println(err)
            return
        }
    }(file)

    topicSystemFTELog := "SYSTEM.FTE/Log/"
    scanner := bufio.NewScanner(file)
    // optionally, resize scanner's capacity for lines over 64K, see next example
    for scanner.Scan() {
        // Consider only those lines in the capture0.log file that contain SYSTEM.FTE/Log/ string for parsing
        if strings.Contains(scanner.Text(), topicSystemFTELog) {
            // Split the line on '!' and then parse to get the latest status
            tokens := strings.Split(scanner.Text(), "!")
            if len(tokens) > 1 {
                transferIdXml := getTransferId(tokens[len(tokens)-1])
                if strings.EqualFold(transferId, "*") {
                    // Display details of all transfers
                    displayTransferDetails(tokens[len(tokens)-1])
                } else {
                    if transferId != "" {
                        if strings.EqualFold(transferIdXml, transferId) {
                            // Display details of specific transfer
                            displayTransferDetails(tokens[len(tokens)-1])
                        }
                    }
                }
            }
        }
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }
}

/**
 * Parse the transfer XML and display details of the given transfer
 * @param xmlMessage - transfer xml
 */
func displayTransferDetails(xmlMessage string){
    // Replace all &quot; with single quote
    strings.ReplaceAll(xmlMessage, "&quot;", "'")
    // Create an parsed XML document
    doc, err := xmlquery.Parse(strings.NewReader(xmlMessage))
    if err != nil {
        panic(err)
    }

    // Get required 'transaction' element from Xml message
    transaction := xmlquery.FindOne(doc, "//transaction")
    if transaction != nil {
        transferId := transaction.SelectAttr("ID")
        if action := transaction.SelectElement("action"); action != nil {
            if strings.EqualFold(action.InnerText(),"completed") {
                // Process transfer complete Xml message
                var supplementMsg string
                status := transaction.SelectElement("status")
                if status != nil {
                    supplementMsg = status.SelectElement("supplement").InnerText()
                    fmt.Printf("\n[%s] TransferID: %s Status: %s\n \tSupplement: %s\n",
                        action.SelectAttr("time"),
                        strings.ToUpper(transferId),
                        action.InnerText(),
                        supplementMsg)
                }

                destAgent := transaction.SelectElement("destinationAgent")
                statistics := transaction.SelectElement("statistics")
                // Retrieve statistics
                var actualStartTimeText = ""
                var retryCount string
                var numFileFailures string
                var numFileWarnings string
                if statistics != nil {
                    actualStartTime := statistics.SelectElement("actualStartTime")
                    if actualStartTime != nil {
                        actualStartTimeText = actualStartTime.InnerText()
                    }
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
                var elapsedTime time.Duration
                if actualStartTimeText != "" {
                    startTime := getFormattedTime(actualStartTimeText)
                    completePublishTIme := getFormattedTime(action.SelectAttr("time"))
                    elapsedTime = completePublishTIme.Sub(startTime)
                }

                fmt.Printf("\tDestination Agent: %s\n\tStart time: %s\n\tCompletion Time: %s\n\tElapsed time: %s\n\tRetry Count: %s\n\tFailures:%s\n\tWarnings:%s\n\n",
                            destAgent.SelectAttr("agent"),
                            actualStartTimeText,
                            action.SelectAttr("time"),
                            elapsedTime,
                            retryCount,
                            numFileFailures,
                            numFileWarnings)
                } else if strings.EqualFold(action.InnerText(),"progress") {
                    // Process transfer progress Xml message
                    destAgent := transaction.SelectElement("destinationAgent")
                    progressPublishTimeText := action.SelectAttr("time")
                    fmt.Printf("\n[%s] %s Status: %s Destination: %s \n", progressPublishTimeText,
                        strings.ToUpper(transferId),
                        action.InnerText(),
                        destAgent.SelectAttr("agent"))
                    transferSet := transaction.SelectElement("transferSet")
                    startTimeText := transferSet.SelectAttr("startTime")
                    //startTime := getFormattedTime(startTimeText)
                    //progressPublishTime := getFormattedTime(progressPublishTimeText)
                    //elapsedTime := progressPublishTime.Sub(startTime)
                    fmt.Printf("\tStart time: %s\n\tTotal items in transfer request: %s\n\tBytes sent: %s\n",
                        startTimeText,
                        transferSet.SelectAttr("total"),
                        transferSet.SelectAttr("bytesSent"))

                    // Loop through all items in the progress message and display details.
                    items := transferSet.SelectElements("item")
                    for i := 0 ; i < len(items); i++ {
                        status := items[i].SelectElement("status")
                        resultCode := status.SelectAttr("resultCode")
                        var sourceName string
                        var sourceSize = "-1"
                        queueSource := items[i].SelectElement("source/queue")
                        if queueSource != nil {
                            sourceName = queueSource.InnerText()
                        } else {
                            fileName := items[i].SelectElement("source/file")
                            if fileName != nil {
                                sourceName = fileName.InnerText()
                                sourceSize = fileName.SelectAttr("size")
                            }
                        }

                        var destinationName string
                        queueDest := items[i].SelectElement("destination/queue")
                        var destinationSize = "-1"
                        if queueDest != nil {
                            destinationName = queueDest.InnerText()
                        } else {
                            fileName := items[i].SelectElement("destination/file")
                            if fileName != nil {
                                destinationName = fileName.InnerText()
                                destinationSize = fileName.SelectAttr("size")
                            }
                        }

                        // Display details of each item
                        fmt.Printf("\tItem # %d\n\t\tSource: %s\tSize: %s bytes\n\t\tDestination: %s\tSize: %s bytes\n",
                            i+1,
                            sourceName, sourceSize,
                            destinationName, destinationSize)
                        // Process result code and append any supplement
                        if resultCode != "0" {
                            supplement := status.SelectElement("supplement")
                            if supplement != nil {
                                fmt.Printf("\t\tResult code %s Supplement %s\n", resultCode, supplement.InnerText())
                            } else {
                                fmt.Printf("\t\tResult code %s\n", resultCode)
                            }
                        } else {
                            fmt.Printf("\t\tResult code %s\n", resultCode)
                        }
                    }
                } else if strings.EqualFold(action.InnerText(),"started") {
                    // Process transfer started Xml message
                    destAgent := transaction.SelectElement("destinationAgent")
                    destinationAgentName := destAgent.SelectAttr("agent")
                    transferSet := transaction.SelectElement("transferSet")
                    startTime := ""
                    if transferSet != nil {
                        startTime = transferSet.SelectAttr("startTime")
                    } else {
                        startTime = action.SelectAttr("time")
                    }
                    fmt.Printf("[%s] TransferID: %s Status: %s Destination: %s\n",
                        startTime,
                        strings.ToUpper(transferId),
                        action.InnerText(),
                        destinationAgentName)
                }
        }
    }
}

func getFormattedTime(timeValue string) time.Time {
    t, err := time.Parse(time.RFC3339, timeValue)
    if err != nil {
        fmt.Printf("%s\n", err)
    }
    return t
}

/*
 * Parse the Transfer XML and display status of transfer in a list format.
 * @param xmlMessage - Transfer XML message.
 * @param displayTransferType - Type of transfers to display like failed only
                                or partial transfer etc
 */
func parseAndDisplay(xmlMessage string, displayTransferType int){
    // Create an parsed XML document
    doc, err := xmlquery.Parse(strings.NewReader(xmlMessage))
    if err != nil {
        panic(err)
    }

    // Get required elements
    transaction := xmlquery.FindOne(doc, "//transaction")
    if transaction != nil {
        transferId := transaction.SelectAttr("ID")
        if !strings.EqualFold(transferId, "") {
            // Check if we have seen this transfer id already. If yes, then don't do anything
            _, exists := transferIdMap[transferId]
            // Does not exist in HashMap
            if !exists {
                if action := transaction.SelectElement("action"); action != nil {
                    if strings.EqualFold(action.InnerText(),"completed") {
                        status := transaction.SelectElement("status")
                        if status != nil {
                            supplement := status.SelectElement("supplement").InnerText()
                            if strings.Contains(supplement, "BFGRP0032I") {
                                if displayTransferType == transferSUCCESSFUL || displayTransferType == 0 {
                                    fmt.Printf("%s\t%s\n", transaction.SelectAttr("ID"), "Successful" )
                                    counter++
                                }
                            } else if strings.Contains(supplement, "BFGRP0034I"){
                                if displayTransferType == transferFAILED  || displayTransferType == 0 {
                                    fmt.Printf("%s\t%s\n", transaction.SelectAttr("ID"), "Failed" )
                                    counter++
                                }
                            } else if strings.Contains(supplement, "BFGRP0033I") {
                                if displayTransferType == transferPARTIALSUCCESS  || displayTransferType == 0 {
                                    fmt.Printf("%s\t%s\n", transaction.SelectAttr("ID"), "Partially successful" )
                                    counter++
                                }
                            } else if strings.Contains(supplement, "BFGRP0036I") {
                                if displayTransferType == transferFAILED  || displayTransferType == 0 {
                                    fmt.Printf("%s\t%s\n", transaction.SelectAttr("ID"), "Completed but no files transferred" )
                                    counter++
                                }
                            } else if strings.Contains(supplement, "BFGRP0037I") {
                                if displayTransferType == transferFAILED  || displayTransferType == 0  {
                                    fmt.Printf("%s\t%s\n", transaction.SelectAttr("ID"), "Failed" )
                                    counter++
                                }
                            }
                        }
                    } else if strings.EqualFold(action.InnerText(),"progress") {
                        if displayTransferType == transferINPROGRESS  || displayTransferType == 0 {
                            fmt.Printf("%s\t%s\n", transaction.SelectAttr("ID"), "In progress")
                            counter++
                        }
                    } else if strings.EqualFold(action.InnerText(),"started") {
                        if displayTransferType == transferSTARTED  || displayTransferType == 0 {
                            fmt.Printf("%s\t%s\n", transaction.SelectAttr("ID"), action.InnerText())
                            counter++
                        }
                    }

                    // Add the transfer id to the map
                    transferIdMap[transaction.SelectAttr("ID")] = action.InnerText()
                }
            }
        }
    }
}
