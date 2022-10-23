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
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/subchen/go-xmldom"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Call fteStartAgent command to submit a request to start an agent.
func StartAgent(agentName string, coordinationQMgr string) bool {
	var outb, errb bytes.Buffer
	var startSubmitted bool = false

	// Get the path of MFT fteStartAgent command.
	cmdStrAgntPath, lookPathErr := exec.LookPath("fteStartAgent")
	if lookPathErr == nil {
		// We are done with creating agent. Start it now.
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_STARTING_0041, agentName))
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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_INFO_0043, outb.String()))
			}
			startSubmitted = true
		}
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}
	return startSubmitted
}

// Verify the status of agent by calling fteListAgents command.
func VerifyAgentStatus(coordinationQMgr string, agentName string) string {
	var outb, errb bytes.Buffer
	var agentStatus string

	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_VRFY_STATUS_0044, agentName))
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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_INFO_0043, outb.String()))
			}
			// Now parse the output of fteListAgents command and take appropriate actions.
			agentStatus = outb.String()
		}
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
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
	var protocolBridgeConfigs []gjson.Result

	// Get the type of the agent from configuration file. Assume type as STANDARD
	// if not specified or an invalid type was specified.
	if gjson.Get(agentConfig, "type").Exists() {
		agentType = strings.ToUpper(strings.Trim(gjson.Get(agentConfig, "type").String(), TEXT_TRIM))
		if !strings.EqualFold(agentType, AGENT_TYPE_STANDARD) && !strings.EqualFold(agentType, AGENT_TYPE_BRIDGE) {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_INVALID_TYPE_0045, agentType, AGENT_TYPE_STANDARD))
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
	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_CREATING_0046, agentType, agentName))

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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
		}
	} else {
		// Initialize BridgeProperties.
		BuildBridgePropertyList()

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

			protocolBridgeConfigs = gjson.Get(agentConfig, "protocolServers").Array()
			// For creating the agent, take the first element in the array. We will updated the
			// ProtocolBridgeProperties.xml with other elements in the array.
			if len(protocolBridgeConfigs) > 0 {
				singleBridgeConfig := protocolBridgeConfigs[0].String()
				if updateBridgeParameters(singleBridgeConfig, params) {
					// Now build the command to create a bridge agent
					cmdCrtAgnt = &exec.Cmd{
						Path: cmdCrtBridgeAgntPath,
						Args: params,
					}
					cmdSetup = true
				} else {
					utils.PrintLog(utils.MFT_CONT_BRIDGE_NOT_ENOUGH_INFO)
				}
			} else {
				utils.PrintLog(utils.MFT_CONT_BRIDGE_NOT_ENOUGH_INFO)
			}
		} else {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			// If it is bridge agent, then update the ProtocolBridgeProperties file with any additional properties specified.
			if !standardAgent {
				// Copy the custom credentials exit to agent's exit directory.
				protocolBridgeCustExit := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQMgr + MFT_AGENTS_SLASH + agentName + MFT_EXITS_SLASH + PBA_CUSTOM_CRED_EXIT_NAME
				utils.CopyFile(PBA_CUSTOM_CRED_EXIT, protocolBridgeCustExit)
				protocolBridgeCustExitDependLib := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQMgr + MFT_AGENTS_SLASH + agentName + MFT_EXITS_SLASH + PBA_CUSTOM_CRED_DEPEND_LIB_NAME
				utils.CopyFile(PBA_CUSTOM_CRED_DEPEND_LIB, protocolBridgeCustExitDependLib)
			}

			// Configuration file has not been provided. Make an attempt to read UID/PWD from agent configuration JSON file and create
			// the MQMFTCredentials file place it in agent's config directory. The credentials file will be encrypted using a fixed key.
			agentCredFilePath := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQMgr + MFT_AGENTS_SLASH + agentName + MFT_AGENT_CRED_SLASH
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog("Agent credentials file path " + agentCredFilePath)
			}

			// Update coordination properties file with additional attributes specified.
			agentPropertiesFile := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQMgr + MFT_AGENTS_SLASH + agentName + MFT_AGENT_PROPS_SLASH
			protocolBridgePropertiesFile := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQMgr + MFT_AGENTS_SLASH + agentName + MFT_PBA_PROPS_SLASH

			// Start XML document for credentials file
			credentialsDoc := InitializeCredentialsDocumentWriter()
			// Configure TLS for agent connections
			created, agentConfig = configTLSAgent(agentConfig, credentialsDoc, agentCredFilePath)

			if created {
				if gjson.Get(agentConfig, "qmgrCredentials").Exists() {
					// Write agent queue manager credentials
					err := UpdateXmlWithQmgrCredentials(credentialsDoc, gjson.Get(agentConfig, "qmgrCredentials").String(), agentQMgrName)
					if err != nil {
						if logLevel >= LOG_LEVEL_VERBOSE {
							utils.PrintLog(err.Error())
						}
					}
				}

				// Create credentials file for agent.
				errorSetCred := SetupCredentials(agentCredFilePath, credentialsDoc.XMLPretty())
				if errorSetCred == nil {
					// Attempt to encrypt the credentials file with a fixed key
					EncryptCredentialsFile(agentCredFilePath)
					agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentQMgrAuthenticationCredentialsFile", agentCredFilePath)
				} else {
					utils.PrintLog(errorSetCred.Error())
				}

				if logLevel >= LOG_LEVEL_VERBOSE && len(agentConfig) > 0 {
					utils.PrintLog(fmt.Sprintf("Updated agent configuration - %v", agentConfig))
				}

				// Update UserSandbox XML file - valid only for STANDARD agents
				if standardAgent {
					errCusbox := CreateUserSandbox(bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQMgr + MFT_AGENTS_SLASH + agentName + MFT_USER_SANDBOX_SLASH)
					if errCusbox != nil {
						utils.PrintLog(errCusbox.Error())
						created = false
					}
				} else {
					// This is a bridge agent. We need to update the ProtocolBridgeProperties.xml file for all other servers specified
					// in configuration JSON file.
					created = updateProtocolBridgePropertiesFile(protocolBridgePropertiesFile, agentConfig)
				}
			}

			if created {
				// Update agent properties file with additional attributes specified.
				created = UpdateAgentProperties(agentPropertiesFile, agentConfig, "additionalProperties", !standardAgent)
				if created {
					// Tell user that agent has been configured.
					utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_CREATED_0047, agentName))
				}
			}
		}
	}
	return created
}

