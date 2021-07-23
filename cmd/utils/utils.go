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

package utils

import (
	"os"
    "io/ioutil"
	"github.com/icza/backscanner"
	"strconv"
	"syscall"
	"io"
	"bytes"
	"time"
	"fmt"
)

const FIXED_BFG_DATAPATH = "/mnt/mftdata"

// Read configuration data from json file
func ReadConfigurationDataFromFile(configFile string) (string, error ) {
	var agentConfig string
	jsonFile, err := os.Open(configFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		PrintLog(fmt.Sprintf("%v", err))
		return agentConfig, err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
  
	// read file
	var data []byte
	data, err = ioutil.ReadAll(jsonFile)
	if err != nil {
		PrintLog(fmt.Sprintf("%v", err))
		return agentConfig, err
	}
	agentConfig = string(data)
  
	return agentConfig, err
}

// Is agent running? 
func IsAgentRunning(agentPid int32) (bool, error) {
	if agentPid <= 0 {
		PrintLog(fmt.Sprintf("Invalid agentPid %d", agentPid))
		return false, nil
	}
	proc, err := os.FindProcess(int(agentPid))
	if err != nil {
		PrintLog(fmt.Sprintf("%v", err))
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "process already ended" {
		PrintLog(fmt.Sprintf("%v", err))
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		PrintLog(fmt.Sprintf("%v", err))
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
}

// Is agent ready - basically check for BFGAG0059I message in agent's log file
func IsAgentReady(bfgDataPath string, agentName string, coordinationQMgr string) (bool) {
	var ready bool = false
	
	// Read the agentPid file from the agent logs directory
	outputLogFilePath := bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentName + "/logs/output0.log"
	outputLogFile, err := os.Open(outputLogFilePath)
	
	// if we os.Open returns an error then handle it
	if err == nil {
		// defer the closing of our jsonFile so that we can parse it later on
		defer outputLogFile.Close()
		fi, err := outputLogFile.Stat()
		if err != nil {
			PrintLog(fmt.Sprintf("Error finding output0.log file. Error is: %v", err))
			ready = false
		} else {
			scanner := backscanner.New(outputLogFile, int(fi.Size()))
			what := []byte("BFGAG0059I")
			for {
				line, _, err := scanner.LineBytes()
				if err != nil {
					if err == io.EOF {
						PrintLog(fmt.Sprintf("%q is not found in file", what))
					} else {
						PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
					}
					break
				}

				if bytes.Contains(line, what) {
					ready = true
					break
				}
			}
		}
	} else {
		PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
	}
	
	return ready
}

// Returns Process ID of agent JVM from .pid file
func GetAgentPid(pidFileName string) (int32) {
	var agentPid int32
  
	pidFile, err := os.Open(pidFileName)
	// if we os.Open returns an error then handle it
	if err == nil {
		// defer the closing of our jsonFile so that we can parse it later on
		defer pidFile.Close()
		// read file
		var data []byte
		data, err = ioutil.ReadAll(pidFile)
		if err == nil {
			agentPidRead, err := strconv.Atoi(string(data))
			if err != nil {
				agentPid = -1
			} else {
				agentPid = int32(agentPidRead)
			}
		} else {
			PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
			agentPid = -1
		}
	} else {
		PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
		agentPid = -1
	}
	
	PrintLog(fmt.Sprintf("Process ID of agent: %d", agentPid))
	return agentPid
}

// Create specified MFT data path if it does not exist
func CreateDataPath(dataPath string ) (error){
	_, err := os.Stat(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			PrintLog(fmt.Sprintf("Datapath %s does not exist. Creating...", dataPath))
			err := os.MkdirAll(dataPath, 0777)
			if err != nil {
				PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
				return err
			} else {
				// Change permissions Linux.
				err = os.Chmod(dataPath, 0777)
				if err != nil {
					PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
					return err
				}
			}
		} else {
			PrintLog(fmt.Sprintf("An error occurred while determining for Managed File Transfer Datapath - BFG_DATA. %v", err))
			return err
		}
	}
 
	return nil
}

// Copy a file from source to destination
func CopyFile(srcPath string, dstPath string) (error) {
    srcFile, err := os.Open(srcPath)
    if err != nil {
		PrintLog(fmt.Sprintf("%v", err))
        return err
    }
    defer srcFile.Close()

    destFile, err := os.Create(dstPath)
    if err != nil {
		PrintLog(fmt.Sprintf("%v", err))
        return err
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, srcFile)
    if err != nil {
		PrintLog(fmt.Sprintf("%v", err))
        return err
    }
	
    return destFile.Close()	
}

// Delete the specified directory
func DeleteDir (dirPath string) {
	err := os.RemoveAll(dirPath)
	if err != nil {
		PrintLog(fmt.Sprintf("Failed to delete directory - %s due to %v. Continuing", dirPath, err))
	}
}

// Is the given string a number
func IsNumeric(s string) bool {
    _, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
	}
    return err == nil
}

// Convert to a number. Return -1 if can't convert
func ToNumber (s string) int64 {
    num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		PrintLog(fmt.Sprintf("Error parsing numeric string %v", err))
		return -1
	}
    return num
}

// Print log statement on console
func PrintLog(logToPrint string) {
	format :="02/01/2006 15:04:05.000"
	now := time.Now()
	zone, _ := now.Zone()
	loc, _ := time.LoadLocation(zone)
	fmt.Printf("[%s %s] %s\n", now.In(loc).Format(format), zone,  logToPrint)
}