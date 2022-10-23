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
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
)

/**
* Create keystore and include user specified certificates
 */
func CreateKeyStore(keyStoreDir string, keyStoreFile string, certFilePath string, certStorePassword string) error {
	// First verify if the certificate file exists
	if !utils.DoesFileExist(certFilePath) {
		errorMsg := "Certificate file " + certFilePath + " does not exist"
		return errors.New(errorMsg)
	}

	keyStorePathFinal := filepath.Join(keyStoreDir, keyStoreFile)
	// Delete if keystore already exists
	finfo, err := os.Lstat(keyStoreDir)
	if err == nil {
		if logLevel >= LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf("Keystore path %v exists.", finfo.Name()))
		}
		certInfo, err := os.Lstat(keyStorePathFinal)
		if err == nil {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf("Keystore %v already exists. Deleting it...", certInfo.Name()))
			}
			delErr := os.Remove(keyStorePathFinal)
			if delErr == nil {
				if logLevel >= LOG_LEVEL_VERBOSE {
					utils.PrintLog(fmt.Sprintf("Existing keystore %v deleted", keyStorePathFinal))
				}
			} else {
				errorMsg := fmt.Sprintf("An error occurred while deleting keystore %v. The error is: %v", keyStorePathFinal, delErr)
				return errors.New(errorMsg)
			}
		}
	}

	// Create directory
	errCreateDataPath := utils.CreatePath(keyStoreDir)
	if errCreateDataPath != nil {
		return errCreateDataPath
	}

	// A new keystore will be created if it does not exist.
	var outb, errb bytes.Buffer
	cmdKeyToolPath, lookPathErr := exec.LookPath("keytool")
	if lookPathErr == nil {
		var cmdArgs []string
		cmdArgs = append(cmdArgs, cmdKeyToolPath,
			"-importcert",
			"-trustcacerts",
			"-keystore", keyStorePathFinal,
			"-storetype", "pkcs12",
			"-storepass", certStorePassword,
			"-noprompt",
			"-v",
			"-alias", "agentstore",
			"-file", certFilePath)

		cmdKeyTool := &exec.Cmd{
			Path: cmdKeyToolPath,
			Args: cmdArgs,
		}

		cmdKeyTool.Stdout = &outb
		cmdKeyTool.Stderr = &errb
		// Execute the keytool command. Log an error an exit in case of any error.
		if err := cmdKeyTool.Run(); err != nil {
			errorMsg := fmt.Sprintf("Error occurred while creating keystore. Command Output: %v Error Output: %vError: %v",
				outb.String(), errb.String(), err.Error())
			return errors.New(errorMsg)
		} else {
			if logLevel >= LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf("Created keystore. Output: %v %v", outb.String(), errb.String()))
			}
		}

		// Change the permisions on the keystore
		err := os.Chmod(keyStorePathFinal, 0600)
		if err != nil {
			errorMsg := fmt.Sprintf(utils.MFT_FAILED_PERMISSION_KEYSTORE, keyStorePathFinal, err)
			return errors.New(errorMsg)
		}
	}
	return nil
}

// Generates a random 12 character password from the characters a-z, A-Z, 0-9
func generateRandomPassword() string {
	rand.Seed(time.Now().Unix())
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	validcharArray := []byte(validChars)
	password := ""
	for i := 0; i < 12; i++ {
		password = password + string(validcharArray[rand.Intn(len(validcharArray))])
	}

	return password
}

// Search the specified directory for certificate files
func getKeyFile(keysDir string, fileType string) string {
	fileList, err := os.ReadDir(keysDir)
	if err == nil && len(fileList) > 0 {
		for _, fileInfo := range fileList {
			if strings.Contains(fileInfo.Name(), fileType) {
				//Return the first file having extension specified
				return filepath.Join(keysDir, fileInfo.Name())
			}
		}
	}
	return ""
}