func configTLSAgent(agentConfig string, credentialsDoc *xmldom.Document, agentCredFilePath string) (bool, string) {
	var created bool = true

	// Create keystore using certificate provided if available.
	cipherName, cipherSet := os.LookupEnv(MFT_AGENT_QMGR_CIPHER)
	if cipherSet && len(strings.Trim(cipherName, TEXT_TRIM)) > 0 {
		password := generateRandomPassword()
		publicKeyFile := getKeyFile(agentQMCertPath, ".crt")
		if len(publicKeyFile) > 0 {
			agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslCipherSpec", cipherName)
			// Update coordination properties file
			errCreateKeyStore := CreateKeyStore(KEYSTORES_PATH, AGENT_QM_TRUSTSTORE, publicKeyFile, password)
			if errCreateKeyStore == nil {
				agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslCipherSpec", cipherName)
				agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslTrustStore", filepath.Join(KEYSTORES_PATH, AGENT_QM_TRUSTSTORE))
				agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslTrustStoreType", "pkcs12")
				agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslTrustStoreCredentialsFile", agentCredFilePath)
				UpdateXmlWithKeyStoreCredentials(credentialsDoc, filepath.Join(KEYSTORES_PATH, AGENT_QM_TRUSTSTORE), password)
			} else {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_KEYSTORE_CREATE_FAILED, AGENT_QM_TRUSTSTORE, errCreateKeyStore.Error()))
				created = false
			}
		}

		// Private key
		privateKeyCertPath := getKeyFile(agentQMCertPath, ".key")
		if len(privateKeyCertPath) > 0 {
			errCreateSslStore := CreateKeyStore(KEYSTORES_PATH, AGENT_QM_KEYSTORE, privateKeyCertPath, password)
			if errCreateSslStore == nil {
				agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslKeyStore", filepath.Join(KEYSTORES_PATH, AGENT_QM_KEYSTORE))
				agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslKeyStoreType", "pkcs12")
				agentConfig, _ = sjson.Set(agentConfig, "additionalProperties.agentSslKeyStoreCredentialsFile", agentCredFilePath)
				UpdateXmlWithKeyStoreCredentials(credentialsDoc, filepath.Join(KEYSTORES_PATH, AGENT_QM_KEYSTORE), password)
			} else {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_KEYSTORE_CREATE_FAILED, AGENT_QM_TRUSTSTORE, errCreateSslStore.Error()))
				created = false
			}
			created = true
		}
	}

	return created, agentConfig
}

