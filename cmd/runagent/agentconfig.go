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
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

// Call fteStartAgent command to submit a request to start an agent.
func StartAgent(agentName string, coordinationQMgr string) bool {
	var outb, errb bytes.Buffer
	var startSubmitted bool = false

	// Get the path of MFT fteStartAgent command.
	cmdStrAgntPath, lookPathErr := exec.LookPath("fteStartAgent")
	if lookPathErr == nil {
		// We are done with creating agent. Start it now.
		utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_STARTING_0041, agentName))
		var cmdArgs []string
		cmdArgs = append(cmdArgs, cmdStrAgntPath, "-p", coordinationQMgr, agentName)
		if commandTracingEnabled {
			cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
			cmdTracePath := GetCommandTracePath()
			if len(cmdTracePath) > 0 {
				cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
			}
		}

		cmdStrAgnt := &exec.Cmd{
			Path: cmdStrAgntPath,
			Args: cmdArgs,
		}

		cmdStrAgnt.Stdout = &outb
		cmdStrAgnt.Stderr = &errb
		// Run fteStartAgent command. Log and exit in case of any error.
		if err := cmdStrAgnt.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel == LOG_LEVEL_DIGANOSTIC {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_INFO_0043, outb.String()))
			}
			startSubmitted = true
		}
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}
	return startSubmitted
}

// Verify the status of agent by calling fteListAgents command.
func VerifyAgentStatus(coordinationQMgr string, agentName string) string {
	var outb, errb bytes.Buffer
	var agentStatus string

	utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_VRFY_STATUS_0044, agentName))
	cmdListAgentPath, lookPathErr := exec.LookPath("fteListAgents")
	if lookPathErr == nil {
		var cmdArgs []string
		cmdArgs = append(cmdArgs, cmdListAgentPath, "-p", coordinationQMgr, agentName)
		if commandTracingEnabled {
			cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
			cmdTracePath := GetCommandTracePath()
			if len(cmdTracePath) > 0 {
				cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
			}
		}

		cmdListAgents := &exec.Cmd{
			Path: cmdListAgentPath,
			Args: cmdArgs,
		}

		cmdListAgents.Stdout = &outb
		cmdListAgents.Stderr = &errb
		// Execute and get the output of the command into a byte buffer
		if err := cmdListAgents.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel == LOG_LEVEL_DIGANOSTIC {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_INFO_0043, outb.String()))
			}
			// Now parse the output of fteListAgents command and take appropriate actions.
			agentStatus = outb.String()
		}
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}

	return agentStatus
}

