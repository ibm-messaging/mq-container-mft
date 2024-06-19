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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/subchen/go-xmldom"
	"github.com/tidwall/gjson"
)

// Update coordination and command properties file with any additional properties specified in
// configuration JSON file.
func UpdateProperties(propertiesFile string, agentConfig string, sectionName string) error {
	f, err := os.OpenFile(propertiesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		errorMsg := fmt.Sprintf(utils.MFT_CONT_ERR_OPN_FILE_0067, propertiesFile, err)
		return errors.New(errorMsg)
	}
	defer f.Close()

	// Iterate throw the given attributes and updathe the specified properties file
	if gjson.Get(agentConfig, sectionName).Exists() {
		// Write a new line character before updating properties
		result := gjson.Get(agentConfig, sectionName)
		if _, err := f.WriteString("\n"); err != nil {
			errorMsg := fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err)
			return errors.New(errorMsg)
		}
		result.ForEach(func(key, value gjson.Result) bool {
			if _, err := f.WriteString(key.String() + "=" + value.String() + "\n"); err != nil {
				utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_ERR_UPDTING_FILE_0066, propertiesFile, err))
				return false // break if an error occurs.
			}
			return true // keep iterating
		})
	}
	return nil
}

// Check if command tracing is enabled. Command tracing can be enabled
// by specifying MFT_TRACE_COMMANDs environment variable.
func IsCommandTracingEnabled() bool {
	var commandTraceEnabled bool = false
	enableCommandTrace, enableCommandTraceSet := os.LookupEnv(MFT_TRACE_COMMANDS)
	if enableCommandTraceSet {
		enableCommandTrace = strings.ToLower(strings.Trim(enableCommandTrace, TEXT_TRIM))
		if strings.EqualFold(enableCommandTrace, TEXT_YES) {
			commandTraceEnabled = true
		}
	}
	return commandTraceEnabled
}

// Return command trace path
func GetCommandTracePath() string {
	commandTracePath, commandTracePathSet := os.LookupEnv(BFG_DATA)
	if commandTracePathSet && strings.Trim(commandTracePath, TEXT_TRIM) != TEXT_BLANK {
		commandTracePath += strings.Trim(commandTracePath, TEXT_TRIM) + "/cmdtrace/"
		// Create command trace path if it does not exist.
		utils.CreatePath(commandTracePath)
		return commandTracePath
	}
	return TEXT_BLANK
}

