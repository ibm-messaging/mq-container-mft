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
	"context"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

// Variable that controls the diagnotstic information level
var logLevel int

// Variable to track if command tracing is enabled or not
var commandTracingEnabled bool = false

// JSON configuration file path
var jsonAgentConfigFilePath string

// Agent name
var agentNameGlobal string

/**
* Main entry point of the runagent program.
*
* This program reads the agent configuration details provided in JSON file,
* configures an agent and starts it. Details of configuration steps are
* displayed on the console when container is being started.
 */
func main() {
	var bfgDataPath string
	var allAgentConfig string
	var e error

	// By default minimal logging is enabled.
	logLevel = LOG_LEVEL_INFO
	// Determine the level of diagnostic information to be logged.
	logLevelStr, logLevelSet := os.LookupEnv(MFT_LOG_LEVEL)
	if logLevelSet {
		// Trim the value specified.
		logLevelStr = strings.Trim(logLevelStr, TEXT_TRIM)
		if strings.EqualFold(logLevelStr, LOG_LEVEL_VERBOSE_TXT) {
			// Verbose level logging.
			logLevel = LOG_LEVEL_VERBOSE
			utils.PrintLog(utils.MFT_CONT_DIAGNOSTIC_LEVEL_0002)
		} else {
			// Any other level or unknown value, set the level to Info
			if !strings.EqualFold(logLevelStr, LOG_LEVEL_INFO_TXT) {
				utils.PrintLog(utils.MFT_CONT_DIAGNOSTIC_LEVEL_0073)
			} else {
				utils.PrintLog(utils.MFT_CONT_DIAGNOSTIC_LEVEL_0001)
			}
		}
	} else {
		utils.PrintLog(utils.MFT_CONT_DIAGNOSTIC_LEVEL_0001)
	}

	// Print container image details
	printImageInfo()

	// There should only be one instance of this process.
	singleErr := verifySingleProcess()
	if singleErr != nil {
		utils.PrintLog(singleErr.Error())
		os.Exit(MFT_CONT_ERR_CODE_24)
	}

	// First check if license is accepted or not.
	accepted, err := checkLicense()
	if err != nil {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_LIC_ERROR_OCCUR_0074, err))
		os.Exit(MFT_CONT_ERR_CODE_1)
	}
	// Exit if license == view
	if !accepted {
		os.Exit(MFT_CONT_SUCCESS_CODE_0)
	}

	// Determine we have the agent name specified in environment variable
	agentNameEnv, agentNameSet := os.LookupEnv(MFT_AGENT_NAME)
	if !agentNameSet {
		utils.PrintLog(utils.MFT_CONT_ENV_AGENT_NAME_NOT_SPECIFIED_0006)
		os.Exit(MFT_CONT_ERR_CODE_3)
	}
	agentNameEnv = strings.Trim(agentNameEnv, TEXT_TRIM)
	utils.PrintLog(fmt.Sprintf(utils.MFT_AGENT_NAME_CONFIGURE, agentNameEnv))
	if len(agentNameEnv) == 0 {
		utils.PrintLog(utils.MFT_CONT_ENV_AGENT_NAME_BLANK_0007)
		os.Exit(MFT_CONT_ERR_CODE_4)
	}
	// Copy the name of agent
	agentNameGlobal = agentNameEnv

	// Time to wait for agent to start. Default wait time is 10 seconds
	delayTimeStatusCheck := time.Duration(10) * time.Second
	timeWaitForAgentStartStr, timeWaitForAgentStartSet := os.LookupEnv(MFT_AGENT_START_WAIT_TIME)
	// Value is numeric and above 0, then use it.
	if timeWaitForAgentStartSet {
		isNum, _ := utils.IsNumeric(timeWaitForAgentStartStr)
		if isNum {
			waitTime, _ := utils.ToNumber(timeWaitForAgentStartStr)
			if waitTime > 0 {
				delayTimeStatusCheck = time.Duration(waitTime) * time.Second
			}
		} else {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(utils.MFT_CONT_ENV_AGENT_START_TIME_0008)
			}
		}
	}

	// Check if MFT command tracing is enabled
	commandTracingEnabled = IsCommandTracingEnabled()

	// Create directory for file transfers
	errVol := utils.CreatePath(MOUNT_PATH_TRANSFERS)
	if errVol != nil {
		utils.PrintLog(errVol.Error())
	} else {
		utils.PrintLog(fmt.Sprintf("Transfer root directory '%s' created", MOUNT_PATH_TRANSFERS))
	}

	// See if we have been given mount point for creating agent configuration and log directory.
	bfgConfigMountPath, bfgCfgMountPathSet := os.LookupEnv(BFG_DATA)
	if bfgCfgMountPathSet {
		bfgConfigMountPath = strings.Trim(bfgConfigMountPath, TEXT_TRIM)
		if len(bfgConfigMountPath) > 0 {
			bfgDataPath = bfgConfigMountPath
			// We have a path specified. Attempt to create the directory
			// Ignore errors if directory already exists
			err = utils.CreatePath(bfgDataPath)
			// Exit the creation if an error occurs
			if err != nil {
				utils.PrintLog(fmt.Sprintf("%v", err))
				os.Exit(MFT_CONT_ERR_CODE_5)
			}
		} else {
			// Blank value was specified, hence use default
			utils.PrintLog(utils.MFT_CONT_ENV_BFG_DATA_BLANK_0009)
			bfgDataPath = utils.FIXED_BFG_DATAPATH
			err = utils.CreatePath(bfgDataPath)
			if err != nil {
				os.Exit(MFT_CONT_ERR_CODE_6)
			}

			// Set BFG_DATA environment variable so that we can run MFT commands.
			os.Setenv(BFG_DATA, bfgDataPath)
		}
	} else {
		// Make the default BFG_DATA path as /mnt/mftdata if BFG_DATA is not set.
		bfgDataPath = utils.FIXED_BFG_DATAPATH
		err = utils.CreatePath(bfgDataPath)
		if err != nil {
			os.Exit(MFT_CONT_ERR_CODE_7)
		}

		// Set BFG_DATA environment variable so that we can run MFT commands.
		os.Setenv(BFG_DATA, bfgDataPath)
	}
	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CONFIG_PATH_0010, bfgDataPath))

	// Read agent configuration data from file specified in the environment
	// variable MFT_AGENT_CONFIG_FILE.
	bfgConfigFilePath, configFileSet := os.LookupEnv(MFT_AGENT_CONFIG_FILE)
	if !configFileSet {
		utils.PrintLog(utils.MFT_CONT_ENV_AGNT_CFG_FILE_NOT_SPECIFIED_0011)
		os.Exit(MFT_CONT_ERR_CODE_8)
	}
	bfgConfigFilePath = strings.Trim(bfgConfigFilePath, TEXT_TRIM)
	if bfgConfigFilePath == TEXT_BLANK {
		utils.PrintLog(utils.MFT_CONT_ENV_AGNT_CFG_FILE_BLANK_0012)
		os.Exit(MFT_CONT_ERR_CODE_9)
	}

	// Copy the JSON configuration file path
	jsonAgentConfigFilePath = bfgConfigFilePath
	// Read the entire agent configuration data from JSON file. The configuration file
	// may contain data for multiple agents. We will choose data for matching agent name
	// specified in MFT_AGENT_NAME environment variable
	allAgentConfig, e = utils.ReadConfigurationDataFromFile(bfgConfigFilePath)
	if e != nil {
		// Exit if we had any error when reading configuration file
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_FILE_READ_0013, bfgConfigFilePath, e))
		os.Exit(MFT_CONT_ERR_CODE_10)
	}

	// Validate coordination queue manager attributes. Throw an error if minimum attributes
	// are not available
	errorCrd := validateCoordinationAttributes(allAgentConfig)
	if errorCrd != nil {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_MISSING_ATTRIBS_0016, bfgConfigFilePath, errorCrd))
		os.Exit(MFT_CONT_ERR_CODE_11)
	}

	// Validate command queue manager attributes. Throw an error if minimum attributes are
	// not available
	errorCmd := validateCommandAttributes(allAgentConfig)
	if errorCmd != nil {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_MISSING_ATTRIBS_0016, bfgConfigFilePath, errorCmd))
		os.Exit(MFT_CONT_ERR_CODE_12)
	}
	if logLevel >= LOG_LEVEL_VERBOSE {
		utils.PrintLog(fmt.Sprintf("All configurations in %s file: %v", bfgConfigFilePath, allAgentConfig))
	}

	// We may have multiple agent configurations defined in the JSON file. Iterate through all
	// definitions and pick the one that has matching agent name specified in environment variable
	var singleAgentConfig string
	configurationFound := false
	agentsJson := gjson.Get(allAgentConfig, "agents").Array()
	// Return an error if no agent configuration is supplied
	if len(agentsJson) == 0 {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_NO_AGENT_CONFIG_SUPPLIED, bfgConfigFilePath))
		os.Exit(MFT_CONT_ERR_CODE_23)
	}

	// Loop through the supplied JSON and identify configuration for agent name supplied
	// via environment variable MFT_AGENT_NAME
	for i := 0; i < len(agentsJson); i++ {
		singleAgentConfig = agentsJson[i].String()
		if logLevel >= LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf(utils.MFT_AGENT_JSON_CONFIG, singleAgentConfig))
		}
		if gjson.Get(singleAgentConfig, "name").Exists() {
			agentNameConfig := gjson.Get(singleAgentConfig, "name").String()
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf(utils.MFT_AGENT_NAME_CONFIG_FILE, agentNameConfig))
			}
			agentNameConfig = strings.Trim(agentNameConfig, TEXT_TRIM)
			if strings.EqualFold(agentNameConfig, agentNameEnv) {
				configurationFound = true
				break
			}
		}
	}

	// Exit if we did not find the configuration for specified agent
	if !configurationFound {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_AGENT_CONFIG_MISSING_0019, agentNameEnv, bfgConfigFilePath))
		os.Exit(MFT_CONT_ERR_CODE_13)
	} else {
		err := ValidateAgentAttributes(singleAgentConfig)
		if err != nil {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_AGENT_CONFIG_ERROR_0023, bfgConfigFilePath, err))
			os.Exit(MFT_CONT_ERR_CODE_14)
		}
	}

	// Cache the coordination queue manager name
	coordinationQMgr := gjson.Get(allAgentConfig, "coordinationQMgr.name").String()

	// Setup coordination configuration
	coordinationCreated := setupCoordination(allAgentConfig, bfgDataPath, agentNameEnv)
	if !coordinationCreated {
		utils.PrintLog(utils.MFT_CONT_CORD_CFG_FAILED_0029)
		os.Exit(MFT_CONT_ERR_CODE_15)
	}

	// Setup command configuration
	commandsCreated := setupCommands(allAgentConfig, bfgDataPath, agentNameEnv)
	if !commandsCreated {
		utils.PrintLog(utils.MFT_CONT_CMD_CFG_FAILED_0030)
		os.Exit(MFT_CONT_ERR_CODE_16)
	}

	// Create the specified agent configuration
	setupAgentDone := setupAgent(singleAgentConfig, bfgDataPath, coordinationQMgr)
	if !setupAgentDone {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_CFG_FAILED_0031, agentNameEnv))
		os.Exit(MFT_CONT_ERR_CODE_17)
	}

	// Clean agent if asked for before starting the agent
	cleanAgent(singleAgentConfig, coordinationQMgr, agentNameEnv)

	// Submit request to start the agent.
	startAgentDone := StartAgent(agentNameEnv, coordinationQMgr)
	if !startAgentDone {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_START_FAILED_0032, agentNameEnv))
		os.Exit(MFT_CONT_ERR_CODE_18)
	}

	// Setup agent log mirroring.
	var wg sync.WaitGroup
	defer func() {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_WAIT_MIRROR_CMP_0035, agentNameEnv))
		wg.Wait()
	}()

	ctxAgentLog, cancelMirrorAgentLog := context.WithCancel(context.Background())
	defer func() {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_WAIT_MIRROR_STOP_0036, agentNameEnv))
		cancelMirrorAgentLog()
	}()

	// Display the contents of agent's output0.log file on the console.
	if logLevel >= LOG_LEVEL_VERBOSE {
		agentLogPath := bfgDataPath + DIR_AGENT_LOGS + coordinationQMgr + DIR_AGENTS + agentNameEnv + "/logs/output0.log"
		mirrorAgentLogs(ctxAgentLog, &wg, agentNameEnv, agentLogPath, "", "", LOG_TYPE_CONSOLE, -1)
	}

	// Verify that agent is ready to accept to requests
	pingWaitTime := strconv.Itoa((int)(delayTimeStatusCheck / time.Second))
	agentReady := PingAgent(coordinationQMgr, agentNameEnv, pingWaitTime)
	if !agentReady {
		//if agent not started yet, wait for some time and then reissue fteListAgents commad
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_NOT_STARTED_0033, agentNameEnv, delayTimeStatusCheck/time.Second))
		time.Sleep(delayTimeStatusCheck)
		agentReady = PingAgent(coordinationQMgr, agentNameEnv, pingWaitTime)
		// Agent has not started, exit.
		if !agentReady {
			if logLevel >= LOG_LEVEL_INFO {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_FAILED_TO_START_0034, agentNameEnv))
			}
			os.Exit(MFT_CONT_ERR_CODE_21)
		}
	}

	// Check if the agent is ready, by searching for BFGAG0059 event ID
	// output0.log file
	isReady, readyError := utils.IsAgentReady(bfgDataPath, agentNameEnv, coordinationQMgr)
	if readyError != nil || !isReady {
		if readyError != nil {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_NOT_READY_ERROR, agentNameEnv, readyError))
		}
		// There was an error or agent is not ready, then exit
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_NOT_READY, agentNameEnv))
		os.Exit(MFT_CONT_ERR_CODE_21)
	}

	// Mirror contents of capture log on the console
	setupMirrorCaptureLogs(ctxAgentLog, &wg, bfgDataPath, coordinationQMgr, agentNameEnv)

	// Mirror contents trace files to console
	setupMirrorTraceLogs(ctxAgentLog, &wg, bfgDataPath, coordinationQMgr, agentNameEnv)

	// Push transfer logs to specified server
	setupMirrorTransferLogs(ctxAgentLog, &wg, bfgDataPath, coordinationQMgr, agentNameEnv)

	// If agent status is READY or ACTIVE, then we are good.
	if agentReady {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_STARTED_0038, agentNameEnv))
		// Execute any commands provided in the cmds file
		postInit()

		// Setup a siganl handle and wait for till container is stopped.
		signalControl := signalHandler(agentNameEnv, coordinationQMgr)
		<-signalControl
		cancelMirrorAgentLog()

		// Delete agent configuration on exit
		deleteAgentOnExit := gjson.Get(singleAgentConfig, "deleteOnTermination").Bool()
		// Delete agent configuration if asked for
		if deleteAgentOnExit {
			deleteAgent(coordinationQMgr, agentNameEnv)
		} else {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_CFG_DELETED_0039, agentNameEnv))
		}

		// Agent has ended. Return success
		os.Exit(MFT_CONT_SUCCESS_CODE_0)
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_START_FAILED_0040, agentNameEnv))
		os.Exit(MFT_CONT_ERR_CODE_22)
	}
}