// Read and process protocol bridge server attributes from configuration JSON file
func updateBridgeParameters(bridgeProperties string, params []string) bool {
	value := false
	var serverType string
	if logLevel >= LOG_LEVEL_VERBOSE {
		utils.PrintLog("Bridge properties: " + bridgeProperties)
	}

	serverName := gjson.Get(bridgeProperties, "name")
	if serverName.Exists() && len(strings.Trim(serverName.String(), TEXT_TRIM)) > 0 {
		value = true
	}

	if value {
		// Set protocol server type
		serverTypeRef := gjson.Get(bridgeProperties, "type")
		if serverTypeRef.Exists() {
			if isBridgeTypeSupported(serverTypeRef.String()) {
				// Use the supplied one.
				params = append(params, "-bt", serverTypeRef.String())
				serverType = serverTypeRef.String()
			} else {
				utils.PrintLog(fmt.Sprintf("%v is not a valid value for Protocol Bridge Type.", serverTypeRef.String()))
				value = false
			}
		} else {
			// Otherwise use default as FTP
			params = append(params, "-bt", "FTP")
			serverType = "FTP"
		}
	}

	if value {
		// Set the protocol server host
		serverHost := gjson.Get(bridgeProperties, "host")
		if serverHost.Exists() {
			params = append(params, "-bh", serverHost.String())
		} else {
			// Use local host if host name is not specified.
			params = append(params, "-bh", "localhost")
		}
	}

	if value {
		// Set the protocol server timezone, valid only for FTP and FTPS server
		if serverType != "SFTP" {
			serverTimezone := gjson.Get(bridgeProperties, "timeZone")
			if serverTimezone.Exists() {
				params = append(params, "-btz", serverTimezone.String())
			} else {
				// TimeZone not specified return an error
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_BRIDGE_PROPERTY_NOT_SET, "timeZone", serverName))
				value = false
			}
		}
	}

	if value {
		// Set the protocol server platform.
		serverPlatform := gjson.Get(bridgeProperties, "platform")
		if serverPlatform.Exists() {
			if isValidBridgePlatform(serverPlatform.String()) {
				params = append(params, "-bm", serverPlatform.String())
			} else {
				utils.PrintLog(fmt.Sprintf("%v is not a valid value for Protocol Bridge platform.", serverPlatform.String()))
				value = false
			}
		}
	}

	if value {
		// Set the protocol server locale
		if serverType != "SFTP" {
			serverLocale := gjson.Get(bridgeProperties, "locale")
			if serverLocale.Exists() {
				params = append(params, "-bsl", serverLocale.String())
			} else {
				// Mandatory property not set
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_BRIDGE_PROPERTY_NOT_SET, "locale", serverName))
				value = false
			}
		}
	}

	if value {
		// Set protocol server file encoding
		serverFileEncoding := gjson.Get(bridgeProperties, "fileEncoding")
		if serverFileEncoding.Exists() {
			params = append(params, "-bfe", serverFileEncoding.String())
		} else {
			// Mandatory property not set
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_BRIDGE_PROPERTY_NOT_SET, "fileEncoding", serverName))
			value = false
		}
	}

	if value {
		// Set the protocol server port
		serverPort := gjson.Get(bridgeProperties, "port")
		if serverPort.Exists() {
			params = append(params, "-bp", serverPort.String())
		}
	}

	if value {
		// Set the protocol server trust store file
		serverTrustStoreFile := gjson.Get(bridgeProperties, "trustStoreFile")
		if serverTrustStoreFile.Exists() {
			params = append(params, "-bts", serverTrustStoreFile.String())
		}
	}

	if value {
		// Set protocol server limited write flag
		serverLimitedWrite := gjson.Get(bridgeProperties, "limitedWrite")
		if serverLimitedWrite.Exists() {
			if strings.EqualFold(serverLimitedWrite.String(), "true") {
				params = append(params, "-blw")
			}
		}
	}

	if value {
		// Set the protocol server list format.
		serverListFormat := gjson.Get(bridgeProperties, "listFormat")
		if serverListFormat.Exists() {
			if isValidBridgeListFormat(serverListFormat.String()) {
				params = append(params, "-blf", serverListFormat.String())
			} else {
				utils.PrintLog(fmt.Sprintf("%v is not a valid Protocol Server List Format", serverListFormat.String()))
				value = false
			}
		}
	}

	return value
}