// Calls fteCreateAgent/fteCreateBridgeAgent commands to setup agent configuration
func SetupAgent(agentConfig string, bfgDataPath string, coordinationQMgr string) bool {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
	var cmdSetup bool = false
	var agentType string = AGENT_TYPE_STANDARD
	var standardAgent bool
	var agentName string

	// Get the type of the agent from configuration file. Assume type as STANDARD
	// if not specified or an invalid type was specified.
	if gjson.Get(agentConfig, "type").Exists() {
		agentType = strings.ToUpper(strings.Trim(gjson.Get(agentConfig, "type").String(), TEXT_TRIM))
		if !strings.EqualFold(agentType, AGENT_TYPE_STANDARD) && !strings.EqualFold(agentType, AGENT_TYPE_BRIDGE) {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_INVALID_TYPE_0045, agentType, AGENT_TYPE_STANDARD))
			agentType = AGENT_TYPE_STANDARD
		}
	}

	// Determine if we will be creating a standard or a bridge agent
	if agentType == AGENT_TYPE_STANDARD {
		standardAgent = true
	} else {
		standardAgent = false
	}

	agentName = gjson.Get(agentConfig, "name").String()
	utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_CREATING_0046, agentType, agentName))

	var cmdCrtAgnt *exec.Cmd
	// Cance the agent attributes
	agentQMgrName := gjson.Get(agentConfig, "qmgrName").String()
	agentQMgrHost := gjson.Get(agentConfig, "qmgrHost").String()
	var agentQMgrPort string
	if gjson.Get(agentConfig, "qmgrPort").Exists() {
		agentQMgrPort = gjson.Get(agentConfig, "qmgrPort").String()
	} else {
		// Not found, use default 1414
		agentQMgrPort = "1414"
	}

	var agentQMgrChannel string
	if gjson.Get(agentConfig, "qmgrChannel").Exists() {
		agentQMgrChannel = gjson.Get(agentConfig, "qmgrChannel").String()
	} else {
		// Default to SYSTEM.DEF.SVRCONN
		agentQMgrChannel = "SYSTEM.DEF.SVRCONN"
	}

	// We are creating a STANDARD agent
	if standardAgent {
		// Get the path of MFT fteCreateAgent command.
		cmdCrtAgntPath, lookPathErr := exec.LookPath("fteCreateAgent")
		if lookPathErr == nil {
			// Creating a standard agent
			var params []string
			params = append(params, cmdCrtAgntPath,
				"-p", coordinationQMgr,
				"-agentName", agentName,
				"-agentQMgr", agentQMgrName,
				"-agentQMgrHost", agentQMgrHost,
				"-agentQMgrPort", agentQMgrPort,
				"-agentQMgrChannel", agentQMgrChannel, "-f")
			if commandTracingEnabled {
				params = append(params, "-trace", "com.ibm.wmqfte=all")
				cmdTracePath := GetCommandTracePath()
				if len(cmdTracePath) > 0 {
					params = append(params, "-tracePath", cmdTracePath)
				}
			}

			// Now build the command to create standard agent
			cmdCrtAgnt = &exec.Cmd{Path: cmdCrtAgntPath, Args: params}
			cmdSetup = true
		} else {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
		}
	} else {
		// We are creating a BRIDGE agent
		// Get the path of MFT fteCreateBridgeAgent command
		cmdCrtBridgeAgntPath, lookPathErr := exec.LookPath("fteCreateBridgeAgent")
		if lookPathErr == nil {
			// Creating a bridge agent
			var params []string
			params = append(params, cmdCrtBridgeAgntPath,
				"-p", coordinationQMgr,
				"-agentName", agentName,
				"-agentQMgr", agentQMgrName,
				"-agentQMgrHost", agentQMgrHost,
				"-agentQMgrPort", agentQMgrPort,
				"-agentQMgrChannel", agentQMgrChannel, "-f")
			if commandTracingEnabled {
				params = append(params, "-trace", "com.ibm.wmqfte=all")
				cmdTracePath := GetCommandTracePath()
				if len(cmdTracePath) > 0 {
					params = append(params, "-tracePath", cmdTracePath)
				}
			}

			protocolBridgeConfigs := gjson.Get(agentConfig, "protocolBridge").Array()
			for i := 0; i < len(protocolBridgeConfigs); i++ {
				singleBridgeConfig := protocolBridgeConfigs[i].String()
				params = processBridgeParameters(singleBridgeConfig, params)
			}

			// Now build the command to create a bridge agent
			cmdCrtAgnt = &exec.Cmd{
				Path: cmdCrtBridgeAgntPath,
				Args: params,
			}
			cmdSetup = true
		} else {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
		}
	}

	// Ready to execute the command
	if cmdSetup == true {
		// Reuse the same buffer
		cmdCrtAgnt.Stdout = &outb
		cmdCrtAgnt.Stderr = &errb

		// Execute the fteCreateAgent/fteCreateBridgeAgent to create agent configuration.
		// Log an error an exit in case of any error.
		if err := cmdCrtAgnt.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			// If it is bridge agent, then update the ProtocolBridgeProperties file with any additional properties specified.
			if !standardAgent {
				// Copy the custom credentials exit to agent's exit directory.
				protocolBridgeCustExit := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/exits/" + PBA_CUSTOM_CRED_EXIT_NAME
				utils.CopyFile(PBA_CUSTOM_CRED_EXIT, protocolBridgeCustExit)
				protocolBridgeCustExitDependLib := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/exits/" + PBA_CUSTOM_CRED_DEPEND_LIB_NAME
				utils.CopyFile(PBA_CUSTOM_CRED_DEPEND_LIB, protocolBridgeCustExitDependLib)

				// Do we have user supplied bridge properties file? If yes, copy it to agent's directory,
				// overwrite the existing one.
				bridgePropsFileUserSupplied, bridgePropFileSet := os.LookupEnv(MFT_BRIDGE_PROPERTIES_FILE)
				if bridgePropFileSet {
					bridgePropsFileUserSupplied = strings.Trim(bridgePropsFileUserSupplied, TEXT_TRIM)
					if bridgePropsFileUserSupplied != TEXT_BLANK {
						bridgePropsDest := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/ProtocolBridgeProperties.xml"
						utils.CopyFile(bridgePropsFileUserSupplied, bridgePropsDest)
					}
				}
			}

			// Check if credentials override has been specified in environment. If not,
			// use the UID/PWD provided in agent configuration file
			_, agentCredPathSet := os.LookupEnv(MFT_CREDENTIAL_FILE)
			if !agentCredPathSet {
				// Configuration file has not been provided. Make an attempt to
				// read UID/PWD from agent configuration JSON file and create
				// the MQMFTCredentials file place it in agent's config directory.
				// The credentials file will be encrypted using a fixed key.
				if gjson.Get(agentConfig, "qmgrCredentials").Exists() {
					agentCredFilePath := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/MQMFTCredentials.xml"
					if SetupCredentials(agentCredFilePath, gjson.Get(agentConfig, "qmgrCredentials").String(), agentQMgrName) {
						// Attempt to encrypt the credentials file with a fixed key
						EncryptCredentialsFile(agentCredFilePath)
						os.Setenv(MFT_CREDENTIAL_FILE, agentCredFilePath)
					}
				}
			}

			// Update agent properties file with additional attributes specified.
			agentPropertiesFile := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/agent.properties"
			updateAgentProperties(agentPropertiesFile, agentConfig, "additionalProperties", !standardAgent)

			// Update UserSandbox XML file - valid only for STANDARD agents
			if standardAgent {
				CreateUserSandbox(bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/UserSandboxes.xml")
			}
			// Tell user that agent has been configured.
			utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_CREATED_0047, agentName))
			created = true
		}
	}
	return created
}

