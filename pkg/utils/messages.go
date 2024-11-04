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

package utils

/**
* This file contains messages displayed by the container.
 */
const MFT_CONT_DIAGNOSTIC_LEVEL_0001 = "Diangostic log level set to 'info'."
const MFT_CONT_DIAGNOSTIC_LEVEL_0002 = "Diagnostic log level set to 'verbose'."

const MFT_CONT_LICENES_NOT_ACCESSPTED_0004 = "License terms and conditions not accepted. License agreements and information can be viewed by setting the environment variable LICENSE=view.  You can also set the LANG environment variable to view the license in a different language. Set environment variable LICENSE=accept to indicate acceptance of license terms and conditions."
const MFT_CONT_RUNTIME_NAME_0005 = "Container Runtime: %s."
const MFT_CONT_ENV_AGENT_NAME_NOT_SPECIFIED_0006 = "Container failed to start as the MFT_AGENT_NAME environment variable was not specified. Resubmit the reqeust with MFT_AGENT_NAME environment variable specified with a valid agent name."
const MFT_CONT_ENV_AGENT_NAME_BLANK_0007 = "Container failed to start as the value specified in MFT_AGENT_NAME environment variable is blank. Resubmit the request with MFT_AGENT_NAME environment with a valid agent name."
const MFT_CONT_ENV_AGENT_START_TIME_0008 = "MFT_AGENT_START_WAIT_TIME is set to an invalid value. Defaulting to wait time of 10 seconds."
const MFT_CONT_ENV_BFG_DATA_BLANK_0009 = "A blank value was specified for BFG_DATA environment variable. Default path '/mnt/mftdata' will be used for agent configuration and logs."
const MFT_CONT_CONFIG_PATH_0010 = "Agent configuration and log directory: %s."
const MFT_CONT_ENV_AGNT_CFG_FILE_NOT_SPECIFIED_0011 = "Container failed start as MFT_AGENT_CONFIG_FILE environment variable was not specified. Resubmit the request specifying the MFT_AGENT_CONFIG_FILE environment variable with a valid path."
const MFT_CONT_ENV_AGNT_CFG_FILE_BLANK_0012 = "Container failed to start as MFT_AGENT_CONFIG_FILE as the value specified is blank. Resubmit the request specifying the MFT_AGENT_CONFIG_FILE environment variable with a valid path."
const MFT_CONT_CFG_FILE_READ_0013 = "An error occurred when attempting to read the configuration file [%s]. The error is: %v. Correct the error and resubmit the request."
const MFT_CONT_CFG_CORD_QM_NAME_MISSING_0014 = "Coordination queue manager name missing."
const MFT_CONT_CFG_CORD_QM_HOST_MISSING_0015 = "Coordination queue manager host name."
const MFT_CONT_CFG_MISSING_ATTRIBS_0016 = "An error occurred when validating agent configuration attributes from file %s. The errors is %s."
const MFT_CONT_CFG_CMD_QM_NAME_MISSING_0017 = "Command queue manager name missing."
const MFT_CONT_CFG_CMD_QM_HOST_MISSING_0018 = "Command queue manager host name missing."
const MFT_CONT_CFG_AGENT_CONFIG_MISSING_0019 = "Information required to configure agent %s was not found in file %s. Container will end now. Update the configuration file with required attributes and resubmit the request."
const MFT_CONT_CFG_AGENT_NAME_MISSING_0020 = "Agent name missing from configuration file."
const MFT_CONT_CFG_AGENT_QM_NAME_MISSING_0021 = "Agent queue manager name missing from configuration file."
const MFT_CONT_CFG_AGENT_QM_HOST_MISSING_0022 = "Agent queue manager host name missing from configuration file."
const MFT_CONT_CFG_AGENT_CONFIG_ERROR_0023 = "An error occurred when attempting validate required agent attributes from configuration file %s. The error is: %v."
const MFT_CONT_CFG_CORD_CONFIG_MSG_0024 = "Setting up coordination configuration for agent %s. Name of the coordination queue manager %s."
const MFT_CONT_CFG_CORD_CONFIG_CRED_NOT_EXIST_0025 = "File %s provided in MFT_AGENT_CREDENTIAL_FILE environment variable does not exist or does not have access permission."
const MFT_CONT_CFG_CORD_CONFIG_CRED_IGNORED_0026 = "Path provided in MFT_AGENT_CREDENTIAL_FILE environment variable is blank and has been ignored."
const MFT_CONT_CFG_CORD_CONFIG_CRED_PATH_0027 = "Coordination queue manager credential path %s."
const MFT_CONT_CMD_NOT_FOUND_0028 = "Command not found. The error is: %v."
const MFT_CONT_CORD_CFG_FAILED_0029 = "Failed to create coordination queue manager configuration. Container will end now. Review and fix errors and resubmit the request."
const MFT_CONT_CMD_CFG_FAILED_0030 = "Failed to create command queue manager configuration. Container will end now. Review and fix errors and resubmit the request."
const MFT_CONT_AGNT_CFG_FAILED_0031 = "Failed to configure agent %s. Container will end now. Review and fix any errors and then resubmit request."
const MFT_CONT_AGNT_START_FAILED_0032 = "Failed to start agent %s. Container will end now. Review and fix any errors and then resubmit request."
const MFT_CONT_AGNT_NOT_STARTED_0033 = "Agent %s has not started yet. Status will be verified again after %d seconds."
const MFT_CONT_AGNT_FAILED_TO_START_0034 = "Agent %s did not start."
const MFT_CONT_AGNT_WAIT_MIRROR_CMP_0035 = "Waiting for log mirroring to complete for agent %s."
const MFT_CONT_AGNT_WAIT_MIRROR_STOP_0036 = "Stopping log mirroring for agent %s."
const MFT_CONT_AGNT_CAPT_LOG_ERROR_0037 = "%s is not a valid value for MFT_AGENT_DISPLAY_CAPTURE_LOG environment variable. Transfer logs will not be displayed on the console."
const MFT_CONT_AGNT_STARTED_0038 = "Agent %s has started."
const MFT_CONT_AGNT_CFG_DELETED_0039 = "Configuration of agent %s not deleted."
const MFT_CONT_AGNT_START_FAILED_0040 = "Agent %s failed to start. Container will end now. Review and fix any errors and resubmit the request."
const MFT_CONT_AGNT_STARTING_0041 = "Starting agent %s."
const MFT_CONT_CMD_ERROR_0042 = "Command output: %s\nError: %s."
const MFT_CONT_CMD_INFO_0043 = "Command output: %s."
const MFT_CONT_AGNT_VRFY_STATUS_0044 = "Verifying status of agent %s."
const MFT_CONT_AGNT_INVALID_TYPE_0045 = "%s is an invalid agent type. Defaulting type to %s."
const MFT_CONT_AGNT_CREATING_0046 = "Creating %s type configuration for agent %s."
const MFT_CONT_AGNT_CREATED_0047 = "Configuration for agent %s has been created."
const MFT_CONT_AGNT_CLN_0048 = "Invalid value %s specified for cleanOnStart attribute. The option has been ignored."
const MFT_CONT_AGNT_DLTNG_0049 = "Deleting configuration for agent %s."
const MFT_CONT_AGNT_DLTED_0050 = "Configuration of agent %s has been deleted."
const MFT_CONT_AGNT_CLN_0051 = "Cleaning %s from agent %s."
const MFT_CONT_AGNT_ITEM_CLN_0052 = "All %s have been deleted from agent %s."
const MFT_CONT_AGNT_RM_CRT_0053 = "Creating resource monitor %s."
const MFT_CONT_CORD_SETUP_COMP_0054 = "Coordination configuration for %s is complete."
const MFT_CONT_CMD_SETUP_STRT_0055 = "Setting up commands configuration for agent %s. Name of the command queue manager: %s."
const MFT_CONT_CMD_QMGR_CRED_PATH_0056 = "Command queue manager credential path %s."
const MFT_CONT_CMD_SETUP_COMP_0057 = "Commands configuration for %s is complete."
const MFT_CONT_CRED_ENCRYPTING_0058 = "Encrypting credentials file %s."
const MFT_CONT_CRED_ENCRYPTED_0059 = "Credentials file %s has been encrypted."
const MFT_CONT_CRED_DECODE_FAILED_0060 = "An error occurred while decoding base64 encoded data. The error is %v."
const MFT_CONT_CRED_NOT_AVAIL_0061 = "Credentials for connecting to queue manager %s have not been provided."
const MFT_CONT_CRED_NOT_AVAIL_ASM_DFLT_0062 = "Failed to decode password provided. Assuming it is not base64 encoded."
const MFT_CONT_ERR_CONT_USER_0063 = "An error occurred while determing current user. The error is: %v."
const MFT_CONT_ERR_OPN_CRED_FILE_0064 = "An error occurred while opening credential file %s. The error is: %v."
const MFT_CONT_ERR_OPN_SNDBOX_FILE_0065 = "An error occurred while opening sandbox file %s. The error is: %v."
const MFT_CONT_ERR_UPDTING_FILE_0066 = "An error occurred while updating file %s. The error is: %v."
const MFT_CONT_ERR_OPN_FILE_0067 = "An error occurred while opening file %s. The error is: %v."
const MFT_CONT_AGENT_STOPPED_0068 = "Agent %s has been stopped."
const MFT_CONT_SIGNAL_CHILD_0069 = "Received SIGCHLD signal."
const MFT_CONT_SIGNAL_LISTEN_0070 = "Listening for SIGCHLD signals."
const MFT_CONT_SIGNAL_RECD_0071 = "Received signal %v."
const MFT_CONT_REAPED_PID_0072 = "Reaped process ID %v."
const MFT_CONT_DIAGNOSTIC_LEVEL_0073 = "Unknown diagnostic level specified. Defaulting to 'info'."
const MFT_CONT_LIC_ERROR_OCCUR_0074 = "An error occurred while checking for license. The error is :%v."
const MFT_CONT_RUNTM_ERROR_OCCUR_0075 = "An error occurred while determining container runtime. The error is :%v."
const MFT_CONT_AGNT_ALL_ITEM_CLN_0076 = "All objects from agent %s have been deleted."
const MFT_CONT_AGNT_PROC_NOT_RUNING_0077 = "An error occurred while determining the agent status. The error is: %v."
const MFT_CONT_AGNT_TRANSFER_LOG_ERROR_0078 = "%s is not a valid value for MFT_AGENT_PUSH_TRANSFER_LOGS_TO_SERVER environment variable. Transfer logs will not be published to specified server."
const MFT_CONT_BRIDGE_PROPERTY_NOT_SET = "A mandatory property '%s' for configuring bridge agent was not specified for server %s."
const MFT_CONT_BRIDGE_NOT_ENOUGH_INFO = "Information required to setup bridge agent not found. Can not continue."
const MFT_FAILED_OPEN_FILE = "An error occurred while opening file %s. The error is: %v"
const MFT_FAILED_WRITE_DATA = "An error occurred while writing data to file %s. The error is: %v"
const MFT_FAILED_DELETE_FILE = "An error occurred while deleting file %s. The error is: %v"
const MFT_FAILED_WRITING_SANDBOX = "An error occurred while updating agent sandbox. The error is: %v"
const MFT_CONT_NO_AGENT_CONFIG_SUPPLIED = "Configuration required for agent creation not found in supplied file %s. Container creation can not continue and will end now."
const MFT_CONT_MTLS_NOT_CONFIGURED = "Mutual TLS not configured."
const MFT_CONT_CORDQMGR_NON_SECURE_CONN = "Commands will use non-secure connection to coordination queue manager."
const MFT_CONT_CMDQMGR_NON_SECURE_CONN = "Commands will use non-secure connection to command queue manager."
const MFT_CONT_UPDATED_CMD_CONFIG = "Updated command configuration - %v."
const MFT_CONT_AGNTQMGR_NON_SECURE_CONN = "Agent will use non-secure connections to agent queue manager."
const MFT_CONT_KEYSTORE_CREATE_FAILED = "An error occurred while creating keystore %s. The error is: %v"
const MFT_CONT_AGNT_NOT_READY = "Agent %s is not ready. Container will end now. Review and fix any errors and then resubmit request."
const MFT_CONT_AGNT_NOT_READY_ERROR = "An error occurred while verifying status of agent %s. The error is %v."
const MFT_AGENT_NAME_CONFIGURE = "Creating configuration for agent %s."
const MFT_AGENT_JSON_CONFIG = "Configuration information of the agent: %v."
const MFT_AGENT_NAME_CONFIG_FILE = "Name of the agent found in configuration file %v."
const MFT_UPDATED_CONFIGURATION = "Updated coordination configuration - %v."
const MFT_PBA_HOST_AND_TYPE_NOT_FOUND = "Protocol server host name and type not supplied in the configuration file %s. Configuration will not be updated."
const MFT_FAILED_PERMISSION_KEYSTORE = "Error occurred while setting persmission to keystore %v. The error is %v."
const MFT_ENV_AGNT_CFG_FILE_NOT_SPECIFIED = "MFT_AGENT_CONFIG_FILE environment variable has not specified. Container will attempt to load agent configuration from file %s."