func isValidBridgeListFormat(platform string) bool {
	var retVal bool
	if strings.EqualFold(platform, "UNIX") ||
		strings.EqualFold(platform, "WINDOWS") ||
		strings.EqualFold(platform, "OS400IFS") {
		retVal = true
	}
	return retVal
}

func isValidBridgePlatform(platform string) bool {
	var retVal bool
	if strings.EqualFold(platform, "UNIX") ||
		strings.EqualFold(platform, "WINDOWS") ||
		strings.EqualFold(platform, "OS400") {
		retVal = true
	}
	return retVal
}

func isBridgeTypeSupported(bridgeType string) bool {
	var retVal bool
	if strings.EqualFold(bridgeType, "SFTP") ||
		strings.EqualFold(bridgeType, "FTPS") ||
		strings.EqualFold(bridgeType, "FTP") {
		retVal = true
	}
	return retVal
}

/*
* Validate agent attributes
 */
func ValidateAgentAttributes(jsonData string) error {
	// Agent name is mandatory
	if !gjson.Get(jsonData, "name").Exists() {
		err := errors.New(utils.MFT_CONT_CFG_AGENT_NAME_MISSING_0020)
		return err
	}

	// Agent queue manager name is mandatory
	if !gjson.Get(jsonData, "qmgrName").Exists() {
		err := errors.New(utils.MFT_CONT_CFG_AGENT_QM_NAME_MISSING_0021)
		return err
	}

	// Agent queue manager host is mandatory
	if !gjson.Get(jsonData, "qmgrHost").Exists() {
		err := errors.New(utils.MFT_CONT_CFG_AGENT_QM_HOST_MISSING_0022)
		return err
	}

	return nil
}

