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

// Setup coordination configuration for agent.
func setupCoordination(allAgentConfig string, bfgDataPath string, agentNameEnv string) bool {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
	coordinationQueueManagerName := gjson.Get(allAgentConfig, "coordinationQMgr.name").String()
	// Setup coordination configuration
	utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_CORD_CONFIG_MSG_0024, agentNameEnv, coordinationQueueManagerName))

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
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
		} else {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf("Command output: %s", outb.String()))
			}

			// Update coordination properties file with additional attributes specified.
			coordinationPropertiesFile := bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQueueManagerName + MFT_CORD_PROPS_SLASH
			// Coordination queue manager credentials file
			var coordCredFilePath string = bfgDataPath + MFT_CONFIG_PATH_SUFFIX + coordinationQueueManagerName + MFT_CORD_CRED_SLASH

			// Start XML document for credentials file
			credentialsDoc := InitializeCredentialsDocumentWriter()

			// Configure TLS security
			created, allAgentConfig = configTLSCoordination(allAgentConfig, credentialsDoc, coordCredFilePath)

			if created {
				// If a credentials file has been specified as environment variable, then set it here
				if gjson.Get(allAgentConfig, "coordinationQMgr.qmgrCredentials").Exists() {
					// Write coordination queue manager credentials
					UpdateXmlWithQmgrCredentials(credentialsDoc, gjson.Get(allAgentConfig, "coordinationQMgr.qmgrCredentials").String(), coordinationQueueManagerName)
				}

				errSetCred := setupCredentials(coordCredFilePath, credentialsDoc.XMLPretty())
				if errSetCred == nil {
					// Attempt to encrypt the credentials file with a fixed key
					EncryptCredentialsFile(coordCredFilePath)
					allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationQMgrAuthenticationCredentialsFile", coordCredFilePath)
				} else {
					utils.PrintLog(errSetCred.Error())
				}

				if logLevel >= LOG_LEVEL_VERBOSE && len(allAgentConfig) > 0 {
					utils.PrintLog(fmt.Sprintf(utils.MFT_UPDATED_CONFIGURATION, allAgentConfig))
				}

				// Update coordination properties file
				err := UpdateProperties(coordinationPropertiesFile, allAgentConfig, "coordinationQMgr.additionalProperties")
				if err != nil {
					utils.PrintLog(err.Error())
				} else {
					if logLevel >= LOG_LEVEL_VERBOSE && len(coordCredFilePath) > 0 {
						utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CFG_CORD_CONFIG_CRED_PATH_0027, coordCredFilePath))
					}
					utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CORD_SETUP_COMP_0054, coordinationQueueManagerName))
					created = true
				}
			}
		}
	} else {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_NOT_FOUND_0028, lookPathErr))
	}

	return created
}

// Create keystore using certificate provided if available. We need cipher name
// at least public key environment variable to be set.
func configTLSCoordination(allAgentConfig string, credentialsDoc *xmldom.Document, coordCredFilePath string) (bool, string) {
	var created bool

	cipherName, cipherSet := os.LookupEnv(MFT_COORD_QMGR_CIPHER)
	if cipherSet && len(strings.Trim(cipherName, TEXT_TRIM)) > 0 {
		// Generate password for keystore
		password := generateRandomPassword()
		// See if we have public key file
		publicKeyCertPath := getKeyFile(coordinationQMCertPath, ".crt")
		if len(publicKeyCertPath) > 0 {
			// Trust Keystore details - public key of queue manager
			errCreateKeyStore := CreateKeyStore(KEYSTORES_PATH, COORD_QM_TRUSTSTORE, publicKeyCertPath, password)
			if errCreateKeyStore == nil {
				// Update coordination properties file
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslCipherSpec", cipherName)
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslTrustStore", filepath.Join(KEYSTORES_PATH, COORD_QM_TRUSTSTORE))
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslTrustStoreType", "pkcs12")
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslTrustStoreCredentialsFile", coordCredFilePath)
				UpdateXmlWithKeyStoreCredentials(credentialsDoc, filepath.Join(KEYSTORES_PATH, COORD_QM_TRUSTSTORE), password)
				created = true
			} else {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_KEYSTORE_CREATE_FAILED, COORD_QM_TRUSTSTORE, errCreateKeyStore.Error()))
				created = false
			}
		}

		// Do we have any private key
		privateKeyCertPath := getKeyFile(coordinationQMCertPath, ".key")
		if len(privateKeyCertPath) > 0 {
			errCreateSslStore := CreateKeyStore(KEYSTORES_PATH, COORD_QM_KEYSTORE, privateKeyCertPath, password)
			if errCreateSslStore == nil {
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslCipherSpec", cipherName)
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslKeyStore", filepath.Join(KEYSTORES_PATH, COORD_QM_KEYSTORE))
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslKeyStoreType", "pkcs12")
				allAgentConfig, _ = sjson.Set(allAgentConfig, "coordinationQMgr.additionalProperties.coordinationSslKeyStoreCredentialsFile", coordCredFilePath)
				UpdateXmlWithKeyStoreCredentials(credentialsDoc, filepath.Join(KEYSTORES_PATH, COORD_QM_KEYSTORE), password)
				created = true
			} else {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_KEYSTORE_CREATE_FAILED, COORD_QM_TRUSTSTORE, errCreateSslStore.Error()))
				created = false
			}
		} else {
			utils.PrintLog(utils.MFT_CONT_MTLS_NOT_CONFIGURED)
			created = true
		}
	} else {
		// TLS not enabled.
		created = true
	}
	return created, allAgentConfig
}

// Validate attributes in JSON file.
// Check if the configuration JSON contains all required attribtes
func validateCoordinationAttributes(jsonData string) error {
	// Coordination queue manager is mandatory
	if !gjson.Get(jsonData, "coordinationQMgr.name").Exists() {
		err := errors.New(utils.MFT_CONT_CFG_CORD_QM_NAME_MISSING_0014)
		return err
	}

	// Coordination queue manager host is mandatory
	if !gjson.Get(jsonData, "coordinationQMgr.host").Exists() {
		err := errors.New(utils.MFT_CONT_CFG_CORD_QM_HOST_MISSING_0015)
		return err
	}

	return nil
}
