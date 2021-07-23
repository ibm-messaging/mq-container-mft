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
	"os"
	"os/exec"
    "io/ioutil"
	"strings"
	"bytes"
	"time"
	"github.com/tidwall/gjson"
	"errors"
	"sync"
	"context"
	"strconv"
	"fmt"
	"github.com/ibm-messaging/mq-container-mft/cmd/utils"
)

// Name of the custom PBA credentials exit
const PBA_CUSTOM_CRED_EXIT_SRC_PATH="/customexits"
const PBA_CUSTOM_CRED_EXIT_NAME="bridgecredexit.jar"
const PBA_CUSTOM_CRED_DEPEND_LIB_NAME="json-20210307.jar"
const PBA_CUSTOM_CRED_EXIT = "/customexits/mqft/pbaexit/bridgecredexit.jar"
const PBA_CUSTOM_CRED_DEPEND_LIB = "/customexits/mqft/pbaexit/json-20210307.jar"

const MOUNT_PATH_FOR_TRANSFERS = "/mountpath/**"
const LOG_LEVEL_INFO =    1
const LOG_LEVEL_VERBOSE = 2

// Main entry point to program. This application configures an agent and starts it.
func main () {
	var bfgDataPath string
	var bfgConfigFilePath string 
	var allAgentConfig string 
	var e error
	
	// By default verbose logging is not enabled.
	logLevel := LOG_LEVEL_INFO
	
	// To display agent logs or not.
	logLevelStr, logLevelSet := os.LookupEnv("MFT_LOG_LEVEL")
	if logLevelSet {
		if logLevelStr == "verbose" {
			logLevel = LOG_LEVEL_VERBOSE
			utils.PrintLog(fmt.Sprintf("Log level '%s'", "verbose"))
		} else {
			utils.PrintLog(fmt.Sprintf("Log level '%s'", "info"))
		}
	}
	
	// First check if license is accepted or not.
	accepted, err := checkLicense()
	if err != nil {
		utils.PrintLog(fmt.Sprintf("Error checking license acceptance: %v", err))
		os.Exit(1)
	}
	
	// License was not accepted
	if !accepted {
		utils.PrintLog(fmt.Sprintf("%v", errors.New("License not accepted")))
		os.Exit(1)
	}
	
	// Exit if we are not running inside a known container type like Docker/Kube etc.
    runtime, err := DetectRuntime()
	if err != nil && err != ErrContainerRuntimeNotFound {
		utils.PrintLog(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
		
	// We are running in a container, so just print it on console
	utils.PrintLog(fmt.Sprintf("Container Runtime: %s", runtime))
	
	agentNameEnv, agentNameSet := os.LookupEnv("MFT_AGENT_NAME")
	if !agentNameSet || agentNameEnv == "" {
		utils.PrintLog("MFT_AGENT_NAME environment variable not set or the value of the variable is empty.")
		os.Exit(1)
	}
	
	// Read config file from a fixed path
	bfgConfigFilePath, configFileSet := os.LookupEnv("MFT_AGENT_CONFIG_FILE")    
	if !configFileSet || bfgConfigFilePath == "" {
		utils.PrintLog("MFT_AGENT_CONFIG_FILE environment variable not set or the value of the variable is empty.")
		os.Exit(1)
	}
	
	// Read the entire agent configuration data from JSON file. The configuration file
	// may contain data for multiple agents. We will choose data for matching agent name.
	allAgentConfig, e = utils.ReadConfigurationDataFromFile(bfgConfigFilePath)
	if e != nil {
		// Exit if we had any error when reading configuration file
		utils.PrintLog(fmt.Sprintf("%v", e))
		os.Exit(1)
	}
	
	// Validate coordination queue manager attributes
	if validateCoordinationAttributes (allAgentConfig) != nil {
		utils.PrintLog("Supplied coordination queue manager parameters are not valid")
		os.Exit(1)
	}

	// Validate command queue manager attributes
	if validateCommandAttributes (allAgentConfig) != nil {
		utils.PrintLog("Supplied command queue manager parameters are not valid")
		os.Exit(1)
	}
	
	// See if we have been given mount point for creating configuration directory.
	bfgConfigMountPath, bfgCfgMountPathSet := os.LookupEnv("BFG_DATA")
	if bfgCfgMountPathSet && len(bfgConfigMountPath) > 0 {
		bfgDataPath = bfgConfigMountPath
		err = utils.CreateDataPath (bfgDataPath)
		// The directory might alrady exist. Ignore error in such cases.
		if err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
		}
	} else {
	    // Make the default BFG_DATA path as /mnt/mftdata if BFG_DATA is not set.
	    bfgDataPath = utils.FIXED_BFG_DATAPATH
		// First check if the specified data directory exists.
		err = utils.CreateDataPath (bfgDataPath)
		if err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
			os.Exit (1)
		}
		
		// Set BFG_DATA environment variable so that we can run MFT commands.
		os.Setenv("BFG_DATA", bfgDataPath)
    }
	utils.PrintLog(fmt.Sprintf("Agent configuration and log directory: %s", bfgDataPath))
		
	// Time to wait for agent to start. 
	// Default wait time is 10 seconds
	delayTimeStatusCheck := time.Duration(10) * time.Second
	timeWaitForAgentStartStr, timeWaitForAgentStartSet := os.LookupEnv("WAIT_TIME_TO_START")
	// Value is numeric and above 0, then use it.
	if timeWaitForAgentStartSet && utils.IsNumeric(timeWaitForAgentStartStr) {
		waitTime := utils.ToNumber(timeWaitForAgentStartStr)
		if waitTime > 0 {
			delayTimeStatusCheck = time.Duration(waitTime) * time.Second
		}
	}
	
	// Cache the coordination queue manager name
	coordinationQMgr := gjson.Get(allAgentConfig, "coordinationQMgr.name").String()

	// Setup coordination configuration
	coordinationCreated := setupCoordination (allAgentConfig, bfgDataPath, agentNameEnv,  logLevel)
	if !coordinationCreated {
		utils.PrintLog("fteSetupCoordination failed. Container will end now.")
		os.Exit(1)	
	}
	
	// Setup command configuration
	commandsCreated := setupCommands(allAgentConfig, bfgDataPath, agentNameEnv, logLevel)
	if !commandsCreated{
		utils.PrintLog("fteSetupCommand failed. Container will end now.")
		os.Exit(1)	
	}
	
	// We may have multiple agent configurations defined in the JSON file. Iterate through all
	// definitions and pick the one that has matching agent name specified in environment variable
	var singleAgentConfig string
	configurationFound := false 
	agentsJson := gjson.Get(allAgentConfig, "agents").Array()
	for i := 0; i < len(agentsJson); i++ {
		singleAgentConfig = agentsJson[i].String()
		agentNameConfig := gjson.Get(singleAgentConfig,"name").String()
		if strings.Contains(agentNameConfig, agentNameEnv) {
			utils.PrintLog(fmt.Sprintf("Required configuration information found for agent '%s'", agentNameEnv))
			configurationFound = true
			break
		}
	}
	
	// Exit if we did not find the configuration for specified agent 
	if !configurationFound {
		utils.PrintLog(fmt.Sprintf("Configuration not found for agent %s. Container will end now.", agentNameEnv))
		os.Exit(1)
	}

	// Validate if required agent attributes have been specified
	if validateAgentAttributes	(singleAgentConfig) != nil {
		utils.PrintLog("Required agent attributes were not found in the configuration file")
		os.Exit(1)	
	}
	
	// Create the specified agent configuration
	setupAgentDone := setupAgent(singleAgentConfig, bfgDataPath, coordinationQMgr, logLevel)
	if !setupAgentDone {
		utils.PrintLog("Agent creation failed. Container will end now.")
		os.Exit(1)	
	}
	
	// Clean agent if asked for before starting the agent
	cleanAgent (singleAgentConfig, coordinationQMgr, agentNameEnv, logLevel)
	
	// Submit request to start the agent.
	startAgentDone := startAgent(agentNameEnv, coordinationQMgr, logLevel) 
	if !startAgentDone {
		utils.PrintLog("Agent start command failed. Container will end now.")
		os.Exit(1)	
	}
	
	// Verify the agent status
	agentStatus := verifyAgentStatus(coordinationQMgr, agentNameEnv, logLevel)
	if strings.Contains(agentStatus,"STOPPED") == true {
		//if agent not started yet, wait for some time and then reissue fteListAgents commad
		utils.PrintLog(fmt.Sprintf("Agent not started yet. Wait for %d seconds and verify agent status again", delayTimeStatusCheck/time.Second))
		time.Sleep(delayTimeStatusCheck)
		agentStatus = verifyAgentStatus(coordinationQMgr, agentNameEnv, logLevel)
	}
	
	var wg sync.WaitGroup
	defer func() {
		fmt.Println("Waiting for log mirroring to complete")
		wg.Wait()
	}()

	ctxAgentLog, cancelMirrorAgentLog := context.WithCancel(context.Background())
	defer func() {
		fmt.Println("Cancel mirror logging")
		cancelMirrorAgentLog()
	}()

	// Display the contents of agent's output0.log file on the console.
	if logLevel == LOG_LEVEL_VERBOSE {
		agentLogPath := bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentNameEnv + "/logs/output0.log"
		mirrorAgentLogs(ctxAgentLog, &wg, agentNameEnv, agentLogPath)
	}
	
	agentTraceEnv, agentTraceEnvSet := os.LookupEnv("MFT_AGENT_TRACE")
	if agentTraceEnvSet {
		if agentTraceEnv == "yes" {
			// Show contents of last trace file
			if gjson.Get(singleAgentConfig, "trace").Exists(){
				// Do this only if trace is enabled in agent properties file.
				var traceValue string = gjson.Get(singleAgentConfig, "trace").String()
				if traceValue != ""{
					agentPidPath := bfgDataPath + "/mqft/logs/" + coordinationQMgr  + "/agents/" + agentNameEnv + "/agent.pid"
					agentPid := utils.GetAgentPid(agentPidPath)
					agentTracePath := bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentNameEnv + "/logs/trace" + strconv.Itoa(int(agentPid)) + "/trace" + strconv.Itoa(int(agentPid)) + ".txt.0"
					mirrorAgentLogs(ctxAgentLog, &wg, agentNameEnv, agentTracePath)
				}
			}
		}
	}

	// Mirror capture0.log
	agentCaptureLogEnv, agentCaptureLogEnvSet := os.LookupEnv("MFT_AGENT_CAPTURE_LOG")
	if agentCaptureLogEnvSet {
		if agentCaptureLogEnv == "yes" {
			captureLogPath := bfgDataPath + "/mqft/logs/" + coordinationQMgr + "/agents/" + agentNameEnv + "/logs/capture0.log"
			mirrorAgentLogs(ctxAgentLog, &wg, agentNameEnv, captureLogPath)
		}
	}

	// If agent status is READY or ACTIVE, then we are good. 
	if (strings.Contains(agentStatus,"READY") == true || strings.Contains(agentStatus,"ACTIVE") == true) ||
		utils.IsAgentReady (bfgDataPath, agentNameEnv, coordinationQMgr) {
		utils.PrintLog(fmt.Sprintf("Agent '%s' has started", agentNameEnv))

		// Setup a siganl handle and wait for till container is stopped.
		signalControl := signalHandler(agentNameEnv, coordinationQMgr)
		<-signalControl
		cancelMirrorAgentLog()
		
		// Delete agent configuration on exit
		deleteAgentOnExit := gjson.Get(singleAgentConfig, "deleteOnTermination").Bool()
		// Delete agent configuration if asked for
		if deleteAgentOnExit {
			deleteAgent(coordinationQMgr, agentNameEnv)
		}
		
		// Agent has ended. Return success
		os.Exit(0)
	} else {
		utils.PrintLog("Agent failed to start. Container will end now.")
		os.Exit(1)
	}
}

