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

/**
* This file contains utility methods for agent, coordination and commnad
* configuration.
 */
import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

// Update coordination and command properties file with any additional properties specified in
// configuration JSON file.
func UpdateProperties(propertiesFile string, agentConfig string, sectionName string,
	credentialPropName string, credentialFileName string) {
	f, err := os.OpenFile(propertiesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_OPN_FILE_0067, propertiesFile, err))
		return
	}
	defer f.Close()

	// Iterate throw the given attributes and updathe the specified properties file
	if gjson.Get(agentConfig, sectionName).Exists() {
		// Write a new line character before updating properties
		result := gjson.Get(agentConfig, sectionName)
		if _, err := f.WriteString("\n"); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
			return
		}
		result.ForEach(func(key, value gjson.Result) bool {
			if !strings.EqualFold(strings.ToUpper(credentialPropName), strings.ToUpper(key.String())) {
				if _, err := f.WriteString(key.String() + "=" + value.String() + "\n"); err != nil {
					utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
					return false
				}
			}
			return true // keep iterating
		})
	}

	// Update credentials property
	if credentialPropName != TEXT_BLANK && credentialFileName != TEXT_BLANK {
		if _, err := f.WriteString(credentialPropName + "=" + credentialFileName + "\n"); err != nil {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
		}
	}
}

// Check if command tracing is enabled. Command tracing can be enabled
// by specifying MFT_TRACE_COMMANDs environment variable.
func IsCommandTracingEnabled() bool {
	var commandTraceEnabled bool = false
	enableCommandTrace, enableCommandTraceSet := os.LookupEnv(MFT_TRACE_COMMANDS)
	commandTracePath, commandTracePathSet := os.LookupEnv(MFT_TRACE_COMMAND_PATH)
	if enableCommandTraceSet && commandTracePathSet {
		enableCommandTrace = strings.ToLower(strings.Trim(enableCommandTrace, TEXT_TRIM))
		if strings.EqualFold(enableCommandTrace, TEXT_YES) {
			if commandTracePath != TEXT_BLANK {
				commandTraceEnabled = true
			}
		}
	}
	return commandTraceEnabled
}

// Return command trace path specified as a value MFT_TRACE_COMMAND
// environment variable
func GetCommandTracePath() string {
	commandTracePath, commandTracePathSet := os.LookupEnv(MFT_TRACE_COMMAND_PATH)
	if commandTracePathSet && strings.Trim(commandTracePath, TEXT_TRIM) != TEXT_BLANK {
		return strings.Trim(commandTracePath, TEXT_TRIM)
	}
	return TEXT_BLANK
}

// Setup userSandBox configuration to restrict access to file system
func CreateUserSandbox(sandboxXmlFileName string) {
	// Open existing UserSandboxes.xml file
	userSandBoxXmlFile, err := os.OpenFile(sandboxXmlFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	// if we os.Open returns an error then handle it
	if err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_OPN_SNDBOX_FILE_0065, sandboxXmlFileName, err))
		return
	}

	// defer the closing of our xml file so that we can parse it later on
	defer userSandBoxXmlFile.Close()

	// Set sandBoxRoot for standard agents to restrict the file system access. Use the value specified in
	// MFT_MOUNT_PATH environment variable if available else use the default "/mountpath" folder.
	// Agent will be able to read from or write to this folder and it will not have access to other parts
	// of the file system.
	mountPathEnv, mountPathEnvSet := os.LookupEnv(MFT_MOUNT_PATH)
	var mountPath string
	if mountPathEnvSet {
		mountPathEnv = strings.Trim(mountPathEnv, TEXT_TRIM)
		if len(mountPathEnv) > 0 {
			//If the supplied path does not have /** suffix, then add it
			if !strings.HasSuffix(mountPathEnv, "/**") {
				if strings.HasSuffix(mountPathEnv, "/*") {
					mountPath = mountPathEnv + "*"
				} else if strings.HasSuffix(mountPathEnv, "/") {
					mountPath = mountPathEnv + "**"
				} else {
					mountPath = mountPathEnv + "/**"
				}
			} else {
				mountPath = mountPathEnv
			}
		} else {
			// Path provided is blank. So use default fixed path.
			mountPath = DEFAULT_MOUNT_PATH_FOR_TRANSFERS
		}
	} else {
		// No environment variable specified. So use fixed path.
		mountPath = DEFAULT_MOUNT_PATH_FOR_TRANSFERS
	}

	// Write a generic
	var sandboxXmlText string
	sandboxXmlText = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	sandboxXmlText += "<tns:userSandboxes\n"
	sandboxXmlText += "         xmlns:tns=\"http://wmqfte.ibm.com/UserSandboxes\"\n"
	sandboxXmlText += "         xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n"
	sandboxXmlText += "         xsi:schemaLocation=\"http://wmqfte.ibm.com/UserSandboxes UserSandboxes.xsd\">\n\n"
	sandboxXmlText += "    <tns:agent>\n"
	sandboxXmlText += "         <tns:sandbox user=\"^[a-zA-Z0-9]*$\" userPattern=\"regex\">\n"
	sandboxXmlText += "              <tns:read>\n"
	sandboxXmlText += "       	          <tns:include name=\"" + mountPath + "\"/>\n"
	sandboxXmlText += "	                  <tns:include name=\"**\" type=\"queue\"/>\n"
	sandboxXmlText += "              </tns:read>\n"
	sandboxXmlText += "              <tns:write>\n"
	sandboxXmlText += "	                  <tns:include name=\"" + mountPath + "\"/>\n"
	sandboxXmlText += "                   <tns:include name=\"**\" type=\"queue\"/>\n"
	sandboxXmlText += "              </tns:write>\n"
	sandboxXmlText += "        </tns:sandbox>\n"
	sandboxXmlText += "     </tns:agent>\n"
	sandboxXmlText += "</tns:userSandboxes>"

	if logLevel == LOG_LEVEL_DIGANOSTIC {
		utils.PrintLog(sandboxXmlText)
	}

	// Write the updated properties to file.
	_, writeErr := userSandBoxXmlFile.WriteString(sandboxXmlText)
	if writeErr != nil {
		utils.PrintLog(fmt.Sprintf("%v", writeErr))
	}
}

