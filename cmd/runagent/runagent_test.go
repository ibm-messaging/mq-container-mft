/*
© Copyright IBM Corporation 2020, 2021

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
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
)

/*
 * Unit test program to test methods of runagent.
 */
func TestReadConfigurationDataFromFile(t *testing.T) {
	var jsonFileName string

	// Test valid JSON data.
	var pathNames []string = strings.Split(t.Name(), "/")
	if len(pathNames) == 2 {
		jsonFileName = pathNames[1]
	} else {
		jsonFileName = pathNames[0]
	}

	validJson, err := ioutil.TempFile("", jsonFileName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(validJson.Name())
	defer os.Remove(validJson.Name())
	validJsonF, err := os.OpenFile(validJson.Name(), os.O_WRONLY, 0700)
	if err != nil {
		t.Fatal(err)
	}
	configDataValid := "{\"dataPath\":\"/mqmft/mftdata\",\"monitoringInterval\":300,\"displayAgentLogs\":true,\"displayLineCount\":50,\"waitTimeToStart\":10,\"coordinationQMgr\":{\"name\":\"QUICKSTART\",\"host\":\"10.254.0.4\",\"port\":1414,\"channel\":\"MFT_HA_CHN\"},\"commandsQMgr\":{\"name\":\"QUICKSTART\",\"host\":\"10.254.0.4\",\"port\":1414,\"channel\":\"MFT_HA_CHN\"},\"agent\":{\"name\":\"KXAGNT\",\"type\":\"STANDARD\",\"qmgrName\":\"QUICKSTART\",\"qmgrHost\":\"10.254.0.4\",\"qmgrPort\":1414,\"qmgrChannel\":\"MFT_HA_CHN\",\"credentialsFile\":\"/usr/local/bin/MQMFTCredentials.xml\",\"protocolBridge\":{\"credentialsFile\":\"/usr/local/bin/ProtocolBridgeCredentials.xml\",\"serverType\":\"SFTP\",\"serverHost\":\"9.199.144.110\",\"serverTimezone\":\"\",\"serverPlatform\":\"UNIX\",\"serverLocale\":\"en-US\",\"serverFileEncoding\":\"UTF-8\",\"serverPort\":22,\"serverTrustStoreFile\":\"\",\"serverLimitedWrite\":\"\",\"serverListFormat\":\"\",\"serverUserId\":\"root\",\"serverPassword\":\"Kitt@n0or\"},\"additionalProperties\":{\"enableQueueInputOutput\":\"true\"}}"
	fmt.Fprintln(validJsonF, configDataValid)

	validJsonData, err := utils.ReadConfigurationDataFromFile(validJson.Name())
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(validJsonData)
	}

	// Test invalid JSON data.
	invalidJson, err := ioutil.TempFile("", jsonFileName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(invalidJson.Name())
	defer os.Remove(invalidJson.Name())
	invalidJsonF, err := os.OpenFile(invalidJson.Name(), os.O_WRONLY, 0700)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(validJson.Name())
	defer os.Remove(invalidJson.Name())

	configDataInValid := "{\"adataPath\":\"/mqmft/mftdata\",\"monitoringInterval\":300,\"displayAgentLogs\":true,\"displayLineCount\":50,\"waitTimeToStart\":10,\"coordinationQMgr\":{\"name\":\"QUICKSTART\",\"host\":\"10.254.0.4\",\"port\":1414,\"channel\":\"MFT_HA_CHN\"},\"commandsQMgr\":{\"name\":\"QUICKSTART\",\"host\":\"10.254.0.4\",\"port\":1414,\"channel\":\"MFT_HA_CHN\"},\"agent\":{\"name\":\"KXAGNT\",\"type\":\"STANDARD\",\"qmgrName\":\"QUICKSTART\",\"qmgrHost\":\"10.254.0.4\",\"qmgrPort\":1414,\"qmgrChannel\":\"MFT_HA_CHN\",\"credentialsFile\":\"/usr/local/bin/MQMFTCredentials.xml\",\"protocolBridge\":{\"credentialsFile\":\"/usr/local/bin/ProtocolBridgeCredentials.xml\",\"serverType\":\"SFTP\",\"serverHost\":\"9.199.144.110\",\"serverTimezone\":\"\",\"serverPlatform\":\"UNIX\",\"serverLocale\":\"en-US\",\"serverFileEncoding\":\"UTF-8\",\"serverPort\":22,\"serverTrustStoreFile\":\"\",\"serverLimitedWrite\":\"\",\"serverListFormat\":\"\",\"serverUserId\":\"root\",\"serverPassword\":\"Kitt@n0or\"},\"additionalProperties\":{\"enableQueueInputOutput\":\"true\"}}"
	fmt.Fprintln(invalidJsonF, configDataInValid)
	invalidatedJson, err := utils.ReadConfigurationDataFromFile(invalidJson.Name())
	if err != nil {
		t.Log(err)
	} else {
		t.Log(invalidatedJson)
		fmt.Println("Supplied agent configuration data has invalid attributes")
	}
}