// Update agent.properties file with any additional properties specified in
// configuration JSON file.
func UpdateAgentProperties(propertiesFile string, agentConfig string, sectionName string, bridgeAgent bool) bool {
	f, err := os.OpenFile(propertiesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_OPN_FILE_0067, propertiesFile, err))
		return false
	}
	defer f.Close()

	// Enable logCapture by default. Customer can turn off by specifying it again in config map
	if _, err := f.WriteString("logCapture=true\n"); err != nil {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
	}

	// Set maximum restart count to 0, so that the container ends immediately
	// if the first attempt fails.
	if _, err := f.WriteString("maxRestartCount=0\n"); err != nil {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
	}

	if gjson.Get(agentConfig, sectionName).Exists() {
		result := gjson.Get(agentConfig, sectionName)
		result.ForEach(func(key, value gjson.Result) bool {
			if _, err := f.WriteString(key.String() + "=" + value.String() + "\n"); err != nil {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
			}
			return true // keep iterating
		})
	}

	retVal := false
	// If this is a bridge agent, then configure custom exit
	if bridgeAgent {
		if _, err := f.WriteString("protocolBridgeCredentialExitClasses=com.ibm.wmq.bridgecredentialexit.ProtocolBridgeCustomCredentialExit\n"); err != nil {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
		} else {
			// enableQueueInputOutput property is not valid for bridge agent
			if _, err := f.WriteString("enableQueueInputOutput=false\n"); err != nil {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
			} else {
				retVal = true
			}
		}
	} else {
		if _, err := f.WriteString("userSandboxes=true"); err != nil {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
		} else {
			retVal = true
		}
	}
	return retVal
}

// Updates ProtocolBridgeProperties file with specified additional attributes
func updateProtocolBridgePropertiesFile(propertiesFile string, agentConfig string) bool {
	// First read the entire contents of the ProtocolBridgeProperties file and build a xml file
	bridgeProperitesXml := readFileContents(propertiesFile)
	if logLevel >= LOG_LEVEL_VERBOSE {
		utils.PrintLog(bridgeProperitesXml)
	}
	protocolBridgeConfigs := gjson.Get(agentConfig, "protocolServers").Array()

	// Create a new protocol bridge properties xml document
	bridgePropetiesDoc := xmldom.NewDocument("tns:serverProperties")
	bridgePropetiesDoc.Root.SetAttributeValue("xmlns:tns", "http://wmqfte.ibm.com/ProtocolBridgeProperties")
	bridgePropetiesDoc.Root.SetAttributeValue("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	bridgePropetiesDoc.Root.SetAttributeValue("xsi:schemaLocation", "http://wmqfte.ibm.com/ProtocolBridgeProperties ProtocolBridgeProperties.xsd")

	// Max destination transfers - global
	if gjson.Get(agentConfig, "maxActiveDestinationTransfers").Exists() {
		maxTransfers := gjson.Get(agentConfig, "maxActiveDestinationTransfers").Int()
		node := bridgePropetiesDoc.Root.CreateNode("tns:maxActiveDestinationTransfers")
		node.SetAttributeValue("value", strconv.Itoa(int(maxTransfers)))
	}

	// failTransferWhenCapacityReached - global
	if gjson.Get(agentConfig, "failTransferWhenCapacityReached").Exists() {
		failTransfer := gjson.Get(agentConfig, "failTransferWhenCapacityReached").Bool()
		node := bridgePropetiesDoc.Root.CreateNode("tns:failTransferWhenCapacityReached")
		node.SetAttributeValue("value", strconv.FormatBool(failTransfer))
	}

	// setup default server if one is provided.
	if gjson.Get(agentConfig, "defaultServer").Exists() {
		defaultServerName := gjson.Get(agentConfig, "defaultServer").String()
		defaultServerNode := bridgePropetiesDoc.Root.CreateNode("tns:defaultServer")
		defaultServerNode.SetAttributeValue("name", defaultServerName)
	}

	for index := 0; index < len(protocolBridgeConfigs); index++ {
		serverJson := protocolBridgeConfigs[index]
		updateServer(bridgePropetiesDoc, serverJson.String())
	}

	if logLevel >= LOG_LEVEL_VERBOSE {
		// Write the updated contents to file
		utils.PrintLog(fmt.Sprintf("%v", bridgePropetiesDoc.XMLPrettyEx("  ")))
	}

	// Write the updated xml to file.
	errWd := utils.WriteData(propertiesFile, bridgePropetiesDoc.XMLPrettyEx("  "))
	if errWd != nil {
		utils.PrintLog(errWd.Error())
		return false
	}
	return true
}

// Update ProtocolBridgeProperties.xml file
func updateServer(bridgePropetiesDoc *xmldom.Document, serverJson string) {
	// Check if server definition already exists in Xml file.
	if gjson.Get(serverJson, "name").Exists() && gjson.Get(serverJson, "type").Exists() {
		serverType := gjson.Get(serverJson, "type").String()
		serverName := gjson.Get(serverJson, "name").String()
		if strings.EqualFold(serverType, "FTP") {
			server := bridgePropetiesDoc.Root.QueryOne("//tns:ftpServer[@name='" + serverName + "']")
			if server != nil {
				// Found matching server. Now update the Xml
				updateFTPServerAttributes(server, serverJson)
			} else {
				// FTP server does not exist. Create a new enty
				server := bridgePropetiesDoc.Root.CreateNode("tns:ftpServer")
				server.SetAttributeValue("name", serverName)
				updateFTPServerAttributes(server, serverJson)
			}
		} else if strings.EqualFold(serverType, "SFTP") {
			server := bridgePropetiesDoc.Root.QueryOne("//tns:sftpServer[@name='" + serverName + "']")
			if server != nil {
				// Found matching server. Now update the Xml
				updateSFTPServerAttributes(server, serverJson)
			} else {
				// FTP server does not exist. Create a new enty
				server := bridgePropetiesDoc.Root.CreateNode("tns:sftpServer")
				server.SetAttributeValue("name", serverName)
				updateSFTPServerAttributes(server, serverJson)
			}
		} else if strings.EqualFold(serverType, "FTPS") {
			utils.PrintLog("FTPS protocol servers not supported in this version.")
		}
	} else {
		// log an informational message
		if logLevel >= LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf(utils.MFT_PBA_HOST_AND_TYPE_NOT_FOUND, jsonAgentConfigFilePath))
		}
	}
}