// Mirror trace file contents to console if we have been asked
func setupMirrorCaptureLogs(ctxCaptureLog context.Context, wg *sync.WaitGroup, bfgDataPath string,
	coordinationQMgr string, agentNameEnv string) {
	agentCaptureLogEnv, agentCaptureLogEnvSet := os.LookupEnv(MFT_AGENT_DISPLAY_CAPTURE_LOG)
	if agentCaptureLogEnvSet {
		if strings.EqualFold(agentCaptureLogEnv, TEXT_YES) {
			captureLogPath := bfgDataPath + DIR_AGENT_LOGS + coordinationQMgr + DIR_AGENTS + agentNameEnv + "/logs/capture0.log"
			mirrorAgentLogs(ctxCaptureLog, wg, agentNameEnv, captureLogPath, "", "", LOG_TYPE_CONSOLE, -1)
		} else {
			if !strings.EqualFold(agentCaptureLogEnv, TEXT_NO) {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_CAPT_LOG_ERROR_0037, agentCaptureLogEnv))
			}
		}
	}
}

// Mirror trace file contents to console if we have been asked
func setupMirrorTraceLogs(ctxCaptureLog context.Context, wg *sync.WaitGroup, bfgDataPath string,
	coordinationQMgr string, agentNameEnv string) {
	agentTraceEnv, agentTraceEnvSet := os.LookupEnv(MFT_AGENT_ENABLE_TRACE)
	if agentTraceEnvSet {
		if strings.EqualFold(agentTraceEnv, TEXT_YES) {
			agentPidPath := bfgDataPath + DIR_AGENT_LOGS + coordinationQMgr + DIR_AGENTS + agentNameEnv + "/agent.pid"
			agentPid, _ := utils.GetAgentPid(agentPidPath)
			agentTracePath := bfgDataPath + DIR_AGENT_LOGS + coordinationQMgr + DIR_AGENTS + agentNameEnv + "/logs/trace" + strconv.Itoa(int(agentPid)) + "/trace" + strconv.Itoa(int(agentPid)) + ".txt.0"
			mirrorAgentLogs(ctxCaptureLog, wg, agentNameEnv, agentTracePath, "", "", LOG_TYPE_CONSOLE, -1)
		}
	}
}

