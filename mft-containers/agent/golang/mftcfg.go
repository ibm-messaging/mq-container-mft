package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"bytes"
	"os/signal"
	"syscall"
	"time"
	"strconv"
	"bufio"
    "encoding/json"
    "io/ioutil"
	"container/list"
)

type CoordinationQMgr struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int `json:"port"`
	Channel string `json:"channel"`
}

type CommandsQMgr struct {
	Name string `json:"name"`
	Host string `json:"host"`
	Port int `json:"port"`
	Channel string `json:"channel"`
}

type Agent struct {
	Name string `json:"name"`
	QMgr string `json:"qmgr"`
	QMgrHost string `json:"qmgrHost"`
	QMgrPort int `json:"qmgrPort"`
	QMgrChannel string `json:"qmgrChannel"`
	CredentialsFile string `json:"credentialsFile"`
}

type ProtocolBridge struct {
    CredentialsFile string `json:"credentialsFile"`
    ServerType string `json:"serverType"`
    ServerHost string `json:"serverHost"`
    ServerTimezone string `json:"serverTimezone"`
    ServerPlatform string `json:"serverPlatform"`
    ServerLocale string `json:"serverLocale"`
    ServerFileEncoding string `json:"serverFileEncoding"`
    ServerPort int `json:"serverPort"`
    ServerTrustStoreFile string `json:"serverTrustStoreFile"`
    ServerLimitedWrite string `json:"serverLimitedWrite"`
    ServerListFormat string `json:"serverListFormat"`
    ServerUserId string `json:"serverUserId"`
    ServerPassword string `json:"serverPassword"`
}

type AgentConfiguration struct {
  DataPath string `json:"dataPath"`
  MonitorInterval int `json:"monitoringInterval"`
  DisplayAgentLogs bool `json:"displayAgentLogs"`
  DisplayLines int `json:"displayLineCount"`
  AgentType string `json:"agentType"`
  CoordQMgr CoordinationQMgr `json:"coordinationQMgr"`
  CmdsQMgr CommandsQMgr `json:"commandsQMgr"`
  Agent Agent `json:"agent"`
  ProtocolBridge ProtocolBridge `json:"protocolBridge"`
}