// Call fteStartAgent command to submit a request to start an agent.
func startAgent(agentName string, coordinationQMgr string, logLevel int) (bool) {
	var outb, errb bytes.Buffer
	var startSubmitted bool = false

	// Get the path of MFT fteStartAgent command. 
	cmdStrAgntPath, lookPathErr:= exec.LookPath("fteStartAgent")
	if lookPathErr == nil {
		// We are done with creating agent. Start it now.
		utils.PrintLog(fmt.Sprintf("Starting agent '%s'", agentName))
		cmdStrAgnt := &exec.Cmd {
				Path: cmdStrAgntPath,
				Args: [] string {cmdStrAgntPath,"-p", coordinationQMgr, agentName}}
  
		cmdStrAgnt.Stdout = &outb
		cmdStrAgnt.Stderr = &errb
		// Run fteStartAgent command. Log an error and exit in case of any error.
		if err := cmdStrAgnt.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf("Command: %s\nError: %s", outb.String(), errb.String()))
		} else {
			if logLevel == LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			}
			startSubmitted = true
		}
	} else {
		utils.PrintLog(fmt.Sprintf("fteStartAgent command not found. %s", lookPathErr))
	}
	return startSubmitted
}

// Verify the status of agent by calling fteListAgents command.
func verifyAgentStatus(coordinationQMgr string, agentName string, logLevel int)(string) {
	var outb, errb bytes.Buffer
	var agentStatus string

	utils.PrintLog(fmt.Sprintf("Verifying status of agent '%s'", agentName))
	cmdListAgentPath, lookPathErr := exec.LookPath("fteListAgents")
	if lookPathErr == nil {
		cmdListAgents := &exec.Cmd {
			Path: cmdListAgentPath,
			Args: [] string {cmdListAgentPath, 
						"-p", coordinationQMgr, agentName}}

		cmdListAgents.Stdout = &outb
		cmdListAgents.Stderr = &errb
		// Execute and get the output of the command into a byte buffer
		if err := cmdListAgents.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			utils.PrintLog(fmt.Sprintf("Error %s", errb.String()))
		} else {
			if logLevel == LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			}
			// Now parse the output of fteListAgents command and take appropriate actions.
			agentStatus = outb.String()
		}
	} else {
		utils.PrintLog("Failed to find fteListAgents command")
	}
	
	return agentStatus
}

