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

/************************************************************************
* Environment variables used by image                                   *
************************************************************************/
// Level of diagnostic information logged. If this variable is
// not specified, LOG_LEVEL_INFO will be the default log level.
const MFT_LOG_LEVEL = "MFT_LOG_LEVEL"

// Name of the agent to be run by this image. This is a required
// parameter.
const MFT_AGENT_NAME = "MFT_AGENT_NAME"

// Name of the input file used for configuring an agent
// The contents of the input fill must be in JSON format
const MFT_AGENT_CONFIG_FILE = "MFT_AGENT_CONFIG_FILE"

// Envirornment variable pointing MFT Configuration and log data.
// If this variable is not specified, /mnt/mftdata will be the
// default path used for creating agent configuration and log files.
const BFG_DATA = "BFG_DATA"

// An agent might take some time to start after fteStartAgent command
// is issued. This is the time, in seconds, the containeror will wait
// for an agent to start. If an agent does not within the specified
// wait time, the container will end.
const MFT_AGENT_START_WAIT_TIME = "MFT_AGENT_START_WAIT_TIME"

// Enable agent tracing. "Yes" and "No" are the valid values with
// "No" being default
const MFT_AGENT_ENABLE_TRACE = "MFT_AGENT_ENABLE_TRACE"

// Display JSON formatted transfer logs on the console. "Yes" and "No"
// are the supported values with "No" being the default.
const MFT_AGENT_DISPLAY_CAPTURE_LOG = "MFT_AGENT_DISPLAY_CAPTURE_LOG"

// Enable tracing of MFT commands. "Yes" and "No" are the supported values
// with "No" being the default.
const MFT_TRACE_COMMANDS = "MFT_TRACE_COMMANDS"

// Environment variable pointing to path where files will be read from or written to
const MFT_MOUNT_PATH = "MFT_MOUNT_PATH"

// BFG_JVM_PROPERTIES
const MFT_BFG_JVM_PROPERTIES = "BFG_JVM_PROPERTIES"

// Configuration file containing details of logDNA or similar server
const MFT_AGENT_TRANSFER_LOG_PUBLISH_CONFIG_FILE = "MFT_TLOG_PUBLISH_INFO"

// Coordination queue manager cipherspec
const MFT_COORD_QMGR_CIPHER = "MFT_COORD_QMGR_CIPHER"

// Command queue manager cipherspec
const MFT_CMD_QMGR_CIPHER = "MFT_CMD_QMGR_CIPHER"

// Agent queue manager cipherspec
const MFT_AGENT_QMGR_CIPHER = "MFT_AGENT_QMGR_CIPHER"
