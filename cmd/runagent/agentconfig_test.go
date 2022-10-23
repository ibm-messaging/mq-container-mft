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
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
)

/*
 * Unit test program to test methods of runagent.
 */
func TestProtocolBridgePropertiesFileUpdates(t *testing.T) {
	allAgentConfig, e := utils.ReadConfigurationDataFromFile("./data/test_agcfg.json")
	if e != nil {
		t.Fatal(e)
	}

	agentCfg := gjson.Get(allAgentConfig, "agents").Array()
	//protocolBridgeConfigs := gjson.Get(agentCfg[0].String(), "protocolServers").Array()
	if updateProtocolBridgePropertiesFile("./data/test_pba.xml", agentCfg[0].String()) {
		updatedXml, err := utils.ReadConfigurationDataFromFile("./data/test_pba.xml")
		if err == nil {
			templateData, err := utils.ReadConfigurationDataFromFile("./data/test_pba_template.xml")
			if err == nil {
				if len(updatedXml) == len(templateData) {
					if strings.EqualFold(strings.Trim(templateData, ""), strings.Trim(updatedXml, "")) {
						fmt.Println("Protocol Bridge Properties file updated successfully")
					} else {
						fmt.Println("Protocol Bridge Properties file not updated successfully")
						fmt.Printf("Original file:\n%s\n", templateData)
						fmt.Printf("Updated file:\n%s\n ", updatedXml)
						t.Fail()
					}
				} else {
					fmt.Printf("Length of Updated XML %d length of template %d\n", len(updatedXml), len(templateData))
					fmt.Printf("Template file:\n%s\n Updated file\n%s\n", templateData, updatedXml)
					t.Fail()
				}
			} else {
				t.Fatal(err)
			}
		} else {
			t.Fatal(err)
		}
	} else {
		t.Fatal("Failed to update file")
	}
}

func TestUpdateBridgeParameters(t *testing.T) {

}