// Calls fteCreateAgent/fteCreateBridgeAgent commands to setup agent configuration
func setupAgent(agentConfig string, bfgDataPath string, coordinationQMgr string, logLevel int) (bool) {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
    var cmdSetup bool = false
	
	// Create agent.
	utils.PrintLog(fmt.Sprintf("Creating '%s' type configuration for agent '%s'", gjson.Get(agentConfig, "type"), gjson.Get(agentConfig, "name")))
			
	var cmdCrtAgnt * exec.Cmd
	var standardAgent bool
	// Determine if we will be creating a standard or a bridge agent
	if strings.EqualFold(gjson.Get(agentConfig, "type").String(), "STANDARD") == true {
		standardAgent = true
	} else {
		standardAgent = false
	}
	
	// Keep the agent name
	agentName := gjson.Get(agentConfig,"name").String()
	agentQMgrName := gjson.Get(agentConfig,"qmgrName").String()
	agentQMgrHost := gjson.Get(agentConfig,"qmgrHost").String()
	
	var agentQMgrPort string
	if gjson.Get(agentConfig,"qmgrPort").Exists() {
		agentQMgrPort = gjson.Get(agentConfig,"qmgrPort").String()
	} else {
		// Not found, use default 1414
		agentQMgrPort = "1414"
	}
	
	var agentQMgrChannel string
	if gjson.Get(agentConfig,"qmgrChannel").Exists() {
		agentQMgrChannel = gjson.Get(agentConfig,"qmgrChannel").String()
	} else {
		// Default to SYSTEM.DEF.SVRCONN
		agentQMgrChannel = "SYSTEM.DEF.SVRCONN"
	}
	
	// We are creating a STANDARD agent
	if  standardAgent {
		// Get the path of MFT fteCreateAgent command. 
		cmdCrtAgntPath, lookPathErr := exec.LookPath("fteCreateAgent")
		if lookPathErr == nil {
			// Creating a standard agent
			var  params [] string
			params = append(params, cmdCrtAgntPath,
						"-p", coordinationQMgr, 
						"-agentName", agentName, 
						"-agentQMgr", agentQMgrName, 
						"-agentQMgrHost", agentQMgrHost, 
						"-agentQMgrPort", agentQMgrPort, 
						"-agentQMgrChannel", agentQMgrChannel, "-f")
			
			// Use credentials file if one is specified
			credFile := gjson.Get(agentConfig,"credentialsFile")
			if credFile.Exists() {
				params = append(params, "-credentialsFile", credFile.String())
			}
			// Now build the command to create standard agent
			cmdCrtAgnt = &exec.Cmd {Path: cmdCrtAgntPath, Args: params }
			cmdSetup = true
		} else {
			utils.PrintLog(fmt.Sprintf("fteCreateAgent command not found. %s", lookPathErr))
		}
	} else {
		// We are creating a BRIDGE agent
		// Get the path of MFT fteCreateBridgeAgent command
		cmdCrtBridgeAgntPath, lookPathErr := exec.LookPath("fteCreateBridgeAgent")
		if lookPathErr == nil {
			// Creating a bridge agent
			var  params [] string 
			params = append(params, cmdCrtBridgeAgntPath,  
						"-p", coordinationQMgr, 
						"-agentName", agentName, 
						"-agentQMgr", agentQMgrName, 
						"-agentQMgrHost", agentQMgrHost, 
						"-agentQMgrPort", agentQMgrPort, 
						"-agentQMgrChannel", agentQMgrChannel, "-f")

			// Use credentials file if one is specified
			credFile := gjson.Get(agentConfig,"credentialsFile")
			if credFile.Exists() {
				params = append(params, "-credentialsFile", credFile.String())
			}
						
			// Now build the command to create a bridge agent
			bridgeParams := processBridgeParameters(agentConfig, params)
			cmdCrtAgnt = &exec.Cmd {
				Path: cmdCrtBridgeAgntPath,
				Args: bridgeParams,
			}
			cmdSetup = true
		} else {
			utils.PrintLog(fmt.Sprintf("fteCreateBridgeAgent command not found. %s", lookPathErr))
		}
	}
	
	// Ready to execute the command
	if cmdSetup == true {
		// Reuse the same buffer
		cmdCrtAgnt.Stdout = &outb
		cmdCrtAgnt.Stderr = &errb
		
		// Execute the fteCreateAgent/fteCreateBridgeAgent to create agent configuration.
		// Log an error an exit in case of any error.
		if err := cmdCrtAgnt.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			utils.PrintLog(fmt.Sprintf("Error %s", errb.String()))
			utils.PrintLog(fmt.Sprintf("Create Agent command failed. The error is %s", err))
		} else {			
			// If it is bridge agent, then update the ProtocolBridgeProperties file with any additional properties specified.
			if !standardAgent {
				// Copy the custom credentials exit to agent's exit directory.
				protocolBridgeCustExit := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/exits/bridgecredexit.jar" 
				utils.CopyFile(PBA_CUSTOM_CRED_EXIT, protocolBridgeCustExit)
				protocolBridgeCustExitDependLib := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/exits/json-20210307.jar" 
				utils.CopyFile(PBA_CUSTOM_CRED_DEPEND_LIB, protocolBridgeCustExitDependLib)
				
				// Delete the custom exit from source directory
				err := os.RemoveAll(PBA_CUSTOM_CRED_EXIT_SRC_PATH)
				if err != nil {
					utils.PrintLog(fmt.Sprintf("%v", err))
				}
				//utils.DeleteDir(PBA_CUSTOM_CRED_EXIT)
				//utils.DeleteDir(PBA_CUSTOM_CRED_DEPEND_LIB)
			}
			
			// Update agent properties file with additional attributes specified.
			agentPropertiesFile := bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/agent.properties"
			updateAgentProperties(agentPropertiesFile, agentConfig, "additionalProperties", !standardAgent);
			
			// Update UserSandbox XML file - valid only for STANDARD agents
			if standardAgent {
				createUserSandbox (bfgDataPath + "/mqft/config/" + coordinationQMgr + "/agents/" + agentName + "/UserSandboxes.xml")
			}
			// Tell user that agent has been configured.
			utils.PrintLog(fmt.Sprintf("Configuration for agent '%s' has been created", gjson.Get(agentConfig,"name").String()))
			created = true
		}
	}
	
	return created
}