// Read and process protocol bridge server attributes from configuration JSON file
func processBridgeParameters(bridgeProperties string, params []string) []string {
	if logLevel == LOG_LEVEL_DIGANOSTIC {
		utils.PrintLog("Bridge properties: " + bridgeProperties)
	}
	// Set protocol server type
	serverType := gjson.Get(bridgeProperties, "serverType")
	if serverType.Exists() {
		// Use the supplied one.
		params = append(params, "-bt", serverType.String())
	} else {
		// Otherwise use default as FTP
		params = append(params, "-bt", "FTP")
	}

	// Set the protocol server host
	serverHost := gjson.Get(bridgeProperties, "serverHost")
	if serverHost.Exists() {
		params = append(params, "-bh", serverHost.String())
	} else {
		// Use local host if host name is not specified.
		params = append(params, "-bh", "localhost")
	}

	// Set the protocol server timezone, valid only for FTP and FTPS server
	if serverType.String() != "SFTP" {
		serverTimezone := gjson.Get(bridgeProperties, "serverTimezone")
		if serverTimezone.Exists() {
			params = append(params, "-btz", serverTimezone.String())
		}
	}

	// Set the protocol server platform.
	serverPlatform := gjson.Get(bridgeProperties, "serverPlatform")
	if serverPlatform.Exists() {
		params = append(params, "-bm", serverPlatform.String())
	}

	// Set the protocol server locale
	if serverType.String() != "SFTP" {
		serverLocale := gjson.Get(bridgeProperties, "serverLocale")
		if serverLocale.Exists() {
			params = append(params, "-bsl", serverLocale.String())
		}
	}

	// Set protocol server file encoding
	serverFileEncoding := gjson.Get(bridgeProperties, "serverFileEncoding")
	if serverFileEncoding.Exists() {
		params = append(params, "-bfe", serverFileEncoding.String())
	}

	// Set the protocol server port
	serverPort := gjson.Get(bridgeProperties, "serverPort")
	if serverPort.Exists() {
		params = append(params, "-bp", serverPort.String())
	}

	// Set the protocol server trust store file
	serverTrustStoreFile := gjson.Get(bridgeProperties, "serverTrustStoreFile")
	if serverTrustStoreFile.Exists() {
		params = append(params, "-bts", serverTrustStoreFile.String())
	}

	// Set protocol server limited write flag
	serverLimitedWrite := gjson.Get(bridgeProperties, "serverLimitedWrite")
	if serverLimitedWrite.Exists() {
		params = append(params, "-blw", serverLimitedWrite.String())
	}

	// Set the protocol server list format.
	serverListFormat := gjson.Get(bridgeProperties, "serverListFormat")
	if serverListFormat.Exists() {
		params = append(params, "-blf", serverListFormat.String())
	}

	return params
}