// Update FTP specific attributes
func updateFTPServerAttributes(serverNode *xmldom.Node, serverJson string) {
	if gjson.Get(serverJson, "host").Exists() {
		value := gjson.Get(serverJson, "host").String()
		serverNode.SetAttributeValue("host", value)
	}

	if gjson.Get(serverJson, "port").Exists() {
		value := gjson.Get(serverJson, "port").Int()
		serverNode.SetAttributeValue("port", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "platform").Exists() {
		value := gjson.Get(serverJson, "platform").String()
		serverNode.SetAttributeValue("platform", value)
	}

	if gjson.Get(serverJson, "timeZone").Exists() {
		value := gjson.Get(serverJson, "timeZone").String()
		serverNode.SetAttributeValue("timeZone", value)
	}

	if gjson.Get(serverJson, "controlEncoding").Exists() {
		value := gjson.Get(serverJson, "controlEncoding").String()
		serverNode.SetAttributeValue("controlEncoding", value)
	}

	if gjson.Get(serverJson, "locale").Exists() {
		value := gjson.Get(serverJson, "locale").String()
		serverNode.SetAttributeValue("locale", value)
	}

	if gjson.Get(serverJson, "fileEncoding").Exists() {
		value := gjson.Get(serverJson, "fileEncoding").String()
		serverNode.SetAttributeValue("fileEncoding", value)
	}

	if gjson.Get(serverJson, "listFormat").Exists() {
		value := gjson.Get(serverJson, "listFormat").String()
		serverNode.SetAttributeValue("listFormat", value)
	}

	if gjson.Get(serverJson, "listFileRecentDateFormat").Exists() {
		value := gjson.Get(serverJson, "listFileRecentDateFormat").String()
		serverNode.SetAttributeValue("listFileRecentDateFormat", value)
	}

	if gjson.Get(serverJson, "listFileOldDateFormat").Exists() {
		value := gjson.Get(serverJson, "listFileOldDateFormat").String()
		serverNode.SetAttributeValue("listFileOldDateFormat", value)
	}

	if gjson.Get(serverJson, "monthShortNames").Exists() {
		value := gjson.Get(serverJson, "monthShortNames").String()
		serverNode.SetAttributeValue("monthShortNames", value)
	}

	if gjson.Get(serverJson, "limitedWrite").Exists() {
		value := gjson.Get(serverJson, "limitedWrite").Bool()
		serverNode.SetAttributeValue("limitedWrite", strconv.FormatBool(value))
	}

	if gjson.Get(serverJson, "passiveMode").Exists() {
		value := gjson.Get(serverJson, "passiveMode").Bool()
		serverNode.SetAttributeValue("passiveMode", strconv.FormatBool(value))
	}

	// Create limits section
	limits := serverNode.CreateNode("tns:limits")
	if gjson.Get(serverJson, "maxListFileNames").Exists() {
		value := gjson.Get(serverJson, "maxListFileNames").Int()
		limits.SetAttributeValue("maxListFileNames", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "maxListDirectoryLevels").Exists() {
		value := gjson.Get(serverJson, "maxListDirectoryLevels").Int()
		limits.SetAttributeValue("maxListDirectoryLevels", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "maxSessions").Exists() {
		value := gjson.Get(serverJson, "maxSessions").Int()
		limits.SetAttributeValue("maxSessions", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "socketTimeout").Exists() {
		value := gjson.Get(serverJson, "socketTimeout").Int()
		limits.SetAttributeValue("socketTimeout", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "maxActiveDestinationTransfers").Exists() {
		value := gjson.Get(serverJson, "maxActiveDestinationTransfers").Int()
		limits.SetAttributeValue("maxActiveDestinationTransfers", strconv.Itoa(int(value)))
	}

	/*if gjson.Get(serverJson, "failTransferWhenCapacityReached").Exists() {
		value := gjson.Get(serverJson, "failTransferWhenCapacityReached").Bool()
		limits.SetAttributeValue("failTransferWhenCapacityReached", strconv.FormatBool(value))
	}*/
}

// Update SFTP specific attributes
func updateSFTPServerAttributes(serverNode *xmldom.Node, serverJson string) {
	if gjson.Get(serverJson, "host").Exists() {
		value := gjson.Get(serverJson, "host").String()
		serverNode.SetAttributeValue("host", value)
	}

	if gjson.Get(serverJson, "port").Exists() {
		value := gjson.Get(serverJson, "port").Int()
		serverNode.SetAttributeValue("port", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "platform").Exists() {
		value := gjson.Get(serverJson, "platform").String()
		serverNode.SetAttributeValue("platform", value)
	}

	if gjson.Get(serverJson, "controlEncoding").Exists() {
		value := gjson.Get(serverJson, "controlEncoding").String()
		serverNode.SetAttributeValue("controlEncoding", value)
	}

	if gjson.Get(serverJson, "fileEncoding").Exists() {
		value := gjson.Get(serverJson, "fileEncoding").String()
		serverNode.SetAttributeValue("fileEncoding", value)
	}

	if gjson.Get(serverJson, "limitedWrite").Exists() {
		value := gjson.Get(serverJson, "limitedWrite").Bool()
		serverNode.SetAttributeValue("limitedWrite", strconv.FormatBool(value))
	}

	// Create limits section
	limits := serverNode.CreateNode("tns:limits")
	if gjson.Get(serverJson, "maxListFileNames").Exists() {
		value := gjson.Get(serverJson, "maxListFileNames").Int()
		limits.SetAttributeValue("maxListFileNames", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "maxListDirectoryLevels").Exists() {
		value := gjson.Get(serverJson, "maxListDirectoryLevels").Int()
		limits.SetAttributeValue("maxListDirectoryLevels", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "maxSessions").Exists() {
		value := gjson.Get(serverJson, "maxSessions").Int()
		limits.SetAttributeValue("maxSessions", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "socketTimeout").Exists() {
		value := gjson.Get(serverJson, "socketTimeout").Int()
		limits.SetAttributeValue("socketTimeout", strconv.Itoa(int(value)))
	}

	if gjson.Get(serverJson, "maxActiveDestinationTransfers").Exists() {
		value := gjson.Get(serverJson, "maxActiveDestinationTransfers").Int()
		limits.SetAttributeValue("maxActiveDestinationTransfers", strconv.Itoa(int(value)))
	}

	/*if gjson.Get(serverJson, "failTransferWhenCapacityReached").Exists() {
		value := gjson.Get(serverJson, "failTransferWhenCapacityReached").Bool()
		limits.SetAttributeValue("failTransferWhenCapacityReached", strconv.FormatBool(value))
	}*/
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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_CLN_0048, cleanItem))
		}
	}
}

// Unregister and delete agent
func deleteAgent(coordinationQMgr string, agentName string) error {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_DLTNG_0049, agentName))

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
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_DLTED_0050, agentName))
	}
	return nil
}