// Main entry point to program.
func main () {
  var bfgDataPath string
  var bfgConfigFilePath string 
  var sleepTime int
  var agentConfig AgentConfiguration
  var agentMonitorIntervalStr  string
  var e error
  var showAgentLogs bool
  var displayLines int
  // Variables for Stdout and Stderr
  var outb, errb bytes.Buffer

  // 1- ENV_BFG_DATA 
  // 2- ENV_AGENT_CONFIG_FILE
  // 3- ENV_MONITOR_INTERVAL
  // 4- ENV_SHOW_LOGS
  // 5- ENV_TOTAL_LINES
  if len(os.Args) == 2 {
	// Configuration file path from environment variable
    bfgConfigFilePath = os.Args[1]
	// Read agent configuration data from JSON file.
    agentConfig, e = ReadConfigurationDataFromFile(bfgConfigFilePath)
	// Exit if we had any error when reading configuration file
    if e != nil {
      panic(e)
      return
    }
	
    // BFG_DATA path
    bfgDataPath = agentConfig.DataPath
	// Agent liveliness monitoring interval
	sleepTime = agentConfig.MonitorInterval
	// To display agent logs or not.
	showAgentLogs = agentConfig.DisplayAgentLogs
	// Display n number of logs from agent log
	displayLines = agentConfig.DisplayLines	
  } else {
    // We don't have configuration file specified, instead we have environment
	// variables specified.
	// BFG_DATA path.
    bfgDataPath = os.Args[1]
    agentMonitorIntervalStr =  os.Args[2]
    // Determine the sleep time.
	sleepTimeInt, err := strconv.Atoi(agentMonitorIntervalStr)
	if err != nil {
	  // There was some error when getting the sleep time. Assume 300 seconds
	  // as the default sleep time
	  sleepTime = 300
	} else {
	  sleepTime = sleepTimeInt
	}
	// To display agent logs or not. 
	if strings.EqualFold(os.Args[3],"YES") == true {
      showAgentLogs = true
	}
	
	displayLineCount, err := strconv.Atoi(os.Args[4])
	if err != nil {
	  // There was some error when getting the sleep time. Assume 300 seconds
	  // as the default sleep time
	  displayLines = 40
	} else {
	  displayLines = displayLineCount
	}
	
    // Read rest of the environment variables and construct an instance 
	// of AgentConfiguration structure.
	agentConfig = ReadConfigurationDataFromEnvironment(os.Args)
  }

  // Set BFG_DATA environment variable so that we can run MFT commands.
  os.Setenv("BFG_DATA", bfgDataPath)

  // Get the path of MFT fteSetupCoordination command. 
  cmdCoordPath, lookErr := exec.LookPath("fteSetupCoordination")
  if lookErr != nil {
    panic(lookErr)
    return
  }

  // Get the path of MFT fteSetupCommands command. 
  cmdCmdsPath, lookErr := exec.LookPath("fteSetupCommands")
  if lookErr != nil {
    panic(lookErr)
    return
  }

  // Get the path of MFT fteCreateAgent command. 
  cmdCrtAgntPath, lookErr:= exec.LookPath("fteCreateAgent")
  if lookErr != nil {
	panic(lookErr)
	return
  }

  // Get the path of MFT fteCreateBridgeAgent command
  cmdCrtBridgeAgntPath, lookErr :=exec.LookPath("fteCreateBridgeAgent")
  if lookErr != nil {
    panic(lookErr)
    return
  }

  // Get the path of MFT fteStartAgent command. 
  cmdStrAgntPath, lookErr:= exec.LookPath("fteStartAgent")
  if lookErr != nil {
    panic(lookErr)
    return
  }

  // Get the path of MFT fteStopAgent command. 
  cmdStopAgntPath, lookErr:= exec.LookPath("fteStopAgent")
  if lookErr != nil {
    panic(lookErr)
    return
  }

  // Get the path of MFT ftePingAgent command. 
  cmdPingAgntPath, lookErr:= exec.LookPath("ftePingAgent")
  if lookErr != nil {
    panic(lookErr)
    return
  }

  // Setup coordination configuration
  fmt.Printf("Setting up coordination configuration %s for agent %s\n", agentConfig.CoordQMgr.Name, agentConfig.Agent.Name)
  cmdSetupCoord := &exec.Cmd {
	Path: cmdCoordPath,
	Args: [] string {cmdCoordPath, "-coordinationQMgr", agentConfig.CoordQMgr.Name, "-coordinationQMgrHost", agentConfig.CoordQMgr.Host, 
	                               "-coordinationQMgrPort",strconv.Itoa(agentConfig.CoordQMgr.Port), "-coordinationQMgrChannel", agentConfig.CoordQMgr.Channel, "-f"},
  }

  // Execute the fteSetupCoordination command. Log an error an exit in case of any error.
  cmdSetupCoord.Stdout = &outb
  cmdSetupCoord.Stderr = &errb
  if err := cmdSetupCoord.Run(); err != nil {
	fmt.Println("fteSetupCoordination command failed. The error is: ", err);
	return
  }

  // Setup commands configuration
  fmt.Printf("Setting up commands configuration %s for agent %s\n", agentConfig.CoordQMgr.Name, agentConfig.Agent.Name)
  cmdSetupCmds := &exec.Cmd {
	Path: cmdCmdsPath,
	Args: [] string {cmdCmdsPath, "-p", agentConfig.CoordQMgr.Name, "-connectionQMgr", agentConfig.CmdsQMgr.Name, "-connectionQMgrHost", agentConfig.CmdsQMgr.Host, 
	                              "-connectionQMgrPort", strconv.Itoa(agentConfig.CmdsQMgr.Port), "-connectionQMgrChannel", agentConfig.CmdsQMgr.Channel,"-f"},
  }
  
  // Reuse the same buffer
  outb.Reset()
  errb.Reset()
  cmdSetupCmds.Stdout = &outb
  cmdSetupCmds.Stderr = &errb
  // Execute the fteSetupCommands command. Log an error an exit in case of any error.
  if err := cmdSetupCmds.Run(); err != nil {
	fmt.Println("fteSetupCommands command failed. The errror is: ", err);
	return
  }

  // Create agent.
  fmt.Printf("Creating %s agent with name %s \n", agentConfig.AgentType, agentConfig.Agent.Name)
  var cmdCrtAgnt * exec.Cmd
  if strings.EqualFold(agentConfig.AgentType, "STANDARD") == true {
    cmdCrtAgnt = &exec.Cmd {
	Path: cmdCrtAgntPath,
	Args: [] string {cmdCrtAgntPath, "-p", agentConfig.CoordQMgr.Name, "-agentName", agentConfig.Agent.Name, "-agentQMgr", agentConfig.Agent.QMgr, 
	                                 "-agentQMgrHost", agentConfig.Agent.QMgrHost, "-agentQMgrPort", strconv.Itoa(agentConfig.Agent.QMgrPort), "-agentQMgrChannel", agentConfig.Agent.QMgrChannel,
									 "-credentialsFile",agentConfig.Agent.CredentialsFile, "-f"},
    }
  } else {
    var  params [] string 
    params = append(params,cmdCrtBridgeAgntPath,  "-p", agentConfig.CoordQMgr.Name, "-agentName", agentConfig.Agent.Name, "-agentQMgr", agentConfig.Agent.QMgr,
                                         "-agentQMgrHost", agentConfig.Agent.QMgrHost, "-agentQMgrPort", strconv.Itoa(agentConfig.Agent.QMgrPort), "-agentQMgrChannel", agentConfig.Agent.QMgrChannel,
                                         "-credentialsFile",agentConfig.Agent.CredentialsFile, "-f")

    if agentConfig.ProtocolBridge.ServerType != "" {
	  params = append(params,"-bt", agentConfig.ProtocolBridge.ServerType)
    } else {
	  params = append(params, "-bt", "FTP")
    }

    if agentConfig.ProtocolBridge.ServerHost != "" {
	  params = append(params,"-bh", agentConfig.ProtocolBridge.ServerHost)
    } else {
	  params = append(params,"-bh", "localhost")
    }

    if agentConfig.ProtocolBridge.ServerTimezone != "" {
	  params = append(params,"-btz", agentConfig.ProtocolBridge.ServerTimezone)
    }

    if agentConfig.ProtocolBridge.ServerPlatform != "" {
	  params = append(params,"-bm", agentConfig.ProtocolBridge.ServerPlatform)
    }

    if agentConfig.ProtocolBridge.ServerType != "SFTP" &&  agentConfig.ProtocolBridge.ServerLocale != "" {
	  params = append(params,"-bsl", agentConfig.ProtocolBridge.ServerLocale)
    }

    if agentConfig.ProtocolBridge.ServerFileEncoding != "" {
	  params = append(params,"-bfe", agentConfig.ProtocolBridge.ServerFileEncoding)
    }

    if agentConfig.ProtocolBridge.ServerPort != 0 {
	  params = append(params,"-bp", strconv.Itoa(agentConfig.ProtocolBridge.ServerPort))
    }

    if agentConfig.ProtocolBridge.ServerTrustStoreFile != "" {
	  params = append(params,"-bts", agentConfig.ProtocolBridge.ServerTrustStoreFile )
    }

    if agentConfig.ProtocolBridge.ServerLimitedWrite != "" {
	  params = append(params,"-blw", agentConfig.ProtocolBridge.ServerLimitedWrite)
    }

    if agentConfig.ProtocolBridge.ServerListFormat != "" {
	  params = append(params,"-blf", agentConfig.ProtocolBridge.ServerListFormat)
    }

    cmdCrtAgnt = &exec.Cmd {
        Path: cmdCrtBridgeAgntPath,
        Args: params,
    }
  }

  // Reuse the same buffer
  outb.Reset()
  errb.Reset()
  cmdCrtAgnt.Stdout = &outb
  cmdCrtAgnt.Stderr = &errb
  // Execute the fteCreateAgent command. Log an error an exit in case of any error.
  if err := cmdCrtAgnt.Run(); err != nil {
	fmt.Println("Create Agent command failed. The error is: ", err);
	return
  }

  fmt.Printf("Starting agent %s\n", agentConfig.Agent.Name)
  cmdStrAgnt := &exec.Cmd {
	Path: cmdStrAgntPath,
	Args: [] string {cmdStrAgntPath,"-p", agentConfig.CoordQMgr.Name, agentConfig.Agent.Name},
  }
  
  // Reuse the same buffer
  outb.Reset()
  errb.Reset()
  cmdStrAgnt.Stdout = &outb
  cmdStrAgnt.Stderr = &errb
  // Run fteStartAgent command. Log an error and exit in case of any error.
  if err := cmdStrAgnt.Run(); err != nil {
	fmt.Println("Error:", err);
	return
  }

  fmt.Printf("Verifying status of agent %s\n", agentConfig.Agent.Name)
  cmdListAgentPath, lookErr := exec.LookPath("fteListAgents")
  if lookErr != nil {
    panic(lookErr)
    return
  }
 
  // Prepare fteListAgents command for execution
  cmdListAgents := &exec.Cmd {
	Path: cmdListAgentPath,
	Args: [] string {cmdListAgentPath, "-p", agentConfig.CoordQMgr.Name, agentConfig.Agent.Name},
  }

  // Reuse the same buffer
  outb.Reset()
  errb.Reset()
  cmdListAgents.Stdout = &outb
  cmdListAgents.Stderr = &errb
  // Execute and get the output of the command into a byte buffer.
  err := cmdListAgents.Run()
  if err != nil {
	fmt.Println("Error: ", err)
	return
  }

  // Now parse the output of fteListAgents command and take appropriate actions.
  var agentStatus string 
  agentStatus = outb.String()

  // Create a go routine to read the agent output0.log file
  if showAgentLogs == true {
    agentLogPath := bfgDataPath + "/mqft/logs/" + agentConfig.CoordQMgr.Name + "/agents/" + agentConfig.Agent.Name + "/logs/output0.log"
    DisplayAgentOutputLog(displayLines, agentLogPath)
  }
 
  if strings.Contains(agentStatus,"STOPPED") == true {
    //if agent status is still stopped, wait for some time and then reissue fteListAgents commad
    fmt.Println("Agent status not started yet. Wait for 5 seconds and recheck status again")
    time.Sleep(5 * time.Second)
	
    // Prepare fteListAgents command for execution
    cmdListAgents := &exec.Cmd {
	  Path: cmdListAgentPath,
	  Args: [] string {cmdListAgentPath, "-p", agentConfig.CoordQMgr.Name, agentConfig.Agent.Name},
    }
    
	// Execute and get the output of the command into a byte buffer.
    outb.Reset()
    errb.Reset()
    cmdListAgents.Stdout = &outb
    cmdListAgents.Stderr = &errb
    err := cmdListAgents.Run()
    if err != nil {
      fmt.Println("Error: ", err)
      return
    }
    // Copy the latest status again.	
	agentStatus = outb.String()
  } // If agent stopped

  // If agent status is READY, then we are good. 
  if strings.Contains(agentStatus,"READY") == true  {
	// Agent is READY, so start monitoring the status. If the status becomes unknown, 
	// this monitoring program terminates thus container also ends.
	fmt.Println("Agent has started. Starting to monitor status")
    // Setup channel to handle signals to stop agent
	sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)
    // Notify monitor program when SIGINT or SIGTERM is issued to container.
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    var stopAgent bool
    // Handler for receiving singals 
    go func() {
      sig := <-sigs
	  fmt.Printf("Received signal %s\n", sig)
	  stopAgent = true
      done <- true
    }()

	// Loop for ever or till asked to stop
	for {
	  if stopAgent {
        fmt.Printf("Stopping agent %s\n", agentConfig.Agent.Name)
	    cmdStopAgnt := &exec.Cmd {
	      Path: cmdStopAgntPath,
	      Args: [] string {cmdStopAgntPath,"-p", agentConfig.CoordQMgr.Name, agentConfig.Agent.Name, "-i"},
	    }
        
		outb.Reset()
        errb.Reset()
        cmdStopAgnt.Stdout = &outb
        cmdStopAgnt.Stderr = &errb
	    err := cmdStopAgnt.Run()
        if err != nil {
	      fmt.Println("An error occured when running fteStopAgent command. The error is: ", err)
        }
	    return
      } // End of stopAgent processing

      // Keep running fteListAgents at specified interval.
	  cmdListAgents := &exec.Cmd {
        Path: cmdListAgentPath,
        Args: [] string {cmdListAgentPath, "-p", agentConfig.CoordQMgr.Name, agentConfig.Agent.Name},
      }

      outb.Reset()
      errb.Reset()
	  cmdListAgents.Stdout = &outb
      cmdListAgents.Stderr = &errb
      err := cmdListAgents.Run()
      if err != nil {
        fmt.Printf("An error occurred when running fteListAgents command. The error is %s\n: ", err)
        return
      }

      // Check if the status of agent is UNKNOWN. If it is run ftePingAgent
	  // to see if the agent is responding. If does not, then stop container.
      var agentStatus string
      agentStatus = outb.String()  
      if strings.Contains(agentStatus,"UNKNOWN") {
        fmt.Println("Agent status unknown. Pinging the agent")
        cmdPingAgent := &exec.Cmd {
	      Path: cmdPingAgntPath,
	      Args: [] string {cmdPingAgntPath, "-p", agentConfig.CmdsQMgr.Name, agentConfig.Agent.Name},
        }
 
        outb.Reset()
        errb.Reset()
	    cmdPingAgent.Stdout = &outb
        cmdPingAgent.Stderr = &errb
	    err := cmdPingAgent.Run()
          if err != nil {
            fmt.Println("An error occurred when running ftePingAgent command. The error is: ", err)
            return
          }
	    
		var pingStatus string
	    pingStatus = outb.String()
	    if strings.Contains(pingStatus, "BFGCL0214I") {
	      fmt.Printf("Agent %s did not respond to ping. Monitor exiting\n", agentConfig.Agent.Name)
	      return
	    }
      } else {
	    // Agent is alive, Then sleep for specified time
	    time.Sleep(time.Duration(sleepTime) * time.Millisecond)
      }
    } // For loop.
  } else {
    fmt.Println("Agent not started. Quitting")
    return
  }
}

