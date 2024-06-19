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
	"path/filepath"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/subchen/go-xmldom"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Calls fteSetupCommands to create command queue manager configuration.
func setupCommands(allAgentConfig string, bfgDataPath string, agentName string) bool {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
	commandQueueManager := gjson.Get(allAgentConfig, "commandQMgr.name").String()

	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_SETUP_STRT_0055, agentName, commandQueueManager))

	// Get the path of MFT fteSetupCommands command.
	cmdCmdsPath, lookPathErr := exec.LookPath("fteSetupCommands")
	if lookPathErr == nil {
		// Setup commands configuration
		if !gjson.Get(allAgentConfig, "commandQMgr.name").Exists() {
			utils.PrintLog("Command queue manager name not provided")
			return false
		}
		var port string
		var channel string
		var hostName string
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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
			os.Exit(1)
		} else {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_INFO_0043, outb.String()))
			}

			coordinationQmgrName := gjson.Get(allAgentConfig, "coordinationQMgr.name").String()
			// Start XML document for credentials file
			credentialsDoc := InitializeCredentialsDocumentWriter()
			cmdCredFilePath := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQmgrName + MFT_CMD_CRED_SLASH
			// Configure TLS for command queue manager
			created, allAgentConfig = configTLSCommand(allAgentConfig, credentialsDoc, cmdCredFilePath)

			if gjson.Get(allAgentConfig, "commandQMgr.qmgrCredentials").Exists() {
				// Write coordination queue manager credentials
				UpdateXmlWithQmgrCredentials(credentialsDoc, gjson.Get(allAgentConfig, "commandQMgr.qmgrCredentials").String(), commandQueueManager)
			}

			errSetCred := setupCredentials(cmdCredFilePath, credentialsDoc.XMLPretty())
			if errSetCred == nil {
				// Attempt to encrypt the credentials file with a fixed key
				EncryptCredentialsFile(cmdCredFilePath)
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionQMgrAuthenticationCredentialsFile", cmdCredFilePath)
			} else {
				utils.PrintLog(errSetCred.Error())
			}

			if logLevel >= LOG_LEVEL_VERBOSE && len(cmdCredFilePath) > 0 {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_QMGR_CRED_PATH_0056, cmdCredFilePath))
			}

			if logLevel >= LOG_LEVEL_VERBOSE && len(allAgentConfig) > 0 {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_UPDATED_CMD_CONFIG, allAgentConfig))
			}

			// Update command properties file with additional attributes specified.
			commandsPropertiesFile := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQmgrName + MFT_CMD_PROPS_SLASH
			err := UpdateProperties(commandsPropertiesFile, allAgentConfig, "commandQMgr.additionalProperties")
			if err != nil {
				utils.PrintLog(err.Error())
			} else {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_SETUP_COMP_0057, commandQueueManager))
				created = true
			}
		}
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}

	return created
}

// Configure TLS for command queue manager
func configTLSCommand(allAgentConfig string, credentialsDoc *xmldom.Document, cmdCredFilePath string) (bool, string) {
	var created bool
	// Create keystore using certificate provided if available.
	cipherName, cipherSet := os.LookupEnv(MFT_CMD_QMGR_CIPHER)
	if cipherSet && len(strings.Trim(cipherName, TEXT_TRIM)) > 0 {
		password := generateRandomPassword()
		// Search for .crt file in the predefined directory
		publicKeyCertPath := getKeyFile(commandQMCertPath, ".crt")
		if len(publicKeyCertPath) > 0 {
			// Trust store - public key of command queue manager
			errCreateKeyStore := CreateKeyStore(KEYSTORES_PATH, CMD_QM_TRUSTSTORE, publicKeyCertPath, password)
			if errCreateKeyStore == nil {
				// Update coordination properties file
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslCipherSpec", cipherName)
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslTrustStore", filepath.Join(KEYSTORES_PATH, CMD_QM_TRUSTSTORE))
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslTrustStoreType", "pkcs12")
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslTrustStoreCredentialsFile", cmdCredFilePath)
				UpdateXmlWithKeyStoreCredentials(credentialsDoc, filepath.Join(KEYSTORES_PATH, CMD_QM_TRUSTSTORE), password)
				created = true
			} else {
				utils.PrintLog(errCreateKeyStore.Error())
			}
		}

		// Key store details - private key
		privateKeyCertPath := getKeyFile(commandQMCertPath, ".key")
		if len(privateKeyCertPath) > 0 {
			errCreateSslStore := CreateKeyStore(KEYSTORES_PATH, CMD_QM_KEYSTORE, privateKeyCertPath, password)
			if errCreateSslStore == nil {
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslCipherSpec", cipherName)
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslKeyStore", filepath.Join(KEYSTORES_PATH, CMD_QM_KEYSTORE))
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslKeyStoreType", "pkcs12")
				allAgentConfig, _ = sjson.Set(allAgentConfig, "commandQMgr.additionalProperties.connectionSslKeyStoreCredentialsFile", cmdCredFilePath)
				UpdateXmlWithKeyStoreCredentials(credentialsDoc, filepath.Join(KEYSTORES_PATH, CMD_QM_KEYSTORE), password)
				created = true
			} else {
				utils.PrintLog(errCreateSslStore.Error())
				created = false
			}
		}
	} else {
		created = true
	}

	return created, allAgentConfig
}

// Validate the configuration for required command qmgr attributes.
func validateCommandAttributes(jsonData string) error {
	// Commands queue manager is mandatory
	if !gjson.Get(jsonData, "commandQMgr.name").Exists() {
		err := errors.New(utils.MFT_CONT_CFG_CMD_QM_NAME_MISSING_0017)
		return err
	}
	// Coordination queue manager host is mandatory
	if !gjson.Get(jsonData, "commandQMgr.host").Exists() {
		err := errors.New(utils.MFT_CONT_CFG_CMD_QM_HOST_MISSING_0018)
		return err
	}

	return nil
}