// Setup a mirror to push transfer logs to specified server
func setupMirrorTransferLogs(ctxTransferLog context.Context, wg *sync.WaitGroup, bfgDataPath string,
	coordinationQMgr string, agentNameEnv string) {
	agentTransferLogEnv, agentTransferLogEnvSet := os.LookupEnv(MFT_AGENT_TRANSFER_LOG_PUBLISH_CONFIG_FILE)
	if agentTransferLogEnvSet {
		if !strings.EqualFold(strings.Trim(agentTransferLogEnv, TEXT_TRIM), TEXT_BLANK) {
			// Read the URL and Injestion key from the given JSON file.
			serverLogData, e := utils.ReadConfigurationDataFromFile(agentTransferLogEnv)
			if e != nil {
				// Exit if we had any error when reading configuration file
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_FILE_READ_0013, agentTransferLogEnv, e))
			} else {
				// Check if the data can be bas64 decoded - as the data may have come from
				// kubernetes secret
				serverLogDataDecoded, decoded := Base64Decode(serverLogData)
				if decoded != nil {
					// decode successful
					serverLogData = serverLogDataDecoded
				} else {
					// Failed to decode, use it as it is.
				}

				// Is it a valid json
				if !gjson.Valid(serverLogData) {
					// Not valid. log a message to console and exit
					utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_FILE_READ_0013, agentTransferLogEnv, e))
					return
				}
				if gjson.Get(serverLogData, KEY_TYPE).Exists() {
					logType := gjson.Get(serverLogData, KEY_TYPE).String()
					if strings.EqualFold(strings.Trim(logType, TEXT_BLANK), LOG_SERVER_TYPE_DNA) {
						if gjson.Get(serverLogData, KEY_URL_DNA).Exists() &&
							gjson.Get(serverLogData, KEY_INJESTION_DNA).Exists() {
							logDNAUrl := gjson.Get(serverLogData, KEY_URL_DNA).String()
							logDNAKey := gjson.Get(serverLogData, KEY_INJESTION_DNA).String()
							transferLogPath := bfgDataPath + DIR_AGENT_LOGS + coordinationQMgr + DIR_AGENTS + agentNameEnv + "/logs/transferlog0.json"
							mirrorAgentLogs(ctxTransferLog, wg, "IBMMQMFT Agent "+agentNameEnv, transferLogPath, logDNAUrl, logDNAKey, LOG_TYPE_TRANSFER,
								LOG_SERVER_TYPE_DNA_NUM)
						}
					} else if strings.EqualFold(strings.Trim(logType, TEXT_BLANK), LOG_SERVER_TYPE_ELK) {
						if gjson.Get(serverLogData, KEY_URL_ELK).Exists() {
							logUrlElk := gjson.Get(serverLogData, KEY_URL_ELK).String()
							transferLogPath := bfgDataPath + DIR_AGENT_LOGS + coordinationQMgr + DIR_AGENTS + agentNameEnv + "/logs/transferlog0.json"
							mirrorAgentLogs(ctxTransferLog, wg, agentNameEnv, transferLogPath, logUrlElk, "", LOG_TYPE_TRANSFER,
								LOG_SERVER_TYPE_ELK_NUM)
						}
					}
				}
			}
		} else {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_TRANSFER_LOG_ERROR_0078, agentTransferLogEnv))
		}
	}
}

