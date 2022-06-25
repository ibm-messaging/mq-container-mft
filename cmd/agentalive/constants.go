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

// Contains constants and messages for angetready probe
// Constants must begin at 4000 as numbers 3000-3999 are reserverd for agentready application
const AGENT_ALIV_ENV_AGENT_NAME_NOT_SET_4001 = "IBMFT4001E: MFT_AGENT_NAME environment variable not specified."
const AGENT_ALIV_ENV_AGENT_CFG_FILE_NOT_SET_4002 = "IBMFT4002E: MFT_AGENT_CONFIG_FILE environment variable not specified."
const AGENT_ALIV_ENV_CFG_FILE_READ_4003 = "IBMFT4003E: An error occurred when attempting to read the configuration file [%s]. The error is: %v."
const AGENT_ALIV_NOT_RUNNING_4004 = "IBMFT4004E: Agent %s is not running."

// Constants
const AGENT_ALIV_EXIT_CODE_0 = 0
const AGENT_ALIV_EXIT_CODE_1 = 1
const AGENT_ALIV_EXIT_CODE_2 = 2
const AGENT_ALIV_EXIT_CODE_3 = 3
const AGENT_ALIV_EXIT_CODE_4 = 4
const AGENT_ALIV_EXIT_CODE_5 = 5
const AGENT_ALIV_EXIT_CODE_6 = 6