func ValidateAgentAttributes(jsonData string) error {
	// Agent name is mandatory
	if !gjson.Get(jsonData, "name").Exists() {
		err := errors.New(MFT_CONT_CFG_AGENT_NAME_MISSING_0020)
		return err
	}

	// Agent queue manager name is mandatory
	if !gjson.Get(jsonData, "qmgrName").Exists() {
		err := errors.New(MFT_CONT_CFG_AGENT_QM_NAME_MISSING_0021)
		return err
	}

	// Agent queue manager host is mandatory
	if !gjson.Get(jsonData, "qmgrHost").Exists() {
		err := errors.New(MFT_CONT_CFG_AGENT_QM_HOST_MISSING_0022)
		return err
	}

	return nil
}

// Update agent.properties file with any additional properties specified in
// configuration JSON file.
func updateAgentProperties(propertiesFile string, agentConfig string, sectionName string, bridgeAgent bool) {
	f, err := os.OpenFile(propertiesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_OPN_FILE_0067, propertiesFile, err))
		return
	}
	defer f.Close()

	// Enable logCapture by default. Customer can turn off by specifying it again in config map
	if _, err := f.WriteString("logCapture=true\n"); err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
	}

	// Set maximum restart count to 0, so that the container ends immediately
	// if the first attempt fails.
	if _, err := f.WriteString("maxRestartCount=0\n"); err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
	}

	if gjson.Get(agentConfig, sectionName).Exists() {
		result := gjson.Get(agentConfig, sectionName)
		result.ForEach(func(key, value gjson.Result) bool {
			if _, err := f.WriteString(key.String() + "=" + value.String() + "\n"); err != nil {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
			}
			return true // keep iterating
		})
	}

	// If agent credentials file has been specified as environment variable, then set it here
	agentCredPath, agentCredPathSet := os.LookupEnv(MFT_CREDENTIAL_FILE)
	if agentCredPathSet && agentCredPath != TEXT_BLANK {
		if _, err := f.WriteString("agentQMgrAuthenticationCredentialsFile=" + agentCredPath + "\n"); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
		}
	}

	// If this is a bridge agent, then configure custom exit
	if bridgeAgent {
		bridgeCredPath, bridgeCredPathSet := os.LookupEnv(MFT_BRIDGE_CREDENTIAL_FILE)
		if bridgeCredPathSet && bridgeCredPath != TEXT_BLANK {
			if _, err := f.WriteString("protocolBridgeCredentialConfiguration=" + bridgeCredPath + "\n"); err != nil {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
			}
		}

		if _, err := f.WriteString("protocolBridgeCredentialExitClasses=com.ibm.wmq.bridgecredentialexit.ProtocolBridgeCustomCredentialExit\n"); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
		}
	} else {
		if _, err := f.WriteString("userSandboxes=true"); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
		}
	}
}