// Setup coordination configuration for agent.
func setupCoordination(allAgentConfig string, bfgDataPath string, agentNameEnv string, logLevel int) (bool) {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
	
	// Get the path of MFT fteSetupCoordination command. 
	cmdCoordPath, lookPathErr := exec.LookPath("fteSetupCoordination")
	if lookPathErr == nil {
		// Setup coordination configuration
		utils.PrintLog(fmt.Sprintf("Setting up coordination configuration '%s' for agent '%s'",
			gjson.Get(allAgentConfig,"coordinationQMgr.name"), agentNameEnv))
		
		var port string
		var channel string
		if gjson.Get(allAgentConfig,"coordinationQMgr.port").Exists() {
			port = gjson.Get(allAgentConfig,"coordinationQMgr.port").String()
		} else {
			port = "1414"
		}
		
		if gjson.Get(allAgentConfig,"coordinationQMgr.channel").Exists() {
			channel = gjson.Get(allAgentConfig,"coordinationQMgr.channel").String()
		} else {
			channel = "SYSTEM.DEF.SVRCONN"
		}
		cmdSetupCoord := &exec.Cmd {
			Path: cmdCoordPath,
			Args: [] string {cmdCoordPath, 
				"-coordinationQMgr", gjson.Get(allAgentConfig,"coordinationQMgr.name").String(),
				"-coordinationQMgrHost", gjson.Get(allAgentConfig,"coordinationQMgr.host").String(), 
				"-coordinationQMgrPort", port, "-coordinationQMgrChannel", channel, "-f", "-default"},
		}

		// Execute the fteSetupCoordination command. Log an error an exit in case of any error.
		cmdSetupCoord.Stdout = &outb
		cmdSetupCoord.Stderr = &errb
		if err := cmdSetupCoord.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf("fteSetupCoordination command failed. The error is %s", err))
			utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			utils.PrintLog(fmt.Sprintf("Error %s", errb.String()))
		} else {
			if logLevel == LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			}
			utils.PrintLog(fmt.Sprintf ("Coordination setup for '%s' complete", gjson.Get(allAgentConfig,"coordinationQMgr.name").String()))
			// Update coordination properties file with additional attributes specified.
			coordinationPropertiesFile := bfgDataPath + "/mqft/config/" + gjson.Get(allAgentConfig, "coordinationQMgr.name").String() + "/coordination.properties"
			updateProperties(coordinationPropertiesFile, allAgentConfig, "coordinationQMgr.additionalProperties")
			created = true
		}
	} else {
		utils.PrintLog(fmt.Sprintf("fteSetupCoordination command not found. %s", lookPathErr))
	}
	
	return created
}

