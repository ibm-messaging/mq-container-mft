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
	"fmt"
	"testing"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
)

func TestValidateCommandAttributesValidProperties(t *testing.T) {
	allAgentConfig, e := utils.ReadConfigurationDataFromFile("./data/test_agcfg.json")
	if e != nil {
		t.Fatal(e)
	}
	validated := ValidateCommandAttributes(allAgentConfig)
	if validated != nil {
		fmt.Printf("Command queue manager attributes validation failed. %v", validated)
		t.Fail()
	}
}

func TestValidateCommandAttributesMissingQmgrName(t *testing.T) {
	agentConfig := "{\"waitTimeToStart\":20,\"commandQMgr\":{\"host\":\"10.254.16.17\",	\"port\":1414,\"channel\":\"QS_SVRCONN\",\"additionalProperties\": {}}}"
	validated := ValidateCommandAttributes(agentConfig)
	if validated != nil {
		t.Log("Command queue manager attributes validation failed.")
	} else {
		t.Fail()
	}
}

func TestValidateCommandAttributesMissingQmgrHostName(t *testing.T) {
	agentConfig := "{\"waitTimeToStart\":20,\"commandQMgr\":{\"name\":\"SECUREQM\",	\"port\":1414,\"channel\":\"QS_SVRCONN\",\"additionalProperties\": {}}}"
	validated := ValidateCommandAttributes(agentConfig)
	if validated != nil {
		t.Log("Command queue manager attributes validation failed.")
	} else {
		t.Fail()
	}
}

func TestValidateCommandAttributesMissingChannelName(t *testing.T) {
	agentConfig := "{\"waitTimeToStart\":20,\"commandQMgr\":{\"name\":\"SECUREQM\",\"host\":\"10.254.16.17\",\"port\":1414,\"additionalProperties\": {}}}"
	validated := ValidateCommandAttributes(agentConfig)
	if validated == nil {
		t.Log("Command queue manager attributes with missing channel name validation passed as expected.")
	} else {
		t.Fail()
	}
}
