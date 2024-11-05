/*
© Copyright IBM Corporation 2020, 2024

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
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"time"

	"github.com/icza/backscanner"
)

// Default mountpath where MFT Configuration will be created.
const FIXED_BFG_DATAPATH = "/mnt/mftdata"

// Default agent configuration json file
const MFT_DEFAULT_CONFIG_JSON = "/run/mqmft/config.json"

// Read configuration data from json file
func ReadConfigurationDataFromFile(configFile string) (string, error) {
	var configData string
	jsonFile, err := os.Open(configFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		return "", err
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read file
	var data []byte
	data, err = io.ReadAll(jsonFile)
	if err != nil {
		return "", err
	}
	// Convert to a string
	configData = string(data)
	return configData, err
}

// Is agent running?
func IsAgentRunning(agentPid int32) (bool, error) {
	if agentPid <= 0 {
		return false, fmt.Errorf("invalid agentPid %d", agentPid)
	}
	proc, err := os.FindProcess(int(agentPid))
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "process already ended" {
		return false, fmt.Errorf("%v", err)
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
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
func IsAgentReady(bfgDataPath string, agentName string, coordinationQMgr string) (bool, error) {
	var ready bool = false
	//var errorMessages []string
	var count int
	var returnError error
	errorMessages := list.New()

	// Read the agentPid file from the agent logs directory
	outputLogFilePath := bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentName + "/logs/output0.log"
	outputLogFile, err := os.Open(outputLogFilePath)

	// if we os.Open returns an error then handle it
	if err == nil {
		// defer the closing of our jsonFile so that we can parse it later on
		defer outputLogFile.Close()
		fi, err := outputLogFile.Stat()
		if err != nil {
			returnError = fmt.Errorf("error finding output0.log file. Error is: %v", err)
			ready = false
		} else {
			scanner := backscanner.New(outputLogFile, int(fi.Size()))
			findBFGAG59I := []byte("BFGAG0059I")
			// If we don't find BFGAG0059I in last 10 lines, then assume agent
			// has not started and return false
			for count < 10 {
				line, _, err := scanner.LineBytes()
				if err != nil {
					if err == io.EOF {
						returnError = fmt.Errorf("%q is not found in file", findBFGAG59I)
					} else {
						returnError = fmt.Errorf("error occurred while processing log file %v", err)
					}
					break
				}

				if bytes.Contains(line, findBFGAG59I) {
					ready = true
					break
				}

				errorMessages.PushFront(string(line))
				count++
			}
			// Agent did not start, return error message
			if !ready {
				// Iterate through list and print its contents.
				errorMsg := ""
				for e := errorMessages.Front(); e != nil; e = e.Next() {
					errorMsg += fmt.Sprintf("%s\n", e.Value)
				}
				returnError = errors.New(errorMsg)
			}
		}
	} else {
		returnError = fmt.Errorf("error occurred while processing log file %v", err)
	}
	return ready, returnError
}

func ListDirectory(dirName string) {
	files, err := os.ReadDir(dirName)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	for _, f := range files {
		fmt.Println(f.Name())
	}
}

// Returns Process ID of agent JVM from .pid file
func GetAgentPid(pidFileName string) (int32, error) {
	var agentPid int32
	var returnError error

	_, err := os.Stat(pidFileName)
	if err != nil {
		return -1, err
	}

	pidFile, err := os.Open(pidFileName)
	// if we os.Open returns an error then handle it
	if err == nil {
		// defer the closing of our jsonFile so that we can parse it later on
		defer pidFile.Close()
		// read file
		var data []byte
		data, err = io.ReadAll(pidFile)
		if err == nil {
			agentPidRead, err := strconv.Atoi(string(data))
			if err != nil {
				agentPid = -1
			} else {
				agentPid = int32(agentPidRead)
			}
		} else {
			returnError = err
			agentPid = -1
		}
	} else {
		returnError = err
		agentPid = -1
	}

	return agentPid, returnError
}

// Create specified path if it does not exist
func CreatePath(dataPath string) error {
	_, err := os.Stat(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dataPath, 0777)
			if err != nil {
				return fmt.Errorf("failed to create path %s due to error: %v", dataPath, err)
			} else {
				// Change permissions Linux.
				err = os.Chmod(dataPath, 0777)
				if err != nil {
					return fmt.Errorf("failed to modify permissions on path %s due to error %v", dataPath, err)
				}
			}
		} else {
			return fmt.Errorf("an error occurred while checking e %v", err)
		}
	}

	return nil
}

// Copy a file from source to destination
func CopyFile(srcPath string, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	return destFile.Close()
}

// Delete the specified directory
func DeleteDir(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("failed to delete directory - %s due to %v. Continuing", dirPath, err)
	}
	return nil
}

func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete directory - %s due to %v. Continuing", filePath, err)
	}
	return nil
}

// Is the given string a number
func IsNumeric(s string) (bool, error) {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil, err
}

// Convert to a number. Return -1 if can't convert
func ToNumber(s string) (int64, error) {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {

		return -1, err
	}
	return num, nil
}

// Print log statement on console
func PrintLog(logToPrint string) {
	format := "02/01/2006 15:04:05.000"
	now := time.Now()
	zone, _ := now.Local().Zone()
	loc, _ := time.LoadLocation(zone)
	fmt.Printf("[%s %s] %s\n", now.In(loc).Format(format), zone, logToPrint)
}

/**
* Check if the specified file exist.
* @param - Name of the file.
* @return - True if file exists else false or some error occurs.
 */
func DoesFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return false
		}
	}
	return true
}

// Write the given buffer to specified file
func WriteData(fileName string, bufferToWrite string) error {
	// Create an empty credentials file, truncate if one exists
	filePointer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}
	defer filePointer.Close()

	// Write the updated properties to file.
	_, writeErr := filePointer.WriteString(bufferToWrite)
	if writeErr != nil {
		return writeErr
	}

	return nil
}