// Calls fteSetupCommands to create command queue manager configuration.
func setupCommands(allAgentConfig string, bfgDataPath string, agentName string, logLevel int) (bool) {
	// Variables for Stdout and Stderr
	var outb, errb bytes.Buffer
	var created bool = false
	
	// Get the path of MFT fteSetupCommands command. 
	cmdCmdsPath, lookPathErr := exec.LookPath("fteSetupCommands")
	if lookPathErr == nil {
		// Setup commands configuration
		utils.PrintLog(fmt.Sprintf("Setting up commands configuration '%s' for agent '%s'", 
			gjson.Get(allAgentConfig,"coordinationQMgr.name"), agentName))
		var port string
		var channel string
		if gjson.Get(allAgentConfig,"commandQMgr.port").Exists() {
			port = gjson.Get(allAgentConfig,"commandQMgr.port").String()
		} else {
			port = "1414"
		}
		
		if gjson.Get(allAgentConfig,"commandQMgr.channel").Exists() {
			channel = gjson.Get(allAgentConfig,"commandQMgr.channel").String()
		} else {
			channel = "SYSTEM.DEF.SVRCONN"
		}
		
		cmdSetupCmds := &exec.Cmd {
			Path: cmdCmdsPath,
			Args: [] string {cmdCmdsPath, 
				"-p", gjson.Get(allAgentConfig,"coordinationQMgr.name").String(), 
				"-connectionQMgr", gjson.Get(allAgentConfig,"commandQMgr.name").String(), 
				"-connectionQMgrHost", gjson.Get(allAgentConfig,"commandQMgr.host").String(), 
				"-connectionQMgrPort", port, "-connectionQMgrChannel", channel,"-f"},
		}
		  
		cmdSetupCmds.Stdout = &outb
		cmdSetupCmds.Stderr = &errb
		// Execute the fteSetupCommands command. Log an error an exit in case of any error.
		if err := cmdSetupCmds.Run(); err != nil {
			utils.PrintLog(fmt.Sprintf("fteSetupCommands command failed. The errror is %s", err))
			utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			utils.PrintLog(fmt.Sprintf("Error %s", errb.String()))
			os.Exit(1)
		} else {
			if logLevel == LOG_LEVEL_VERBOSE {
				utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
			}
			utils.PrintLog(fmt.Sprintf("Command setup for '%s' is complete", gjson.Get(allAgentConfig,"coordinationQMgr.name").String()))
			
			// Update command properties file with additional attributes specified.
			commandsPropertiesFile := bfgDataPath + "/mqft/config/" + gjson.Get(allAgentConfig, "coordinationQMgr.name").String() + "/command.properties"
			updateProperties(commandsPropertiesFile, allAgentConfig, "commandQMgr.additionalProperties");
			created = true
		}
	} else {
		utils.PrintLog(fmt.Sprintf("fteSetupCommands command not found. %s", lookPathErr))
	}
	
	return created
}

// Read and process protocol bridge server attributes from configuration JSON file
func processBridgeParameters(agentConfig string, params [] string ) ([] string) {
	// Set protocol server type
	serverType := gjson.Get(agentConfig, "protocolBridge.serverType")
	if serverType.Exists() {
		// Use the supplied one.
		params = append(params, "-bt", serverType.String())
	} else {
		// Otherwise use default as FTP
		params = append(params, "-bt", "FTP")
	}
			
	// Set the protocol server host
	serverHost := gjson.Get(agentConfig, "protocolBridge.serverHost")
	if serverHost.Exists(){
		params = append(params, "-bh", serverHost.String())
	} else {
		// Use local host if host name is not specified.
		params = append(params, "-bh", "localhost")
	}
			
	// Set the protocol server timezone, valid only for FTP and FTPS server
	if serverType.String() != "SFTP" {
		serverTimezone := gjson.Get(agentConfig, "protocolBridge.serverTimezone")
		if serverTimezone.Exists() {
			params = append(params, "-btz", serverTimezone.String())
		}
	}
	
	// Set the protocol server platform.
	serverPlatform := gjson.Get(agentConfig, "protocolBridge.serverPlatform")
	if serverPlatform.Exists() {
		params = append(params, "-bm", serverPlatform.String())
	}
			
	// Set the protocol server locale
	if serverType.String() != "SFTP" {
		serverLocale := gjson.Get(agentConfig, "protocolBridge.serverLocale")
		if serverLocale.Exists() {
			params = append(params, "-bsl", serverLocale.String())
		}
	}

	// Set protocol server file encoding
	serverFileEncoding := gjson.Get(agentConfig, "protocolBridge.serverFileEncoding")
	if serverFileEncoding.Exists() {
		params = append(params, "-bfe", serverFileEncoding.String())
	}
			
	// Set the protocol server port
	serverPort := gjson.Get(agentConfig, "protocolBridge.serverPort")
	if serverPort.Exists() {
		params = append(params, "-bp", serverPort.String())
	}
			
	// Set the protocol server trust store file
	serverTrustStoreFile := gjson.Get(agentConfig, "protocolBridge.serverTrustStoreFile")
	if serverTrustStoreFile.Exists () {
		params = append(params, "-bts", serverTrustStoreFile.String())
	}
			
	// Set protocol server limited write flag
	serverLimitedWrite := gjson.Get(agentConfig, "protocolBridge.serverLimitedWrite")
	if serverLimitedWrite.Exists () {
		params = append(params, "-blw", serverLimitedWrite.String())
	}
			
	// Set the protocol server list format.
	serverListFormat := gjson.Get(agentConfig, "protocolBridge.serverListFormat")
	if serverListFormat.Exists () {
		params = append(params, "-blf", serverListFormat.String())
	}
	return params
}