// Method to display agent logs from output0.log file
func DisplayAgentOutputLog(displayLines int, agentLogPath string) {
  // A channel to display logs continuosly.
  go func() {
    f, err := os.Open(agentLogPath)
	defer f.Close()
    if err != nil {
      fmt.Printf("error opening file: %v\n",err)
      return
    }

	fmt.Println("=======================================================================")
	fmt.Println("============================= Agent logs ==============================")
	fmt.Println("=======================================================================")
    logFileLines := list.New()
    r := bufio.NewReader(f)
    for {
      s, e := Readln(r)
	  if e == nil {
		logFileLines.PushBack(s)
	    if logFileLines.Len() == displayLines {
		  element := logFileLines.Front()
		  logFileLines.Remove(element)
	    }
	  } else {
		break
	  }
    }
      
	for element := logFileLines.Front(); element != nil; element = element.Next() {
	  fmt.Println(element.Value)
    }
	  
	for {
      s, e := Readln(r)
      for e == nil {
        fmt.Println(s)
        s,e = Readln(r)
      }
	}
  }()
}

// Method to read a line from agents output0.log file
func Readln(r *bufio.Reader) (string, error) {
  var (isPrefix bool = true
       err error = nil
       line, ln []byte
      )
  for isPrefix && err == nil {
      line, isPrefix, err = r.ReadLine()
      ln = append(ln, line...)
  }
  return string(ln),err
}