const AGENT_REDY_ENV_AGENT_NAME_NOT_SET_3001 = "IBMFT3001E: MFT_AGENT_NAME environment variable not specified."
const AGENT_REDY_ENV_AGENT_CFG_FILE_NOT_SET_3002 = "IBMFT3002E: MFT_AGENT_CONFIG_FILE environment variable not specified."
const AGENT_REDY_ENV_CFG_FILE_READ_3003 = "IBMFT3003E: An error occurred when attempting to read the configuration file [%s]. The error is: %v."
const AGENT_REDY_NOT_RUNNING_3004 = "IBMFT3004E: Agent %s is not running."
const AGENT_REDY_EVNT_NOT_FOUND_3005 = "IBMFT3005E: Agent ready event not found in output0.log file."

// Contains constants and messages for angetready probe
// Constants must begin at 4000 as numbers 3000-3999 are reserverd for agentready application
const AGENT_ALIV_ENV_AGENT_NAME_NOT_SET_4001 = "IBMFT4001E: MFT_AGENT_NAME environment variable not specified."
const AGENT_ALIV_ENV_AGENT_CFG_FILE_NOT_SET_4002 = "IBMFT4002E: MFT_AGENT_CONFIG_FILE environment variable not specified."
const AGENT_ALIV_ENV_CFG_FILE_READ_4003 = "IBMFT4003E: An error occurred when attempting to read the configuration file [%s]. The error is: %v."
const AGENT_ALIV_NOT_RUNNING_4004 = "IBMFT4004E: Agent %s is not running."