// Output the contents of agent logs to stdout
func mirrorAgentLogs(ctx context.Context, wg *sync.WaitGroup, agentName string, logPathName string) (error){
	mf, err := configureLogger(agentName)
	if err != nil {
		logTermination(err)
		return err
	}

	_, err = mirrorAgentEventLogs(ctx, wg, logPathName, true, mf)
	if err != nil {
		logTermination(err)
		return err
	}
	return nil
}

// Validate attributes in JSON file.
// Check if the configuration JSON contains all required attribtes 
func validateCoordinationAttributes(jsonData string) (error){  
	// Coordination queue manager is mandatory
	if !gjson.Get(jsonData, "coordinationQMgr.name").Exists() {
		err := errors.New("Coordination queue manager name missing. Can't setup agent configuration")
		return err
	}

	// Coordination queue manager host is mandatory
	if !gjson.Get(jsonData, "coordinationQMgr.host").Exists() {
		err := errors.New("Coordination queue manager host name missing. Can't setup agent configuration")
		return err
	}
	
	return nil
}

func validateCommandAttributes(jsonData string) (error){
	// Commands queue manager is mandatory
	if !gjson.Get(jsonData, "commandQMgr.name").Exists() {
		err := errors.New("Command queue manager name missing. Can't setup agent configuration")
		return err
	}
	// Coordination queue manager host is mandatory
	if !gjson.Get(jsonData, "commandQMgr.host").Exists() {
		err := errors.New("Coordination queue manager host name missing. Can't setup agent configuration")
		return err
	}
	
	return nil
}

func validateAgentAttributes(jsonData string) (error) {
	// Agent name is mandatory
	if !gjson.Get(jsonData, "name").Exists() {
		err := errors.New("Agent name missing. Can not setup agent configuration")
		return err
	}

	// Agent queue manager name is mandatory
	if !gjson.Get(jsonData, "qmgrName").Exists() {
		err := errors.New("Agent queue manager name missing. Can not setup agent configuration")
		return err
	}

	// Agent queue manager host is mandatory
	if !gjson.Get(jsonData, "qmgrHost").Exists() {
		err := errors.New("Agent queue manager host name missing. Can not setup agent configuration")
		return err
	}

	return nil
}

// Update agent.properties file with any additional properties specified in
// configuration JSON file.
func updateAgentProperties(propertiesFile string, agentConfig string, sectionName string, bridgeAgent bool) {
	f, err := os.OpenFile(propertiesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		utils.PrintLog(fmt.Sprintf("%v",err))
		return
	}
	defer f.Close()
	
	// Enable logCapture by default. Customer can turn off by specifying it again in config map
	if _, err := f.WriteString("\nlogCapture=true\n"); err != nil {
		utils.PrintLog(fmt.Sprintf("%v", err))
	}
  
	if gjson.Get(agentConfig, sectionName).Exists() {
		result := gjson.Get(agentConfig, sectionName)
		result.ForEach(func(key, value gjson.Result) bool {
		if _, err := f.WriteString("\n" + key.String() + "=" + value.String() + "\n"); err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
		}
		return true // keep iterating
      })
	}

	// If agent credentials file has been specified as environment variable, then set it here
	agentCredPath, agentCredPathSet := os.LookupEnv("MFT_AGENT_CREDENTIAL_FILE")    
	if agentCredPathSet && agentCredPath != "" {
		if _, err := f.WriteString ("\n" + "agentQMgrAuthenticationCredentialsFile=" + agentCredPath + " \n"); err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
		}
	}
  
  // If this is a bridge agent, then configure custom exit
  if bridgeAgent {
	bridgeCredPath, bridgeCredPathSet := os.LookupEnv("MFT_BRIDGE_CREDENTIAL_FILE")    
	if bridgeCredPathSet && bridgeCredPath != "" {
		if _, err := f.WriteString ("\n" + "protocolBridgeCredentialConfiguration=" + bridgeCredPath + " \n"); err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
		}
	}
	
	if _, err := f.WriteString ("\n" + "protocolBridgeCredentialExitClasses=com.ibm.bridgecredentialexit.ProtocolBridgeCustomCredentialExit\n"); err != nil {
		utils.PrintLog(fmt.Sprintf("%v", err))
	}
  } else {
	if _, err := f.WriteString ("\n" + "userSandboxes=true"); err != nil {
		utils.PrintLog(fmt.Sprintf("%v", err))
	}
  }
}