// Read configuration data from json file
func ReadConfigurationDataFromFile(configFile string) (AgentConfiguration, error ) {
  var agentConfig AgentConfiguration
  jsonFile, err := os.Open(configFile)
  //agentConfig = nil

  // if we os.Open returns an error then handle it
  if err != nil {
    fmt.Println(err)
	return agentConfig, err
  }
  
  fmt.Println("Successfully Opened file " + configFile)
  // defer the closing of our jsonFile so that we can parse it later on
  defer jsonFile.Close()
  
  // read file
  var data []byte
  data, err = ioutil.ReadAll(jsonFile)
  if err != nil {
     fmt.Print(err)
	 return agentConfig, err
  }

     // unmarshall it
  err = json.Unmarshal(data, &agentConfig)
  if err != nil {
     fmt.Println("error:", err)
	 return agentConfig, err
  }
  return agentConfig, err
}

func ReadConfigurationDataFromEnvironment(envs [] string) (AgentConfiguration){
  var agentConfig AgentConfiguration
  agentConfig.CoordQMgr.Name = envs[3]
  agentConfig.CoordQMgr.Host  = envs[4]
  //agentConfig.CoordQMgr.Port  = strconv.Atoi(envs[5])
  agentConfig.CoordQMgr.Channel = envs[6]
  agentConfig.CmdsQMgr.Name  = envs[7]
  agentConfig.CmdsQMgr.Host  = envs[8]
  //agentConfig.CmdsQMgr.Port  = strconv.Atoi(envs[9])
  agentConfig.CmdsQMgr.Channel  = envs[10]
  agentConfig.Agent.Name  = envs[11]
  agentConfig.Agent.QMgr  = envs[12]
  agentConfig.Agent.QMgrHost  = envs[13]
  //agentConfig.Agent.QMgrPort  =strconv.Atoi(envs[14])
  agentConfig.Agent.QMgrChannel  = envs[15]
  agentConfig.Agent.CredentialsFile  = envs[16]
  return agentConfig
}

