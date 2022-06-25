/*
© Copyright IBM Corporation 2022, 2022

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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

// Setup coordination configuration for agent.
func SetupCoordination(allAgentConfig string, bfgDataPath string, agentNameEnv string) bool {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
	coordinationQueueManagerName := gjson.Get(allAgentConfig, "coordinationQMgr.name").String()
	// Setup coordination configuration
	utils.PrintLog(fmt.Sprintf(MFT_CONT_CFG_CORD_CONFIG_MSG_0024, coordinationQueueManagerName, agentNameEnv))

	// Get the path of MFT fteSetupCoordination command.
	cmdCoordPath, lookPathErr := exec.LookPath("fteSetupCoordination")
	if lookPathErr == nil {
		var port string
		var channel string
		var host string

		if gjson.Get(allAgentConfig, "coordinationQMgr.host").Exists() {
			host = gjson.Get(allAgentConfig, "coordinationQMgr.host").String()
		} else {
			host = "localhost"
		}

		if gjson.Get(allAgentConfig, "coordinationQMgr.port").Exists() {
			port = gjson.Get(allAgentConfig, "coordinationQMgr.port").String()
		} else {
			port = "1414"
		}

		if gjson.Get(allAgentConfig, "coordinationQMgr.channel").Exists() {
			channel = gjson.Get(allAgentConfig, "coordinationQMgr.channel").String()
		} else {
			channel = "SYSTEM.DEF.SVRCONN"
		}

		var cmdArgs []string
		cmdArgs = append(cmdArgs, cmdCoordPath,
			"-coordinationQMgr", coordinationQueueManagerName,
			"-coordinationQMgrHost", host,
			"-coordinationQMgrPort", port, "-coordinationQMgrChannel", channel, "-f", "-default")
		if commandTracingEnabled {
			cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
			cmdTracePath := GetCommandTracePath()
			if len(cmdTracePath) > 0 {
				cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
			}
		}

		cmdSetupCoord := &exec.Cmd{
			Path: cmdCoordPath,
			Args: cmdArgs,
		}

		// Execute the fteSetupCoordination command. Log an error an exit in case of any error.
		cmdSetupCoord.Stdout = &outb
		cmdSetupCoord.Stderr = &errb
		if err := cmdSetupCoord.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel == LOG_LEVEL_DIGANOSTIC {
				utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			}

			// Update coordination properties file with additional attributes specified.
			coordinationPropertiesFile := bfgDataPath + "/mqft/config/" + coordinationQueueManagerName + "/coordination.properties"
			var coordCredFilePath string = TEXT_BLANK
			// If a credentials file has been specified as environment variable, then set it here
			credPath, credPathSet := os.LookupEnv(MFT_CREDENTIAL_FILE)
			if credPathSet {
				credPath = strings.Trim(credPath, TEXT_TRIM)
				if credPath != TEXT_BLANK {
					if utils.DoesFileExist(credPath) {
						coordCredFilePath = credPath
					} else {
						utils.PrintLog(fmt.Sprintf(MFT_CONT_CFG_CORD_CONFIG_CRED_NOT_EXIST_0025, credPath))
						return false
					}
				} else {
					utils.PrintLog(MFT_CONT_CFG_CORD_CONFIG_CRED_IGNORED_0026)
				}
			} else {
				if gjson.Get(allAgentConfig, "coordinationQMgr.qmgrCredentials").Exists() {
					// Update coordination queue manager credentials file
					// Configuration file has not been provided. Make an attempt to
					// read UID/PWD from agent configuration JSON file and create
					// the MQMFTCredentials file place it in agent's config directory.
					// The credentials file will be encrypted using a fixed key.
					coordCredFilePath = bfgDataPath + "/mqft/config/" + coordinationQueueManagerName + "/CoordCredentials.xml"
					if SetupCredentials(coordCredFilePath, gjson.Get(allAgentConfig, "coordinationQMgr.qmgrCredentials").String(),
						coordinationQueueManagerName) {
						// Attempt to encrypt the credentials file with a fixed key
						EncryptCredentialsFile(coordCredFilePath)
					}
				}
			}
			if logLevel == LOG_LEVEL_VERBOSE && len(coordCredFilePath) > 0 {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_CFG_CORD_CONFIG_CRED_PATH_0027, coordCredFilePath))
			}

			UpdateProperties(coordinationPropertiesFile, allAgentConfig, "coordinationQMgr.additionalProperties",
				"coordinationQMgrAuthenticationCredentialsFile", coordCredFilePath)
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CORD_SETUP_COMP_0054, coordinationQueueManagerName))
			created = true
		}
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}

	return created
}

// Validate attributes in JSON file.
// Check if the configuration JSON contains all required attribtes
func ValidateCoordinationAttributes(jsonData string) error {
	// Coordination queue manager is mandatory
	if !gjson.Get(jsonData, "coordinationQMgr.name").Exists() {
		err := errors.New(MFT_CONT_CFG_CORD_QM_NAME_MISSING_0014)
		return err
	}

	// Coordination queue manager host is mandatory
	if !gjson.Get(jsonData, "coordinationQMgr.host").Exists() {
		err := errors.New(MFT_CONT_CFG_CORD_QM_HOST_MISSING_0015)
		return err
	}

	return nil
}