// Updates ProtocolBridgeProperties file with specified additional attributes
func updateProtocolBridgePropertiesFile(propertiesFile string, agentConfig string, sectionName string) {
	// First read the entire contents of the ProtocolBridgeProperties file.
	bridgeProperites := readFileContents(propertiesFile)
	if len(bridgeProperites) > 0 {
		// Find the last index of ProtocolBridgeProperties.xsd in the file contents.
		// We will be inserting new attributes of that "ProtocolBridgeProperties.xsd>"
		lastIndex := strings.LastIndex(bridgeProperites, "ProtocolBridgeProperties.xsd")
		// Add 30 to index because of the length of "ProtocolBridgeProperties.xsd>" is 30.
		insertIndex := lastIndex + 30

		f, err := os.OpenFile(propertiesFile, os.O_WRONLY, 0644)
		if err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
			return
		}
		defer f.Close()

		if gjson.Get(agentConfig, sectionName).Exists() {
			result := gjson.Get(agentConfig, sectionName)
			result.ForEach(func(key, value gjson.Result) bool {
				// We support only two properties that can be set from config file. Others are ignored.
				if strings.EqualFold(key.String(), "credentialsFile") == true {
					insertString := "\n<tns:credentialsFile path=\"" + value.String() + "\" />"
					bridgeProperites = bridgeProperites[:insertIndex] + insertString + bridgeProperites[insertIndex:]
					insertIndex += len(insertString)
				} else if strings.EqualFold(key.String(), "credentialsKeyFile") == true {
					//<tns:credentialsKeyFile path="c:\temp\agentinitkey.key"/>
					insertString := "\n<tns:credentialsKeyFile path=\"" + value.String() + "\" />"
					bridgeProperites = bridgeProperites[:insertIndex] + insertString + bridgeProperites[insertIndex:]
					insertIndex += len(insertString)
				}
				return true // go on till we insert all properties
			})
		}

		// Print the updated contents of the ProtocolBridgeProperties.xml file
		if logLevel == LOG_LEVEL_DIGANOSTIC {
			utils.PrintLog(bridgeProperites)
		}
		// Write the updated properties to file.
		_, writeErr := f.WriteString(bridgeProperites)
		if writeErr != nil {
			utils.PrintLog(fmt.Sprintf("%v", writeErr))
		}
	}
}

// Clean agent before starting it.
func cleanAgent(agentConfig string, coordinationQMgr string, agentName string) {
	cleanOnStart := gjson.Get(agentConfig, "cleanOnStart")
	if cleanOnStart.Exists() {
		cleanItem := cleanOnStart.String()
		if cleanItem == "transfers" {
			cleanAgentItem(coordinationQMgr, agentName, cleanItem, "-trs")
		} else if cleanItem == "monitors" {
			cleanAgentItem(coordinationQMgr, agentName, cleanItem, "-ms")
		} else if cleanItem == "scheduledTransfers" {
			cleanAgentItem(coordinationQMgr, agentName, cleanItem, "-ss")
		} else if cleanItem == "invalidMessages" {
			cleanAgentItem(coordinationQMgr, agentName, cleanItem, "-ims")
		} else if cleanItem == "all" {
			cleanAgentItem(coordinationQMgr, agentName, cleanItem, "-all")
		} else {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_CLN_0048, cleanItem))
		}
	}
}

// Unregister and delete agent
func deleteAgent(coordinationQMgr string, agentName string) error {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_DLTNG_0049, agentName))

	// Get the path of MFT fteDeleteAgent command.
	cmdDltAgentPath, lookErr := exec.LookPath("fteDeleteAgent")
	if lookErr != nil {
		return lookErr
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, cmdDltAgentPath, "-p", coordinationQMgr, "-f", agentName)
	if commandTracingEnabled {
		cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
		cmdTracePath := GetCommandTracePath()
		if len(cmdTracePath) > 0 {
			cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
		}
	}

	// -f force option is not used so that monitor is not recreated if it already exists.
	cmdDltAgentCmd := &exec.Cmd{
		Path: cmdDltAgentPath,
		Args: cmdArgs,
	}

	// Reuse the same buffer
	cmdDltAgentCmd.Stdout = &outb
	cmdDltAgentCmd.Stderr = &errb
	// Execute the fteDeleteAgent command. Log an error an exit in case of any error.
	if err := cmdDltAgentCmd.Run(); err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_DLTED_0050, agentName))
	}
	return nil
}

