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
	"fmt"
	"os"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

const AGENT_REDY_EXIT_CODE_0 = 0
const AGENT_REDY_EXIT_CODE_1 = 1
const AGENT_REDY_EXIT_CODE_2 = 2
const AGENT_REDY_EXIT_CODE_3 = 3
const AGENT_REDY_EXIT_CODE_4 = 4
const AGENT_REDY_EXIT_CODE_5 = 5
const AGENT_REDY_EXIT_CODE_6 = 6
const AGENT_REDY_EXIT_CODE_7 = 7

/*
* This file contains the source code for the readiness probe. The
* readiness probe is program that tests if an IBM MQ Managed File
* Transfer agent is ready or not. The program scans the output0.log
* of agent for BFGAG0059I and BFGAG0191I events. Returns true if
* any of the above events are found in the log file else false.
* Based on the return value of this probe, a container orchestration
* platform like Kubernetes can recycle an agent.
 */
func main() {
	var bfgDataPath string
	var bfgConfigFilePath string
	var agentConfig string
	var e error

	//Name of the agent is retrieved from environment variable MFT_AGENT_NAME
	agentNameEnv, agentNameEnvSet := os.LookupEnv("MFT_AGENT_NAME")
	if !agentNameEnvSet {
		utils.PrintLog(utils.AGENT_REDY_ENV_AGENT_NAME_NOT_SET_3001)
		os.Exit(AGENT_REDY_EXIT_CODE_1)
	}

	/*
	 * Read the name of an agent configuration file from environment
	 * variable MFT_AGENT_CONFIG_FILE
	 */
	bfgConfigFilePath, bfgConfigFilePathSet := os.LookupEnv("MFT_AGENT_CONFIG_FILE")
	if !bfgConfigFilePathSet {
		// MFT_AGENT_CONFIG_FILE environment variable not specified. Looking for
		// config.json file in /run/mqmft directory.
		msg := fmt.Sprintf(utils.MFT_ENV_AGNT_CFG_FILE_NOT_SPECIFIED, utils.MFT_DEFAULT_CONFIG_JSON)
		utils.PrintLog(msg)
		// Assign the default config filename, so that rest of the processing goes on.
		bfgConfigFilePath = utils.MFT_DEFAULT_CONFIG_JSON
	} else {
		bfgConfigFilePath = strings.TrimSpace(bfgConfigFilePath)
		if bfgConfigFilePath == "" {
			utils.PrintLog(utils.MFT_CONT_ENV_AGNT_CFG_FILE_BLANK_0012)
			os.Exit(AGENT_REDY_EXIT_CODE_7)
		}
	}

	// Read agent configuration data from JSON file.
	agentConfig, e = utils.ReadConfigurationDataFromFile(bfgConfigFilePath)
	// Exit if we had any error when reading configuration file
	if e != nil {
		utils.PrintLog(fmt.Sprintf(utils.AGENT_REDY_ENV_CFG_FILE_READ_3003, bfgConfigFilePath, e))
		os.Exit(AGENT_REDY_EXIT_CODE_3)
	}

	// Get path from environment variable
	bfgConfigMountPath, bfgConfigMountPathSet := os.LookupEnv("BFG_DATA")
	if bfgConfigMountPathSet {
		bfgConfigMountPath = strings.Trim(bfgConfigMountPath, "")
		if len(bfgConfigMountPath) > 0 {
			bfgDataPath = bfgConfigMountPath
		} else {
			// BFG_DATA is not set. So fixed path
			bfgDataPath = utils.FIXED_BFG_DATAPATH
		}
	} else {
		// BFG_DATA is not set. So fixed path
		bfgDataPath = utils.FIXED_BFG_DATAPATH
	}

	coordinationQMgr := gjson.Get(agentConfig, "coordinationQMgr.name").String()
	// Read the agentPid file from the agent logs directory
	agentPidPath := bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentNameEnv + "/agent.pid"
	// Open agent.pid file and read the pid from the file.
	agentPid, _ := utils.GetAgentPid(agentPidPath)
	if agentPid > 1 {
		agentRunning, err := utils.IsAgentRunning(agentPid)
		if err != nil {
			utils.PrintLog(fmt.Sprintf(utils.AGENT_REDY_NOT_RUNNING_3004, agentNameEnv))
			os.Exit(AGENT_REDY_EXIT_CODE_4)
		} else {
			if agentRunning {
				// Agent is running, so check if it is ready.
				agentStatus, _ := utils.IsAgentReady(bfgDataPath, agentNameEnv, coordinationQMgr)
				if agentStatus {
					os.Exit(AGENT_REDY_EXIT_CODE_0)
				} else {
					utils.PrintLog(utils.AGENT_REDY_EVNT_NOT_FOUND_3005)
					os.Exit(AGENT_REDY_EXIT_CODE_5)
				}
			} else {
				utils.PrintLog(fmt.Sprintf(utils.AGENT_REDY_NOT_RUNNING_3004, agentNameEnv))
				os.Exit(AGENT_REDY_EXIT_CODE_6)
			}
		}
	} else {
		utils.PrintLog(fmt.Sprintf(utils.AGENT_REDY_NOT_RUNNING_3004, agentNameEnv))
		os.Exit(AGENT_REDY_EXIT_CODE_7)
	}
}
