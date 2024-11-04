/*
© Copyright IBM Corporation 2020, 2022

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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
)

var supportedCommands []string = []string{"fteCancelTransfer",
	"fteClearMonitorHistory",
	"fteCreateMonitor",
	"fteCreateTemplate",
	"fteCreateTransfer",
	"fteDeleteMonitor",
	"fteDeleteScheduledTransfer",
	"fteDeleteTemplates",
	"fteDisplayVersion",
	"fteListAgents",
	"fteListMonitors",
	"fteListScheduledTransfers",
	"fteListTemplates",
	"ftePingAgent",
	"fteSetAgentLogLevel",
	"fteSetAgentTraceLevel",
	"fteShowAgentDetails",
	"fteStartMonitor",
	"fteStopMonitor"}

// Process additional MFT commands that have been specified through
// file with "mftc" extension mounted at /etc/mqft/config directory
// Each line in the mftc file must be a valid MFT command with all
// applicable parameters specified
func postInit() {
	cmdsDir := "/etc/mqft/config"
	fileList, err := os.ReadDir(cmdsDir)
	if err == nil && len(fileList) > 0 {
		for _, fileInfo := range fileList {
			if !fileInfo.IsDir() && fileInfo.Type().IsRegular() {
				if strings.Contains(fileInfo.Name(), ".mftc") {
					processCommand(filepath.Join(cmdsDir, fileInfo.Name()))
				}
			}
		}
	}
}

// Read one line at a time from the specified file and process
// if it is a valid MFT command
func processCommand(cmdFilePath string) {
	cmdFile, err := os.OpenFile(cmdFilePath, os.O_RDONLY, 0)
	if err != nil {
		utils.PrintLog(fmt.Sprintf("Error occurred while opening file %s. The error is %v", cmdFilePath, err))
		return
	}
	defer cmdFile.Close()

	utils.PrintLog(fmt.Sprintf("Processing commands from file %s", cmdFilePath))

	// Iterate through the lines and attempt to execute the valid MFT commands found
	scanner := bufio.NewScanner(cmdFile)
	for scanner.Scan() {
		cmdLine := strings.TrimSpace(scanner.Text())
		// Line is not blank and it begins with valid MFT command.
		if len(cmdLine) > 0 && isValidCommand(cmdLine) {
			// Split and parameterize the line
			parameters := strings.Split(cmdLine, " ")
			if len(parameters) > 0 {
				cmdPath, lookPathErr := exec.LookPath(parameters[0])
				if lookPathErr == nil {
					cmdParams := make([]string, 0)
					// Ignore the first token as it will be the command name itself.
					for index := 0; index < len(parameters); index++ {
						parameter := strings.TrimSpace(parameters[index])
						if len(parameter) > 0 {
							cmdParams = append(cmdParams, parameter)
						}
					}

					cmdExec := &exec.Cmd{
						Path: cmdPath,
						Args: cmdParams,
					}

					var outb, errb bytes.Buffer
					cmdExec.Stdout = &outb
					cmdExec.Stderr = &errb
					// Change current working directory that is writable because some commands create files
					chgErr := os.Chdir("/tmp")
					if chgErr != nil {
						utils.PrintLog("Failed to change current working directory to /tmp")
					}

					if err := cmdExec.Run(); err != nil {
						utils.PrintLog(fmt.Sprintf("Command execution error: %v %v\n", cmdExec, err))
						utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
					} else {
						utils.PrintLog(fmt.Sprintf("Command output: \n%v\n%v", outb.String(), errb.String()))
					}
				} else {
					utils.PrintLog(fmt.Sprintf("Command not found. %v", lookPathErr))
				}
			}
		} else {
			utils.PrintLog(fmt.Sprintf("%s is not a valid IBM MQ Managed File Transfer command", cmdLine))
		}
	}
}

// Does the line begins with a valid MFT command
func isValidCommand(cmdLine string) bool {
	for _, supportedCmd := range supportedCommands {
		if strings.HasPrefix(cmdLine, strings.TrimSpace(supportedCmd)) {
			return true
		}
	}
	return false
}