// Update coordination and command properties file with any additional properties specified in
// configuration JSON file.
func updateProperties(propertiesFile string, agentConfig string, sectionName string) {
	f, err := os.OpenFile(propertiesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		utils.PrintLog(fmt.Sprintf("%v", err))
		return
	}
	defer f.Close()
  
	if gjson.Get(agentConfig, sectionName).Exists() {
		result := gjson.Get(agentConfig, sectionName)
		result.ForEach(func(key, value gjson.Result) bool {
		if _, err := f.WriteString("\n" + key.String() + "=" + value.String() + "\n"); err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
		}
		return true // keep iterating
      })
  }
}

// Returns the contents of the specified file.
func readFileContents(propertiesFile string, logLevel int) (string) {
	// Open our xmlFile
    bridgeProperiesXmlFile, err := os.Open(propertiesFile)
	// if we os.Open returns an error then handle it
    if err != nil {
		if logLevel == LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf("%v", err))
			return ""
		}
    }

    // defer the closing of our xml file so that we can parse it later on
    defer bridgeProperiesXmlFile.Close()

	xmlData,_ := ioutil.ReadAll(bridgeProperiesXmlFile)
	xmlText := string(xmlData)
	return xmlText
}

// Updates ProtocolBridgeProperties file with specified additional attributes
func updateProtocolBridgePropertiesFile(propertiesFile string, agentConfig string, sectionName string, logLevel int) {
	// First read the entire contents of the ProtocolBridgeProperties file.
	bridgeProperites := readFileContents(propertiesFile, logLevel)
	if len(bridgeProperites) > 0 {
		// Find the last index of ProtocolBridgeProperties.xsd in the file contents. 
		// We will be inserting new attributes of that "ProtocolBridgeProperties.xsd>"
		lastIndex := strings.LastIndex(bridgeProperites,"ProtocolBridgeProperties.xsd")
		// Add 30 to index because of the length of "ProtocolBridgeProperties.xsd>" is 30.
		insertIndex := lastIndex + 30

		f, err := os.OpenFile(propertiesFile, os.O_WRONLY, 0644)
		if err != nil {
			utils.PrintLog(fmt.Sprintf("%v", err))
			return
		}
		defer f.Close()  
		
		if gjson.Get(agentConfig, sectionName).Exists() {
			result := gjson.Get(agentConfig, sectionName)
			result.ForEach(func(key, value gjson.Result) bool {
				// We support only two properties that can be set from config file. Others are ignored.
				if strings.EqualFold(key.String(),"credentialsFile") == true {
					insertString := "\n<tns:credentialsFile path=\"" + value.String() + "\" />"
					bridgeProperites = bridgeProperites[:insertIndex] + insertString + bridgeProperites[insertIndex:]
					insertIndex += len(insertString)
				} else if strings.EqualFold(key.String(),"credentialsKeyFile") == true {
					//<tns:credentialsKeyFile path="c:\temp\agentinitkey.key"/>
					insertString := "\n<tns:credentialsKeyFile path=\"" + value.String() + "\" />"
					bridgeProperites = bridgeProperites[:insertIndex] + insertString + bridgeProperites[insertIndex:]
					insertIndex += len(insertString)
				}
				return true // go on till we insert all properties
			})
		}

		// Print the updated contents of the ProtocolBridgeProperties.xml file
		if logLevel == LOG_LEVEL_VERBOSE {
			utils.PrintLog(bridgeProperites)
		}
		// Write the updated properties to file.
		_, writeErr := f.WriteString(bridgeProperites)
		if writeErr != nil {
			utils.PrintLog(fmt.Sprintf("%v", writeErr))
		}
	}
}

