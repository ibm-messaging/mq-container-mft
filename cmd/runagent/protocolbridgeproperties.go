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

/*
* Contains a list of properties as defined in ProtocolBridgeProperties.xml.
* These properties are valid only for prototcol bridge agents. The methods
* in this file validate the properties specified in agent configuration
* JSON file.
 */
// HashMap of Protocol bridge property name and it's type
// 1 - string
// 2 - integer
// 3 - boolean
var bridgeProperies map[string]int = make(map[string]int)

func BuildBridgePropertyList() {
	bridgeProperies["defaultServer"] = DATA_TYPE_STRING
	bridgeProperies["maxActiveDestinationTransfers"] = DATA_TYPE_INT
	bridgeProperies["failTransferWhenCapacityReached"] = DATA_TYPE_BOOL
	bridgeProperies["name"] = DATA_TYPE_STRING
	bridgeProperies["type"] = DATA_TYPE_STRING
	bridgeProperies["host"] = DATA_TYPE_STRING
	bridgeProperies["port"] = DATA_TYPE_INT
	bridgeProperies["platform"] = DATA_TYPE_STRING
	bridgeProperies["timeZone"] = DATA_TYPE_STRING
	bridgeProperies["locale"] = DATA_TYPE_STRING
	bridgeProperies["fileEncoding"] = DATA_TYPE_STRING
	bridgeProperies["listFormat"] = DATA_TYPE_STRING
	bridgeProperies["listFileRecentDateFormat"] = DATA_TYPE_STRING
	bridgeProperies["listFileOldDateFormat"] = DATA_TYPE_STRING
	bridgeProperies["monthShortNames"] = DATA_TYPE_STRING
	bridgeProperies["limitedWrite"] = DATA_TYPE_BOOL
	bridgeProperies["maxListFileNames"] = DATA_TYPE_INT
	bridgeProperies["maxListDirectoryLevels"] = DATA_TYPE_INT
	bridgeProperies["maxSessions"] = DATA_TYPE_INT
	bridgeProperies["socketTimeout"] = DATA_TYPE_INT
	bridgeProperies["passiveMode"] = DATA_TYPE_BOOL
	bridgeProperies["connectionTimeout"] = DATA_TYPE_INT
	bridgeProperies["controlEncoding"] = DATA_TYPE_STRING
}

// Determine if the specified property is valid
func ValidateBridgeProperty(propertyName string) (bool, int) {
	typeV, ok := bridgeProperies[propertyName]
	return ok, typeV
}
