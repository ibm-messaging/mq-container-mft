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

// Calls fteSetupCommands to create command queue manager configuration.
func SetupCommands(allAgentConfig string, bfgDataPath string, agentName string) bool {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
	utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_SETUP_STRT_0055, agentName))

	// Get the path of MFT fteSetupCommands command.
	cmdCmdsPath, lookPathErr := exec.LookPath("fteSetupCommands")
	if lookPathErr == nil {
		// Setup commands configuration
		if !gjson.Get(allAgentConfig, "commandQMgr.name").Exists() {
			utils.PrintLog("Command queue manager name not provided")
			return false
		}
		var commandQueueManager string
		var port string
		var channel string
		var hostName string
		commandQueueManager = gjson.Get(allAgentConfig, "commandQMgr.name").String()
		if gjson.Get(allAgentConfig, "commandQMgr.host").Exists() {
			hostName = gjson.Get(allAgentConfig, "commandQMgr.host").String()
		} else {
			hostName = "localhost"
		}

		if gjson.Get(allAgentConfig, "commandQMgr.port").Exists() {
			port = gjson.Get(allAgentConfig, "commandQMgr.port").String()
		} else {
			port = "1414"
		}

		if gjson.Get(allAgentConfig, "commandQMgr.channel").Exists() {
			channel = gjson.Get(allAgentConfig, "commandQMgr.channel").String()
		} else {
			channel = "SYSTEM.DEF.SVRCONN"
		}

		var cmdArgs []string
		cmdArgs = append(cmdArgs, cmdCmdsPath,
			"-p", gjson.Get(allAgentConfig, "coordinationQMgr.name").String(),
			"-connectionQMgr", commandQueueManager,
			"-connectionQMgrHost", hostName,
			"-connectionQMgrPort", port, "-connectionQMgrChannel", channel, "-f")
		if commandTracingEnabled {
			cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
			cmdTracePath := GetCommandTracePath()
			if len(cmdTracePath) > 0 {
				cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
			}
		}

		cmdSetupCmds := &exec.Cmd{
			Path: cmdCmdsPath,
			Args: cmdArgs,
		}

		cmdSetupCmds.Stdout = &outb
		cmdSetupCmds.Stderr = &errb
		// Execute the fteSetupCommands command. Log an error an exit in case of any error.
		if err := cmdSetupCmds.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
			os.Exit(1)
		} else {
			if logLevel == LOG_LEVEL_DIGANOSTIC {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_INFO_0043, outb.String()))
			}
			var cmdCredFilePath string = TEXT_BLANK
			// If a credentials file has been specified as environment variable, then set it here
			credPath, credPathSet := os.LookupEnv(MFT_CREDENTIAL_FILE)
			if credPathSet {
				credPath = strings.Trim(credPath, TEXT_TRIM)
				if credPath != TEXT_BLANK {
					if utils.DoesFileExist(credPath) {
						cmdCredFilePath = credPath
					} else {
						utils.PrintLog(fmt.Sprintf(MFT_CONT_CFG_CORD_CONFIG_CRED_NOT_EXIST_0025, credPath))
						return false
					}
				} else {
					utils.PrintLog(fmt.Sprintf(MFT_CONT_CFG_CORD_CONFIG_CRED_IGNORED_0026))
				}
			} else {
				if gjson.Get(allAgentConfig, "commandQMgr.qmgrCredentials").Exists() {
					coordinationQmgrName := gjson.Get(allAgentConfig, "coordinationQMgr.name").String()
					cmdCredFilePath = bfgDataPath + "/mqft/config/" + coordinationQmgrName + "/CmdCredentials.xml"
					if SetupCredentials(cmdCredFilePath, gjson.Get(allAgentConfig, "commandQMgr.qmgrCredentials").String(), commandQueueManager) {
						// Attempt to encrypt the credentials file with a fixed key
						EncryptCredentialsFile(cmdCredFilePath)
					}
				}
			}
			if logLevel == LOG_LEVEL_VERBOSE && len(cmdCredFilePath) > 0 {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_QMGR_CRED_PATH_0056, cmdCredFilePath))
			}

			// Update command properties file with additional attributes specified.
			commandsPropertiesFile := bfgDataPath + "/mqft/config/" + gjson.Get(allAgentConfig, "coordinationQMgr.name").String() + "/command.properties"
			UpdateProperties(commandsPropertiesFile, allAgentConfig, "commandQMgr.additionalProperties",
				"connectionQMgrAuthenticationCredentialsFile", cmdCredFilePath)
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_SETUP_COMP_0057, commandQueueManager))
			created = true
		}
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}

	return created
}

func ValidateCommandAttributes(jsonData string) error {
	// Commands queue manager is mandatory
	if !gjson.Get(jsonData, "commandQMgr.name").Exists() {
		err := errors.New(MFT_CONT_CFG_CMD_QM_NAME_MISSING_0017)
		return err
	}
	// Coordination queue manager host is mandatory
	if !gjson.Get(jsonData, "commandQMgr.host").Exists() {
		err := errors.New(MFT_CONT_CFG_CMD_QM_HOST_MISSING_0018)
		return err
	}

	return nil
}