/**
* SetupCredentials
* This method creats specified credentials file which will contain userid and password
* required for connecting to agent queue manager. The userid and password are from
* agent configuration file, like agentconfig.json. This method exepect the userid to
* be in plain text while the password to be base64 encoded.
 */
func SetupCredentials(mqmftCredentialsXmlFileName string, configData string, qmName string) bool {
	// Check if mqUserId attribute exists in JSON file. The password must be base64 encoded
	if gjson.Get(configData, "mqUserId").Exists() && gjson.Get(configData, "mqPassword").Exists() {
		mqUserId := gjson.Get(configData, "mqUserId").String()
		mqPassword := gjson.Get(configData, "mqPassword").String()
		mqUserId = strings.Trim(mqUserId, TEXT_TRIM)
		mqPassword = strings.Trim(mqPassword, TEXT_TRIM)

		if mqUserId != TEXT_BLANK && mqPassword != TEXT_BLANK {
			// Create an empty credentials file, truncate if one exists
			mqmftCredentialsXmlFile, err := os.OpenFile(mqmftCredentialsXmlFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			// if we os.Open returns an error then handle it
			if err != nil {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_OPN_CRED_FILE_0064, mqmftCredentialsXmlFileName, err))
				return false
			}
			defer mqmftCredentialsXmlFile.Close()

			// Decode the password from base64 format
			plainTextPassword, errDecode := Base64Decode(mqPassword)
			if errDecode != nil {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_CRED_DECODE_FAILED_0060, err))
				//MFT_CONT_CRED_NOT_AVAIL_ASM_DFLT_0062)
				plainTextPassword = mqPassword
			} else {
				plainTextPassword = mqPassword
			}

			// Get the current process user id
			user, err := user.Current()
			if err != nil {
				utils.PrintLog(fmt.Sprintf(MFT_CONT_ERR_CONT_USER_0063, err))
				return false
			}
			userId := user.Username

			// Write a credentials
			var credetiXmlText string
			credetiXmlText = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
			credetiXmlText += "<tns:mqmftCredentials\n"
			credetiXmlText += "         xmlns:tns=\"http://wmqfte.ibm.com/MQMFTCredentials\"\n"
			credetiXmlText += "         xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n"
			credetiXmlText += "         xsi:schemaLocation=\"http://wmqfte.ibm.com/MQMFTCredentials MQMFTCredentials.xsd\">\n\n"
			credetiXmlText += "<tns:qmgr name=\""
			credetiXmlText += qmName
			credetiXmlText += "\" user=\""
			credetiXmlText += userId
			credetiXmlText += "\" mqUserId=\""
			credetiXmlText += mqUserId
			credetiXmlText += "\" mqPassword=\""
			credetiXmlText += plainTextPassword
			credetiXmlText += "\"/>\n"
			credetiXmlText += "</tns:mqmftCredentials>"

			// Write the updated properties to file.
			_, writeErr := mqmftCredentialsXmlFile.WriteString(credetiXmlText)
			if writeErr != nil {
				utils.PrintLog(fmt.Sprintf("%v", writeErr))
				return false
			}

			if logLevel == LOG_LEVEL_DIGANOSTIC {
				utils.PrintLog(credetiXmlText)
			}
			return true
		}
	} else {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CRED_NOT_AVAIL_0061, qmName))
	}
	return false
}