// Output the contents of agent logs to stdout
func mirrorAgentLogs(ctx context.Context, wg *sync.WaitGroup, agentName string, logPathName string,
	logDNAUrl string, logDNAKey string, logType string, logServerType int16) error {
	mf, err := configureLogger(agentName, logDNAUrl, logDNAKey, logType, logServerType)
	if err != nil {
		logTermination(err)
		return err
	}

	_, err = mirrorAgentEventLogs(ctx, wg, logPathName, true, mf)
	if err != nil {
		logTermination(err)
		return err
	}
	return nil
}

// Display details of image and user
func printImageInfo() {
	// Print CPU architecture
	utils.PrintLog(fmt.Sprintf("CPU Architecture: %s", runtime.GOARCH))
	// Detect the type of container runtime we are running in. Exit if we are not
	// running inside a known container type like Docker/Kube/Oci etc.
	runtime, err := DetectRuntime()
	if err != nil && err != ErrContainerRuntimeNotFound {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_RUNTM_ERROR_OCCUR_0075, err))
		os.Exit(MFT_CONT_ERR_CODE_2)
	} else {
		// We are running in a container, so just print it on console
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_RUNTIME_NAME_0005, runtime))
	}
	utils.PrintLog(fmt.Sprintf("Base image: %s %s", os.Getenv("ENV_BASE_IMAGE_NAME"), os.Getenv("ENV_BASE_IMAGE_VERSION")))

	// Print current user.
	currentUser, err := user.Current()
	if err == nil {
		utils.PrintLog(fmt.Sprintf("Running as user ID %s with primary group %v", currentUser.Username, currentUser.Gid))
	}
	// Image creation time
	imageTime, imgErr := utils.ReadConfigurationDataFromFile("/usr/tmp/imgdetails.json")
	if imgErr == nil {
		if gjson.Get(imageTime, "imageCreateTime").Exists() {
			utils.PrintLog(fmt.Sprintf("Image created: %s", gjson.Get(imageTime, "imageCreateTime").String()))
		}
	}

	// MFT Redistributable package version
	utils.PrintLog(fmt.Sprintf("IBM MQ Managed File Transfer Redistributable Agent: %s Build Level: %s", os.Getenv("ENV_MQ_VERSION"), os.Getenv("ENV_MQ_BUILD_LEVEL")))
}
