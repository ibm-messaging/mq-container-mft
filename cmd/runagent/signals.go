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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"

	"golang.org/x/sys/unix"
)

const (
	startReaping = iota
	reapNow      = iota
)

func signalHandler(agentName string, coordinationQMgr string) chan int {
	control := make(chan int)
	// Use separate channels for the signals, to avoid SIGCHLD signals swamping
	// the buffer, and preventing other signals.
	stopSignals := make(chan os.Signal)
	reapSignals := make(chan os.Signal)
	signal.Notify(stopSignals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			select {
			case sig := <-stopSignals:
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_SIGNAL_RECD_0071, sig))
				signal.Stop(reapSignals)
				signal.Stop(stopSignals)
				// #nosec G104
				stopAgent(agentName, coordinationQMgr)
				// One final reap
				reapZombies()
				close(control)
				// End the goroutine
				return
			case <-reapSignals:
				if logLevel >= LOG_LEVEL_VERBOSE {
					utils.PrintLog(utils.MFT_CONT_SIGNAL_CHILD_0069)
				}
				reapZombies()
			case job := <-control:
				switch {
				case job == startReaping:
					// Add SIGCHLD to the list of signals we're listening to
					if logLevel >= LOG_LEVEL_VERBOSE {
						utils.PrintLog(utils.MFT_CONT_SIGNAL_LISTEN_0070)
					}
					signal.Notify(reapSignals, syscall.SIGCHLD)
				case job == reapNow:
					reapZombies()
				}
			}
		}
	}()
	return control
}

// reapZombies reaps any zombie (terminated) processes now.
// This function should be called before exiting.
func reapZombies() {
	for {
		var ws unix.WaitStatus
		pid, err := unix.Wait4(-1, &ws, unix.WNOHANG, nil)
		// If err or pid indicate "no child processes"
		if pid == 0 || err == unix.ECHILD {
			return
		}
		if logLevel >= LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_REAPED_PID_0072, pid))
		}
	}
}

// Stops an agent when container stop is issued.
func stopAgent(agentName string, coordinationQMgr string) {
	var outb, errb bytes.Buffer
	// Get the path of MFT fteStopAgent command.
	cmdStopAgntPath, lookPathErr := exec.LookPath("fteStopAgent")
	if lookPathErr != nil {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
		os.Exit(1)
	}
	cmdStopAgnt := &exec.Cmd{
		Path: cmdStopAgntPath,
		Args: []string{cmdStopAgntPath, "-p", coordinationQMgr, agentName, "-i"},
	}

	outb.Reset()
	errb.Reset()
	cmdStopAgnt.Stdout = &outb
	cmdStopAgnt.Stderr = &errb
	err := cmdStopAgnt.Run()
	if err != nil {
		utils.PrintLog(fmt.Sprintf("An error occured when running fteStopAgent command. The error is: %s", err.Error()))
		utils.PrintLog(fmt.Sprintf("Command: %s\n", outb.String()))
		utils.PrintLog(fmt.Sprintf("Error %s\n", errb.String()))
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGENT_STOPPED_0068, agentName))
	}
}

// Temporary logging
func writeLog(messageToLog string) {
	logPath := os.Getenv("BFG_DATA") + "/mqft/logs/signal.log"
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		utils.PrintLog(err.Error())
		return
	}
	defer f.Close()

	// If we can't write to file
	_, errWrite := f.WriteString(messageToLog + "\n")
	if errWrite != nil {
		utils.PrintLog(errWrite.Error())
		return
	}
}