// Setup userSandBox configuration to restrict access to file system
func createUserSandbox(sandboxXmlFileName string) error {
	var errCusbox error = nil

	// Open existing UserSandboxes.xml file
	userSandBoxXmlFile, err := os.OpenFile(sandboxXmlFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	// if we os.Open returns an error then handle it
	if err != nil {
		errorMsg := fmt.Sprintf(utils.MFT_CONT_ERR_OPN_SNDBOX_FILE_0065, sandboxXmlFileName, err)
		return errors.New(errorMsg)
	}
	// defer the closing of our xml file so that we can parse it later on
	defer userSandBoxXmlFile.Close()

	// Set sandBoxRoot for standard agents to restrict the file system access. Use the value specified in
	// MFT_MOUNT_PATH environment variable if available else use the default "/mountpath" folder.
	// Agent will be able to read from or write to this folder and it will not have access to other parts
	// of the file system.
	var transferRootPath string = DEFAULT_MOUNT_PATH_FOR_TRANSFERS
	mountPathEnv, mountPathEnvSet := os.LookupEnv(MFT_MOUNT_PATH)
	if mountPathEnvSet {
		mountPathEnv = strings.Trim(mountPathEnv, TEXT_TRIM)
		if len(mountPathEnv) > 0 {
			//If the supplied path does not have /** suffix, then add it
			if !strings.HasSuffix(mountPathEnv, "/**") {
				if strings.HasSuffix(mountPathEnv, "/*") {
					transferRootPath = mountPathEnv + "*"
				} else if strings.HasSuffix(mountPathEnv, "/") {
					transferRootPath = mountPathEnv + "**"
				} else {
					transferRootPath = mountPathEnv + "/**"
				}
			} else {
				transferRootPath = mountPathEnv
			}
		}
	}

	sandBoxDoc := xmldom.NewDocument("tns:userSandboxes")
	sandBoxDoc.Root.SetAttributeValue("xmlns:tns", "http://wmqfte.ibm.com/UserSandboxes")
	sandBoxDoc.Root.SetAttributeValue("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	sandBoxDoc.Root.SetAttributeValue("xsi:schemaLocation", "http://wmqfte.ibm.com/UserSandboxes UserSandboxes.xsd")
	agentNode := sandBoxDoc.Root.CreateNode("tns:agent")
	sandBoxNode := agentNode.CreateNode("tns:sandbox")
	// Override sandbox user name and pattern.
	sandBoxNode.SetAttributeValue("user", "^[a-zA-Z0-9]*$")
	sandBoxNode.SetAttributeValue("userPattern", "regex")
	// Read Sandbox Attributes
	readNode := sandBoxNode.CreateNode("tns:read")
	includeNodeRead := readNode.CreateNode("tns:include")
	includeNodeRead.SetAttributeValue("name", transferRootPath)
	includeNodeReadQueue := readNode.CreateNode("tns:include")
	includeNodeReadQueue.SetAttributeValue("name", "**")
	includeNodeReadQueue.SetAttributeValue("type", "queue")
	// Write Sandbox attributes
	writeNode := sandBoxNode.CreateNode("tns:write")
	includeWriteNode := writeNode.CreateNode("tns:include")
	includeWriteNode.SetAttributeValue("name", transferRootPath)
	includeWriteNodeQueue := writeNode.CreateNode("tns:include")
	includeWriteNodeQueue.SetAttributeValue("name", "**")
	includeWriteNodeQueue.SetAttributeValue("type", "queue")

	if logLevel >= LOG_LEVEL_VERBOSE {
		utils.PrintLog(sandBoxDoc.XMLPretty())
	}

	// Write the updated properties to file.
	_, writeErr := userSandBoxXmlFile.Write([]byte(sandBoxDoc.XMLPretty()))
	if writeErr != nil {
		errorMsg := fmt.Sprintf(utils.MFT_FAILED_WRITING_SANDBOX, writeErr)
		errCusbox = errors.New(errorMsg)
	}

	return errCusbox
}

/**
* SetupCredentials
*
* This method creats specified credentials file which will contain userid and password
* required for connecting to agent queue manager. The userid and password are from
* agent configuration file, like agentconfig.json. This method exepect the userid to
* be in plain text while the password to be base64 encoded.
 */
func setupCredentials(mqmftCredentialsXmlFileName string, bufferCred string) error {
	// Create an empty credentials file, truncate if one exists
	mqmftCredentialsXmlFile, err := os.OpenFile(mqmftCredentialsXmlFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	// if we os.Open returns an error then handle it
	if err != nil {
		errorMsg := fmt.Sprintf(utils.MFT_CONT_ERR_OPN_CRED_FILE_0064, mqmftCredentialsXmlFileName, err)
		return errors.New(errorMsg)
	}
	defer mqmftCredentialsXmlFile.Close()

	// Write the updated properties to file.
	_, writeErr := mqmftCredentialsXmlFile.WriteString(bufferCred)
	if writeErr != nil {
		errorMsg := fmt.Sprintf("%v", writeErr)
		return errors.New(errorMsg)
	}

	if logLevel >= LOG_LEVEL_VERBOSE {
		utils.PrintLog(bufferCred)
	}

	return nil
}

/**
* Update XML data with credentials of queue manager
 */
func UpdateXmlWithQmgrCredentials(xmlWriter *xmldom.Document, configData string, qmName string) error {
	var mqUserId string
	var mqPassword string
	var err error
	var currentUser string
	var plainTextPassword string
	var errReturn error = nil

	if gjson.Get(configData, "mqUserId").Exists() &&
		gjson.Get(configData, "mqPassword").Exists() {
		mqUserId = strings.Trim(gjson.Get(configData, "mqUserId").String(), TEXT_TRIM)
		mqPassword = strings.Trim(gjson.Get(configData, "mqPassword").String(), TEXT_TRIM)

		if len(mqUserId) > 0 && len(mqPassword) > 0 {
			// Decode the password from base64 format
			plainTextPassword, err = Base64Decode(mqPassword)
			if err != nil {
				// Password has not been Base64 encoded. Use as it is.
				if logLevel >= LOG_LEVEL_VERBOSE {
					utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CRED_DECODE_FAILED_0060, err))
				}
				plainTextPassword = mqPassword
			} else {
				plainTextPassword = mqPassword
			}

			// Get the current process user id
			user, err := user.Current()
			if err == nil {
				currentUser = user.Username
				childNode := xmlWriter.Root.CreateNode("tns:qmgr")
				childNode.SetAttributeValue("name", qmName)
				childNode.SetAttributeValue("user", currentUser)
				childNode.SetAttributeValue("mqUserId", mqUserId)
				childNode.SetAttributeValue("mqPassword", plainTextPassword)
			} else {
				errorMsg := fmt.Sprintf(utils.MFT_CONT_ERR_CONT_USER_0063, err)
				errReturn = errors.New(errorMsg)
				childNode := xmlWriter.Root.CreateNode("tns:qmgr")
				childNode.SetAttributeValue("name", qmName)
				childNode.SetAttributeValue("user", "unknown")
				childNode.SetAttributeValue("mqUserId", mqUserId)
				childNode.SetAttributeValue("mqPassword", plainTextPassword)
			}
		}
	} else {
		if logLevel >= LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CRED_NOT_AVAIL_0061, qmName))
		}
	}
	return errReturn
}

/**
* Update XML data with credentials of key store/trust store to xml file.
 */
func UpdateXmlWithKeyStoreCredentials(xmlWriter *xmldom.Document, trustStore string, trustStorePassword string) {
	if len(trustStore) > 0 && len(trustStorePassword) > 0 {
		childNode := xmlWriter.Root.CreateNode("tns:file")
		childNode.SetAttributeValue("path", trustStore)
		childNode.SetAttributeValue("password", trustStorePassword)
	}
}

/**
* Initialize credentials file.
 */
func InitializeCredentialsDocumentWriter() *xmldom.Document {
	credentialsDoc := xmldom.NewDocument("tns:mqmftCredentials")
	credentialsDoc.Root.SetAttributeValue("xmlns:tns", "http://wmqfte.ibm.com/MQMFTCredentials")
	credentialsDoc.Root.SetAttributeValue("xmlns:xsi", "http://www.w3.org/2001/XMLSchema-instance")
	credentialsDoc.Root.SetAttributeValue("xsi:schemaLocation", "http://wmqfte.ibm.com/MQMFTCredentials MQMFTCredentials.xsd")
	return credentialsDoc
}

/**
* Encrypts specified credentils file using a fixed key. The actual
* file itself is encrypted.
* Returns error if the method fails to encrypt the file.
 */
func EncryptCredentialsFile(credentialsFile string) error {
	var outb, errb bytes.Buffer
	if logLevel >= LOG_LEVEL_VERBOSE {
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CRED_ENCRYPTING_0058, credentialsFile))
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
		utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CMD_ERROR_0042, outb.String(), errb.String()))
	} else {
		if logLevel >= LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf(utils.MFT_CONT_CRED_ENCRYPTED_0059, credentialsFile))
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
		return true
	}
	return false
}
