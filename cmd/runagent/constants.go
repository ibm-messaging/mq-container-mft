/*
© Copyright IBM Corporation 2020, 2022

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
* This file contains the list of constants used by the program
 */
// Source path of custom PBA credentials exit to copy from
const PBA_CUSTOM_CRED_EXIT_SRC_PATH = "/customexits"

// Name of the PBA custom credentials exit
const PBA_CUSTOM_CRED_EXIT_NAME = "com.ibm.wmq.bridgecredentialexit.jar"

// Complete path of PBA custom credentials exit
const PBA_CUSTOM_CRED_EXIT = "/customexits/mqft/pbaexit/com.ibm.wmq.bridgecredentialexit.jar"

// Third party jar used for reading JSON formatted data
const PBA_CUSTOM_CRED_DEPEND_LIB_NAME = "org.json.jar"

// Source path of third party JSON jar
const PBA_CUSTOM_CRED_DEPEND_LIB = "/customexits/mqft/pbaexit/org.json.jar"

// Path where files will be written to or read from. If a external storage is mounted
// to this path, then files will be written/read from the external storage otherwise
// path will be local to container and will get removed when container ends
const DEFAULT_MOUNT_PATH_FOR_TRANSFERS = "/mountpath/**"

// Source and Destination path for transfers
const MOUNT_PATH_TRANSFERS = "/mountpath"

// Controls the amount of diagnostic logs displayed on the console.
// Minimal diagnostic information. This is the default log level
const LOG_LEVEL_INFO = 1

// Detailed information
const LOG_LEVEL_VERBOSE = 2

// Supported agent types
const AGENT_TYPE_STANDARD = "STANDARD"
const AGENT_TYPE_BRIDGE = "BRIDGE"

// Supported log levels
const LOG_LEVEL_INFO_TXT = "info"
const LOG_LEVEL_VERBOSE_TXT = "verbose"

// log file types
const LOG_TYPE_TRANSFER = "tlog"
const LOG_TYPE_CAPTURE = "capure"
const LOG_TYPE_CONSOLE = "console"

// Server type
const LOG_SERVER_TYPE_DNA = "logDNA"
const LOG_SERVER_TYPE_ELK = "elk"
const LOG_SERVER_TYPE_DNA_NUM = 1
const LOG_SERVER_TYPE_ELK_NUM = 2

const KEY_TYPE = "type"
const KEY_URL_DNA = "logDNA.url"
const KEY_INJESTION_DNA = "logDNA.injestionKey"
const KEY_URL_ELK = "elk.url"

const DIR_AGENT_LOGS = "/mqft/logs/"
const DIR_AGENTS = "/agents/"

// License file path
const DIR_LICENSE_FILES = "/opt/mqm/mqft/licences/"

// Path used for storing keystores and truststores
const KEYSTORES_PATH = "/run/keystores"

// Coordination queue manager keystore file
const COORD_QM_KEYSTORE = "coordkeystore.p12"
const COORD_QM_TRUSTSTORE = "coordtruststore.p12"

// Command queue manager keystore file
const CMD_QM_KEYSTORE = "cmdkeystore.p12"
const CMD_QM_TRUSTSTORE = "cmdtruststore.p12"

// Agent queue manager keystore and trust file
const AGENT_QM_KEYSTORE = "agentkeystore.p12"
const AGENT_QM_TRUSTSTORE = "agenttruststore.p12"

const coordinationQMCertPath = "/etc/mqmft/pki/coordination"
const commandQMCertPath = "/etc/mqmft/pki/command"
const agentQMCertPath = "/etc/mqmft/pki/agent"

// Trim text
const TEXT_TRIM = " "

// Blank
const TEXT_BLANK = ""
const TEXT_YES = "yes"
const TEXT_NO = "no"

// Credential file names
const MFT_CMD_CRED_SLASH = "/cmdcredentials.xml"
const MFT_CORD_CRED_SLASH = "/coordcredentials.xml"
const MFT_AGENT_CRED_SLASH = "/agentcredentials.xml"
const MFT_USER_SANDBOX_SLASH = "/UserSandboxes.xml"

// MFT config path
const MFT_CONFIG_PATH_SUFFIX = "/mqft/config/"

// MFT log path
const MFT_LOG_PATH_SUFFIX = "/mqft/logs"

// Agents
const MFT_AGENTS_SLASH = "/agents/"
const MFT_EXITS_SLASH = "/exits/"

// command properties
const MFT_CMD_PROPS_SLASH = "/command.properties"
const MFT_CORD_PROPS_SLASH = "/coordination.properties"
const MFT_AGENT_PROPS_SLASH = "/agent.properties"
const MFT_PBA_PROPS_SLASH = "/ProtocolBridgeProperties.xml"

// Error codes returned by runagent process
const MFT_CONT_SUCCESS_CODE_0 = 0
const MFT_CONT_ERR_CODE_1 = 1
const MFT_CONT_ERR_CODE_2 = 2
const MFT_CONT_ERR_CODE_3 = 3
const MFT_CONT_ERR_CODE_4 = 4
const MFT_CONT_ERR_CODE_5 = 5
const MFT_CONT_ERR_CODE_6 = 6
const MFT_CONT_ERR_CODE_7 = 7
const MFT_CONT_ERR_CODE_8 = 8
const MFT_CONT_ERR_CODE_9 = 9
const MFT_CONT_ERR_CODE_10 = 10
const MFT_CONT_ERR_CODE_11 = 11
const MFT_CONT_ERR_CODE_12 = 12
const MFT_CONT_ERR_CODE_13 = 13
const MFT_CONT_ERR_CODE_14 = 14
const MFT_CONT_ERR_CODE_15 = 15
const MFT_CONT_ERR_CODE_16 = 16
const MFT_CONT_ERR_CODE_17 = 17
const MFT_CONT_ERR_CODE_18 = 18
const MFT_CONT_ERR_CODE_19 = 19
const MFT_CONT_ERR_CODE_20 = 20
const MFT_CONT_ERR_CODE_21 = 21
const MFT_CONT_ERR_CODE_22 = 22
const MFT_CONT_ERR_CODE_23 = 23
const MFT_CONT_ERR_CODE_24 = 24

// Data types used by ProtocolBridgeProperties.xml
const DATA_TYPE_STRING = 1
const DATA_TYPE_INT = 2
const DATA_TYPE_BOOL = 3
