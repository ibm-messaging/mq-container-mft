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
	"os"
	"fmt"
    "github.com/tidwall/gjson"
	"github.com/ibm-messaging/mq-container-mft/cmd/utils"
)

/*
* This file contains the source code for the readiness probe. The
* readiness probe is program that tests if an IBM MQ Managed File
* Transfer agent is ready or not. The program scans the output0.log
* of agent for BFGAG0059I and BFGAG0191I events. Returns true if
* any of the above events are found in the log file else false.
* Based on the return value of this probe, a container orchestration
* platform like Kubernetes can recycle an agent.
 */
func main () {
	var bfgDataPath string
	var bfgConfigFilePath string 
	var agentConfig string 
	var e error
	//Name of the agent is retrieved from environment variable MFT_AGENT_NAME
	var agentNameEnv string = os.Getenv("MFT_AGENT_NAME")

	/*
	 * Read the name of an agent configuration file from environment
	 * variable MFT_AGENT_CONFIG_FILE
	 */
	bfgConfigFilePath = os.Getenv("MFT_AGENT_CONFIG_FILE")
	// Read agent configuration data from JSON file.
	agentConfig, e = utils.ReadConfigurationDataFromFile(bfgConfigFilePath)
	// Exit if we had any error when reading configuration file
	if e != nil {
		utils.PrintLog(fmt.Sprintf("%v", e))
		os.Exit(1)
	}
	
    // Get path from environment variable
	bfgConfigMountPath := os.Getenv("BFG_DATA")
	if len(bfgConfigMountPath) > 0 {
		bfgDataPath = bfgConfigMountPath
	} else {
	   // BFG_DATA is not set. So fixed path
	   bfgDataPath = utils.FIXED_BFG_DATAPATH
    }
	
	coordinationQMgr := gjson.Get(agentConfig, "coordinationQMgr.name").String()

	// Read the agentPid file from the agent logs directory
	agentPidPath := bfgDataPath + "/mqft/logs/" + coordinationQMgr  + "/agents/" + agentNameEnv + "/agent.pid"
	// Open agent.pid file and read the pid from the file.
	agentPid := utils.GetAgentPid(agentPidPath)
	if agentPid > 1 {
		agentRunning, err := utils.IsAgentRunning(agentPid)
		if err != nil {
			utils.PrintLog("Agent is not running")
			os.Exit(1)
		} else {
			if agentRunning {
				utils.PrintLog("Readiness Probe: Agent PID is valid")
				// Agent is running, so check if it is ready.
				if utils.IsAgentReady(bfgDataPath, agentNameEnv, coordinationQMgr) == true{
					utils.PrintLog("Readiness Probe: Agent ready event found")
					os.Exit(0)
				} else {
					utils.PrintLog("Readiness Probe: Agent ready event not found")
					os.Exit(1)
				}
			} else {
				utils.PrintLog("Agent is not running")
				os.Exit(1)
			}
		}
	} else {
		utils.PrintLog("Agent is not running")
		os.Exit(1)
	}
}