// Clean agent on start of container.
func cleanAgentItem(coordinationQMgr string, agentName string, item string, option string) error {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_CLN_0051, item, agentName))

	// Get the path of MFT fteCleanAgent command.
	cmdCleanAgentPath, lookErr := exec.LookPath("fteCleanAgent")
	if lookErr != nil {
		return lookErr
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, cmdCleanAgentPath, "-p", coordinationQMgr, option, agentName)
	if commandTracingEnabled {
		cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
		cmdTracePath := GetCommandTracePath()
		if len(cmdTracePath) > 0 {
			cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
		}
	}

	// Clean agent before starting
	cmdCleanAgentCmd := &exec.Cmd{
		Path: cmdCleanAgentPath,
		Args: cmdArgs,
	}

	// Reuse the same buffer
	cmdCleanAgentCmd.Stdout = &outb
	cmdCleanAgentCmd.Stderr = &errb
	// Execute the fteCleanAgent command. Log an error an exit in case of any error.
	if err := cmdCleanAgentCmd.Run(); err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
	} else {
		if logLevel == LOG_LEVEL_DIGANOSTIC {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, outb.String()))
		}
		if item == "all" {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_ALL_ITEM_CLN_0076, agentName))
		} else {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_ITEM_CLN_0052, item, agentName))
		}
	}
	return nil
}

// Create resource monitor
func createResourceMonitor(coordinationQMgr string, agentName string, agentQMgr string,
	monitorName string, fileName string) error {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_RM_CRT_0053, monitorName))

	// Get the path of MFT fteCreateAgent command.
	cmdCrtMonitorPath, lookErr := exec.LookPath("fteCreateMonitor")
	if lookErr != nil {
		return lookErr
	}
	// -f force option is not used so that monitor is not recreated if it already exists.
	cmdCrtMonitorCmd := &exec.Cmd{
		Path: cmdCrtMonitorPath,
		Args: []string{cmdCrtMonitorPath, "-p", coordinationQMgr,
			"-mm", agentQMgr,
			"-ma", agentName,
			"-mn", monitorName,
			"-ix", fileName},
	}

	// Reuse the same buffer
	cmdCrtMonitorCmd.Stdout = &outb
	cmdCrtMonitorCmd.Stderr = &errb
	// Execute the fteSetupCommands command. Log an error an exit in case of any error.
	if err := cmdCrtMonitorCmd.Run(); err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
		return nil
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, outb.String()))
	}
	return nil
}

// Returns the contents of the specified file.
func readFileContents(propertiesFile string) string {
	// Open our xmlFile
	bridgeProperiesXmlFile, err := os.Open(propertiesFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		if logLevel == LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf("%v", err))
			return TEXT_BLANK
		}
	}

	// defer the closing of our xml file so that we can parse it later on
	defer bridgeProperiesXmlFile.Close()

	xmlData, _ := ioutil.ReadAll(bridgeProperiesXmlFile)
	xmlText := string(xmlData)
	return xmlText
}

/**
Ping the agent to determine if it's ready to process transfer requests
*/
func PingAgent(coordinationQMgr string, agentName string, waitTime string) bool {
	var outb, errb bytes.Buffer
	retVal := false

	utils.PrintLog(fmt.Sprintf(MFT_CONT_AGNT_VRFY_STATUS_0044, agentName))
	cmdPingAgentPath, lookPathErr := exec.LookPath("ftePingAgent")
	if lookPathErr == nil {
		var cmdArgs []string
		cmdArgs = append(cmdArgs, cmdPingAgentPath, "-p", coordinationQMgr, agentName, "-w", waitTime)
		if commandTracingEnabled {
			cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
			cmdTracePath := GetCommandTracePath()
			if len(cmdTracePath) > 0 {
				cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
			}
		}

		cmdPingAgentPath := &exec.Cmd{
			Path: cmdPingAgentPath,
			Args: cmdArgs,
		}

		cmdPingAgentPath.Stdout = &outb
		cmdPingAgentPath.Stderr = &errb
		if err := cmdPingAgentPath.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel == LOG_LEVEL_DIGANOSTIC {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_INFO_0043, outb.String()))
			}
			// The output must contain BFGCL0793I. Ideally we should check
			// for return code of 0 from command execution. Need to figure
			// out a way to do that in Go.
			if strings.Contains(outb.String(), "BFGCL0793I:") == true {
				retVal = true
			}
		}
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}
	return retVal
}