// Clean agent on start of container.
func cleanAgentItem(coordinationQMgr string, agentName string, item string, option string) error {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_CLN_0051, item, agentName))

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
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
	} else {
		if logLevel >= LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, outb.String()))
		}
		if item == "all" {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_ALL_ITEM_CLN_0076, agentName))
		} else {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_ITEM_CLN_0052, item, agentName))
		}
	}
	return nil
}

// Create resource monitor
func createResourceMonitor(coordinationQMgr string, agentName string, agentQMgr string,
	monitorName string, fileName string) error {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_RM_CRT_0053, monitorName))

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
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
		return nil
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, outb.String()))
	}
	return nil
}

// Returns the contents of the specified file.
func readFileContents(propertiesFile string) string {
	// Open our xmlFile
	bridgeProperiesXmlFile, err := os.Open(propertiesFile)
	// if we os.Open returns an error then handle it
	if err != nil {
		if logLevel >= LOG_LEVEL_VERBOSE {
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

/*
*
Ping the agent to determine if it's ready to process transfer requests
*/
func PingAgent(coordinationQMgr string, agentName string, waitTime string) bool {
	var outb, errb bytes.Buffer
	retVal := false

	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_AGNT_VRFY_STATUS_0044, agentName))
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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_INFO_0043, outb.String()))
			}
			// The output must contain BFGCL0793I. Ideally we should check
			// for return code of 0 from command execution. Need to figure
			// out a way to do that in Go.
			if strings.Contains(outb.String(), "BFGCL0793I:") == true {
				retVal = true
			}
		}
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}
	return retVal
}

func getMonitorXml(monitorConfig string) string {
	startXml := "<?xml version=\"1.0\" encoding=\"UTF-8\"?><monitor:monitor version=\"6.00\"" +
		" xmlns:monitor=\"http://www.ibm.com/xmlns/wmqfte/7.0.1/MonitorDefinition\"" +
		" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"" +
		" xsi:schemaLocation=\"http://www.ibm.com/xmlns/wmqfte/7.0.1/MonitorDefinition ./Monitor.xsd\">"
	endXml := "</monitor:monitor>"
	finalXml := ""

	// Add originator
	originator := ""
	finalXml = startXml + originator + endXml
	return finalXml
}

func getOriginatorXml(monitorConfig string) string {
	originatorXmlBegin := "<originator>"
	originatorXmlEnd := "</originator>"
	osHostname, err := os.Hostname()
	hostName := "<hostName>"
	if err == nil {
		hostName = "<hostName>" + osHostname + "</hostName>"
	} else {
		hostName = "<hostName>" + "</hostName>"
	}

	curUser, err := user.Current()
	userId := "<userID>"
	if err == nil {
		userId += curUser.Username
	}
	userId += "</userID>"
	return originatorXmlBegin + hostName + userId + originatorXmlEnd
}