// Unregister and delete agent 
func deleteAgent (coordinationQMgr string, agentName string) (error) {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf("Deleting configuration for agent '%s'", agentName))
  
	// Get the path of MFT fteDeleteAgent command. 
	cmdDltAgentPath, lookErr:= exec.LookPath("fteDeleteAgent")
	if lookErr != nil {
		return lookErr
	}
	// -f force option is not used so that monitor is not recreated if it already exists.
	cmdDltAgentCmd := &exec.Cmd {
		Path: cmdDltAgentPath,
		Args: [] string {cmdDltAgentPath, "-p", coordinationQMgr, "-f", agentName},
	}

	// Reuse the same buffer
	cmdDltAgentCmd.Stdout = &outb
	cmdDltAgentCmd.Stderr = &errb
	// Execute the fteDeleteAgent command. Log an error an exit in case of any error.
	if err := cmdDltAgentCmd.Run(); err != nil {
		utils.PrintLog(fmt.Sprintf("fteDeleteAgent command failed. The errror is: %s", err))
		utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
		utils.PrintLog(fmt.Sprintf("Error %s", errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
	} else {
		utils.PrintLog(fmt.Sprintf("Configuration for agent '%s' has been deleted", agentName))
	}
	return nil
}

// Clean agent before starting it.
func cleanAgent (agentConfig string, coordinationQMgr string, agentName string, logLevel int){
	cleanOnStart := gjson.Get(agentConfig, "cleanOnStart")
	if cleanOnStart.Exists() {
		cleanItem := cleanOnStart.String()
		if cleanItem == "transfers" {
			cleanAgentItem (coordinationQMgr, agentName, cleanItem, "-trs", logLevel)
		} else if cleanItem == "monitors" {
			cleanAgentItem (coordinationQMgr, agentName, cleanItem, "-ms", logLevel)
		} else if cleanItem == "scheduledTransfers" {
			cleanAgentItem (coordinationQMgr, agentName, cleanItem, "-ss", logLevel)
		} else if cleanItem == "invalidMessages" {
			cleanAgentItem (coordinationQMgr, agentName, cleanItem, "-ims", logLevel)
		} else if cleanItem == "all" {
			cleanAgentItem (coordinationQMgr, agentName, cleanItem, "-all", logLevel)
		} else {
			utils.PrintLog(fmt.Sprintf("Invalid value '%s' specified for cleanOnStart attribte has been ignored", cleanItem))
		}
	}
}

// Clean agent on start of container.
func cleanAgentItem (coordinationQMgr string, agentName string, item string, option string, logLevel int) (error) {
	var outb, errb bytes.Buffer
	utils.PrintLog(fmt.Sprintf("Cleaning %s of agent %s", item, agentName))
  
	// Get the path of MFT fteCleanAgent command. 
	cmdCleanAgentPath, lookErr:= exec.LookPath("fteCleanAgent")
	if lookErr != nil {
		return lookErr
	}
	
	// -f force option is not used so that monitor is not recreated if it already exists.
	cmdCleanAgentCmd := &exec.Cmd {
		Path: cmdCleanAgentPath,
		Args: [] string {cmdCleanAgentPath, "-p", coordinationQMgr, option, agentName},
	}

	// Reuse the same buffer
	cmdCleanAgentCmd.Stdout = &outb
	cmdCleanAgentCmd.Stderr = &errb
	// Execute the fteCleanAgent command. Log an error an exit in case of any error.
	if err := cmdCleanAgentCmd.Run(); err != nil {
		utils.PrintLog(fmt.Sprintf("fteCleanAgent command failed. The errror is: %s", err))
		utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
		utils.PrintLog(fmt.Sprintf("Error %s", errb.String()))
		// Return no error even if we fail to create monitor. We have output the
		// information to console.
	} else {
		if logLevel == LOG_LEVEL_VERBOSE {
			utils.PrintLog(fmt.Sprintf("Command: %s", outb.String()))
		} else {
			utils.PrintLog("Agent cleaned")
		}
	}
	return nil
}

// Setup userSandBox configuration to restrict access to file system
func createUserSandbox(sandboxXmlFileName string) {
	// Open existing UserSandboxes.xml file
	userSandBoxXmlFile, err := os.OpenFile(sandboxXmlFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	// if we os.Open returns an error then handle it
    if err != nil {
		utils.PrintLog(fmt.Sprintf("%v", err))
		return
    }

    // defer the closing of our xml file so that we can parse it later on
    defer userSandBoxXmlFile.Close()
	
	// Set sandBoxRoot for standard agents to restrict the file system access. Use the value specified in 
	// MFT_MOUNT_PATH environment variable if available else use the default "/mountpath" folder.
	// Agent will be able to read from or write to this folder and it will not have access to other parts
	// of the file system.
	mountPathEnv := os.Getenv("MFT_MOUNT_PATH")
	var mountPath string
	if len(mountPathEnv) > 0 {
		//If the supplied path does not have /** suffix, then add it
		if !strings.HasSuffix(mountPathEnv, "/**") {
			if strings.HasSuffix(mountPathEnv, "/*") {
				mountPath = mountPathEnv + "*"
			} else if strings.HasSuffix(mountPathEnv, "/") {
				mountPath = mountPathEnv + "**"
			} else {
				mountPath = mountPathEnv + "/**"
			}
		} else {
			mountPath = mountPathEnv
		}
	} else {
		// No environment variable specified. So use fixed path.
		mountPath = MOUNT_PATH_FOR_TRANSFERS
	}

    // Write a generic 
	var sandboxXmlText string
	sandboxXmlText =  "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	sandboxXmlText += "<tns:userSandboxes\n"
	sandboxXmlText += "         xmlns:tns=\"http://wmqfte.ibm.com/UserSandboxes\"\n"
	sandboxXmlText += "         xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n"
	sandboxXmlText += "         xsi:schemaLocation=\"http://wmqfte.ibm.com/UserSandboxes UserSandboxes.xsd\">\n\n"
	sandboxXmlText += "    <tns:agent>\n"
	sandboxXmlText += "         <tns:sandbox user=\"^[a-zA-Z0-9]*$\" userPattern=\"regex\">\n"
	sandboxXmlText += "              <tns:read>\n"
    sandboxXmlText += "       	          <tns:include name=\"" + mountPath + "\"/>\n"
    sandboxXmlText += "	                  <tns:include name=\"**\" type=\"queue\"/>\n"
	sandboxXmlText += "              </tns:read>\n"
	sandboxXmlText += "              <tns:write>\n"
    sandboxXmlText += "	                  <tns:include name=\"" + mountPath + "\"/>\n"
	sandboxXmlText += "                   <tns:include name=\"**\" type=\"queue\"/>\n"
	sandboxXmlText += "              </tns:write>\n"
	sandboxXmlText += "        </tns:sandbox>\n"
	sandboxXmlText += "     </tns:agent>\n"
	sandboxXmlText += "</tns:userSandboxes>"
	
	// Write the updated properties to file.
	_, writeErr := userSandBoxXmlFile.WriteString(sandboxXmlText)
	if writeErr != nil {
		utils.PrintLog(fmt.Sprintf("%v", writeErr))
	}
}