/**
* Encrypts specified credentils file using a fixed key. The actual
* file itself is encrypted.
* Returns error if the method fails to encrypt the file.
 */
func EncryptCredentialsFile(credentialsFile string) error {
	var outb, errb bytes.Buffer
	if logLevel == LOG_LEVEL_VERBOSE {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CRED_ENCRYPTING_0058, credentialsFile))
	}

	// Get the path of MFT fteObfuscate command.
	cmdObfuscatePath, lookErr := exec.LookPath("fteObfuscate")
	if lookErr != nil {
		return lookErr
	}

	var cmdArgs []string
	cmdArgs = append(cmdArgs, cmdObfuscatePath, "-f", credentialsFile)
	if commandTracingEnabled {
		cmdArgs = append(cmdArgs, "-trace", "com.ibm.wmqfte=all")
		cmdTracePath := GetCommandTracePath()
		if len(cmdTracePath) > 0 {
			cmdArgs = append(cmdArgs, "-tracePath", cmdTracePath)
		}
	}

	// Encrypt the credentials file with default key
	cmdObfucateCmd := &exec.Cmd{
		Path: cmdObfuscatePath,
		Args: cmdArgs,
	}

	// Reuse the same buffer
	cmdObfucateCmd.Stdout = &outb
	cmdObfucateCmd.Stderr = &errb
	// Execute the fteObfuscate command. Log an error an exit in case of any error.
	if err := cmdObfucateCmd.Run(); err != nil {
		utils.PrintLog(fmt.Sprintf(MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
	} else {
		if logLevel == LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf(MFT_CONT_CRED_ENCRYPTED_0059, credentialsFile))
		}
	}
	return nil
}

/**
* Attempt to decode the password provided. This method expects
* input to be base64 encoded. If the input is not base64 encoded
* the input is returned as is.
 */
func Base64Decode(encodedText string) (string, error) {
	// Must unquote the text before decoding.
	encodedText, err := strconv.Unquote(`"` + encodedText + `"`)
	data, err := base64.StdEncoding.DecodeString(encodedText)
	if err != nil {
		// Failed to decode the input. It may not be base64 encoded
		// Return false to indicate the failure
		return TEXT_BLANK, err
	}

	// Decoding successful.
	return string(data), nil
}

/**
* Process connection security attributes
 */
func ProcessSecureConnectionProperties(secureConfig string) bool {
	// All attributes must be provided.
	if gjson.Get(secureConfig, "cipherSpec").Exists() &&
		gjson.Get(secureConfig, "keyStore").Exists() &&
		gjson.Get(secureConfig, "keyStorePassword").Exists() &&
		gjson.Get(secureConfig, "trustStore").Exists() &&
		gjson.Get(secureConfig, "trustStorePassword").Exists() &&
		gjson.Get(secureConfig, "keyStoreType").Exists() &&
		gjson.Get(secureConfig, "trustStoreType").Exists() {
		//cipherSpec := strings.Trim(gjson.Get(secureConfig, "cipherSpec").String(), TEXT_TRIM)
		//keyStore := strings.Trim(gjson.Get(secureConfig, "keyStore").String(), TEXT_TRIM)
		//keyStorePassword := strings.Trim(gjson.Get(secureConfig, "keyStorePassword").String(), TEXT_TRIM)
		//keyStoreType := strings.Trim(gjson.Get(secureConfig, "keyStoreType").String(), TEXT_TRIM)
		//trustStore := strings.Trim(gjson.Get(secureConfig, "trustStore").String(), TEXT_TRIM)
		//trustStorePassword := strings.Trim(gjson.Get(secureConfig, "trustStorePassword").String(), TEXT_TRIM)
		//trustStoreType := strings.Trim(gjson.Get(secureConfig, "trustStoreType").String(), TEXT_TRIM)
		return true
	}

	return false
}
