/*
© Copyright IBM Corporation 2020, 2026

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

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
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/icza/backscanner"
)

// Default mountpath where MFT Configuration will be created.
const FIXED_BFG_DATAPATH = "/mnt/mftdata"

// Default agent configuration json file
const MFT_DEFAULT_CONFIG_JSON = "/run/mqmft/config.json"

// MFT_PROBE_LOG_DIR holds hourly probe-log segment files named mft-probe-YYYYMMDD-HH.log.
// /tmp is always writable regardless of OpenShift SCC.
const MFT_PROBE_LOG_DIR = "/tmp/mft-probe-logs"

func ReadConfigurationDataFromFile(configFile string) (string, error) {
	var configData string
	jsonFile, err := os.Open(configFile)
	if err != nil {
		return "", err
	}
	defer jsonFile.Close()

	var data []byte
	data, err = io.ReadAll(jsonFile)
	if err != nil {
		return "", err
	}
	configData = string(data)
	return configData, err
}

// IsAgentRunning checks whether the process with the given PID is still alive.
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

// IsAgentReady scans output0.log backwards for the BFGAG0059I ready message.
// Returns false if the "End Display Current Environment" marker is found first,
// or if BFGAG0059I is not present within the last 500 lines.
func IsAgentReady(bfgDataPath string, agentName string, coordinationQMgr string) (bool, error) {
	var ready bool = false
	var count int
	var returnError error
	errorMessages := list.New()

	outputLogFilePath := bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentName + "/logs/output0.log"
	outputLogFile, err := os.Open(outputLogFilePath)

	if err == nil {
		defer outputLogFile.Close()
		fi, err := outputLogFile.Stat()
		if err != nil {
			returnError = fmt.Errorf("error finding output0.log file. Error is: %v", err)
			ready = false
		} else {
			scanner := backscanner.New(outputLogFile, int(fi.Size()))
			findBFGAG0059I := []byte("BFGAG0059I")
			findEndDisplayEnv := []byte("End Display Current Environment")
			for count < 500 {
				line, _, err := scanner.LineBytes()
				if err != nil {
					if err == io.EOF {
						returnError = fmt.Errorf("%q is not found in file", findBFGAG0059I)
					} else {
						returnError = fmt.Errorf("error occurred while processing log file %v", err)
					}
					break
				}

				if bytes.Contains(line, findBFGAG0059I) {
					ready = true
					break
				}

				if bytes.Contains(line, findEndDisplayEnv) {
					returnError = fmt.Errorf("End of the Current Environment detected but ready message is not found in file.")
					break
				}

				errorMessages.PushFront(string(line))
				count++
			}
			if !ready {
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

// GetAgentPid returns the process ID of the agent JVM from its .pid file.
func GetAgentPid(pidFileName string) (int32, error) {
	var agentPid int32
	var returnError error

	_, err := os.Stat(pidFileName)
	if err != nil {
		return -1, err
	}

	pidFile, err := os.Open(pidFileName)
	if err == nil {
		defer pidFile.Close()
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

// CreatePath creates the directory at dataPath (and any parents) if it does not already exist.
func CreatePath(dataPath string) error {
	_, err := os.Stat(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dataPath, 0700)
			if err != nil {
				return fmt.Errorf("failed to create path %s due to error: %v", dataPath, err)
			}
		} else {
			return fmt.Errorf("an error occurred while checking e %v", err)
		}
	}

	return nil
}

// CopyFile copies a file from srcPath to dstPath.
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

// DeleteDir removes the directory at dirPath and all its contents.
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

func IsNumeric(s string) (bool, error) {
	_, err := strconv.ParseInt(s, 10, 64)
	return err == nil, err
}

// ToNumber parses s as a base-10 integer. Returns -1 on failure.
func ToNumber(s string) (int64, error) {
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1, err
	}
	return num, nil
}

func PrintLog(logToPrint string) {
	format := "2006/01/02 15:04:05.000"
	now := time.Now()
	zone, _ := now.Local().Zone()
	loc, _ := time.LoadLocation(zone)
	fmt.Printf("[%s %s] %s\n", now.In(loc).Format(format), zone, logToPrint)
}

func DoesFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return false
		}
	}
	return true
}

// WriteData writes bufferToWrite to fileName, creating or truncating the file.
func WriteData(fileName string, bufferToWrite string) error {
	filePointer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer filePointer.Close()

	_, writeErr := filePointer.WriteString(bufferToWrite)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

// MFT_PROBE_LOG_RETAIN_HOURS controls how many completed hours of probe logs to
// retain in addition to the current hour. Valid range: 1–23. Default: 2.
const MFT_PROBE_LOG_RETAIN_HOURS = "MFT_PROBE_LOG_RETAIN_HOURS"

// probeRetainHours reads MFT_PROBE_LOG_RETAIN_HOURS and returns the configured
// retain count. Falls back to 2 when unset, blank, non-numeric, or out of range.
func probeRetainHours() int {
	const defaultRetain = 2
	const minRetain = 1
	const maxRetain = 23
	val, set := os.LookupEnv(MFT_PROBE_LOG_RETAIN_HOURS)
	if !set || strings.TrimSpace(val) == "" {
		return defaultRetain
	}
	n, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil || n < minRetain || n > maxRetain {
		return defaultRetain
	}
	return n
}

// probeSegmentPaths returns the ordered slice of segment file paths for the
// given retain count. Index 0 is the active file (mft-probe-00.log);
// higher indices are progressively older.
func probeSegmentPaths(retain int) []string {
	paths := make([]string, retain+1)
	for i := range paths {
		paths[i] = filepath.Join(MFT_PROBE_LOG_DIR, fmt.Sprintf("mft-probe-%02d.log", i))
	}
	return paths
}

// rotateProbeSegments shifts all segment files one slot older and removes the
// slot that falls outside the retain window. After this call mft-probe-00.log
// does not exist and is ready to be created fresh by the next write.
// Errors are silently ignored — a probe must never fail because of logging.
func rotateProbeSegments(paths []string) {
	last := len(paths) - 1
	_ = os.Remove(paths[last])
	for i := last; i > 0; i-- {
		_ = os.Rename(paths[i-1], paths[i])
	}
}

// needsRotation reports whether the current UTC hour differs from the UTC hour
// recorded in the modification time of mft-probe-00.log.
func needsRotation(paths []string, now time.Time) bool {
	fi, err := os.Stat(paths[0])
	if err != nil {
		return false // file absent — no rotation needed
	}
	return fi.ModTime().UTC().Truncate(time.Hour).Before(now.UTC().Truncate(time.Hour))
}

// WriteProbeLog appends a single timestamped line to mft-probe-00.log under
// MFT_PROBE_LOG_DIR. Files are rotated at the start of each new UTC hour:
//
//	mft-probe-00.log (latest) → mft-probe-01.log → … → deleted
//
// The number of retained files is controlled by MFT_PROBE_LOG_RETAIN_HOURS
// (default 2, range 1–23).
// Errors are silently ignored — a probe must never fail because of logging.
func WriteProbeLog(probeType string, result string, detail string) {
	const timestampFormat = "2006/01/02 15:04:05.000"

	now := time.Now()
	zone, _ := now.Zone()

	var sb strings.Builder
	sb.WriteByte('[')
	sb.WriteString(now.Format(timestampFormat))
	sb.WriteByte(' ')
	sb.WriteString(zone)
	sb.WriteString("] [")
	sb.WriteString(probeType)
	sb.WriteString("] ")
	sb.WriteString(result)
	if detail != "" {
		sb.WriteString(": ")
		sb.WriteString(detail)
	}
	sb.WriteByte('\n')
	line := sb.String()

	if err := os.MkdirAll(MFT_PROBE_LOG_DIR, 0755); err != nil {
		return
	}

	paths := probeSegmentPaths(probeRetainHours())

	if needsRotation(paths, now) {
		rotateProbeSegments(paths)
	}

	// O_APPEND is atomic for writes ≤ PIPE_BUF on Linux so no locking needed.
	f, err := os.OpenFile(paths[0], os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	f.WriteString(line) //nolint:errcheck
	f.Close()
}
