/*
© Copyright IBM Corporation 2022, 2023

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
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

func TestUpdateAgentProperties(t *testing.T) {
	allAgentConfig, e := utils.ReadConfigurationDataFromFile("./data/test_agcfg.json")
	if e != nil {
		t.Fatal(e)
	}

	propFileName := "./data/agent.properties"
	singleAgentConfig := gjson.Get(allAgentConfig, "agents").Array()[0].String()
	err := UpdateProperties(propFileName, singleAgentConfig, "additionalProperties")
	if err == nil {
		updatedAgentPropFileContent, err := utils.ReadConfigurationDataFromFile(propFileName)
		if err == nil {
			result := gjson.Get(singleAgentConfig, "additionalProperties")
			result.ForEach(func(key, value gjson.Result) bool {
				searchToken := key.String() + "=" + value.String()
				if !strings.Contains(updatedAgentPropFileContent, searchToken) {
					t.Log("TestUpdateAgentProperties - failed: Contents don't match.")
					fmt.Printf("Updated agent properties: \n===\n%v\n===\n", updatedAgentPropFileContent)
					t.Fail()
					return false
				}
				return true // keep iterating
			})
			t.Log("Test TestUpdateAgentProperties passed as expected")
		} else {
			t.Log("TestUpdateAgentProperties - failed: Failed to read updated agent properties file.")
			fmt.Printf("TestUpdateAgentProperties - %v\n", err)
			t.Fail()
		}
	} else {
		t.Log("TestUpdateAgentProperties - failed: Failed to agent properties file.")
		fmt.Printf("TestUpdateAgentProperties - %v\n", err)
		t.Fail()
	}
}

func TestCreateUserSandbox(t *testing.T) {
	userSandboxFile := "UserSandbox.xml"
	userSBoxErr := createUserSandbox(userSandboxFile)
	if userSBoxErr == nil {
		sandBoxContents, _ := utils.ReadConfigurationDataFromFile(userSandboxFile)
		if strings.Contains(sandBoxContents, DEFAULT_MOUNT_PATH_FOR_TRANSFERS) {
			t.Log("Success")
		} else {
			t.Log("Sandbox does not contain required settings")
			t.Fail()
		}
	} else {
		t.Log("Errored: " + userSBoxErr.Error())
		t.Fail()
	}

	// Delete the UserSandBox.xml if it exists
	_, err := os.Lstat(userSandboxFile)
	if err == nil {
		os.Remove(userSandboxFile)
	}
}

func TestSetupCredentials(t *testing.T) {
	mqmftCredFile := "MQMFTCredentials.xml"
	var credentialsXmlToWrite string = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>" +
		"<tns:mqmftCredentials xmlns:tns=\"http://wmqfte.ibm.com/MQMFTCredentials\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\" xsi:schemaLocation=\"http://wmqfte.ibm.com/MQMFTCredentials MQMFTCredentials.xsd\">" +
		"    <tns:logger name=\"logger-name\" user=\"database-user\" password=\"database-password\" />" +
		"    <tns:file path=\"file-path\" password=\"file-password\"/>" +
		"    <tns:qmgr name=\"CoordQueueMgr\" user=\"John\" mqUserId=\"John\" mqPassword=\"JohnsPassword\" />" +
		"</tns:mqmftCredentials>"

	setupCredErr := setupCredentials(mqmftCredFile, credentialsXmlToWrite)
	if setupCredErr != nil {
		t.Log(setupCredErr.Error())
		t.Fail()
	} else {
		readCredentials, err := utils.ReadConfigurationDataFromFile(mqmftCredFile)
		if err == nil {
			if strings.EqualFold(credentialsXmlToWrite, readCredentials) {
				t.Log("Credentials file updated successfully")
				t.Log(fmt.Sprintf("Template Credentials:\n%v\n", credentialsXmlToWrite))
				t.Log(fmt.Sprintf("Credentials from file:\n%v\n", readCredentials))
			} else {
				t.Log("Credentials file contents different.")
				t.Log(fmt.Sprintf("Template Credentials:\n%v\n", credentialsXmlToWrite))
				t.Log(fmt.Sprintf("Credentials from file:\n%v\n", readCredentials))
				t.Fail()
			}
		} else {
			t.Log("Error occurred reading credentials file. " + err.Error())
			t.Fail()
		}
	}

	// Delete the UserSandBox.xml if it exists
	_, err := os.Lstat(mqmftCredFile)
	if err == nil {
		rErr := os.Remove(mqmftCredFile)
		if rErr != nil {
			t.Log(rErr.Error())
		}
	} else {
		t.Log(err.Error())
		t.Fail()
	}
}

func TestCreateKeyStore(t *testing.T) {
	certFilePath := "NonExistent.crt"
	keyStorePath := "key.p12"
	keyStorePwd := "StrongPassword"
	keyStoreDir := "data/tls"

	// Test for invalid certificate file path
	certNonExistentFileErr := CreateKeyStore(keyStoreDir, keyStorePath, certFilePath, keyStorePwd)
	if certNonExistentFileErr != nil {
		t.Log(certNonExistentFileErr.Error())
	} else {
		t.Fail()
	}

	err := os.MkdirAll("data/tls", 0700)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	certFilePath = "./data/certificate.crt"
	// Test with a valid certificate
	certExistErr := CreateKeyStore(keyStoreDir, keyStorePath, certFilePath, keyStorePwd)
	if certExistErr != nil {
		t.Log(certExistErr.Error())
		t.Fail()
	} else {
		t.Log("Certificate added")
	}
	_, err = os.Lstat(keyStorePath)
	if err == nil {
		rErr := os.RemoveAll(keyStorePath)
		if rErr != nil {
			t.Log(rErr.Error())
		}
	} else {
		t.Log(err.Error())
		// Just log error if we are unable to delete keystore
	}
}
