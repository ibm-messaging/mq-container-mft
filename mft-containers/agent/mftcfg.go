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
	"encoding/json"
)

func main () {
 fmt.Printf("Input parameters: %s %s %s %s %s %s\n", os.Args[0], os.Args[1])
 
 if os.Args[1] == nil {
   fmt.Println ("Agent configuration file not found. Exiting.")
   return
 }
 data, err := ioutil.ReadFile(os.Args[1])
    if err != nil {
      fmt.Print(err)
    }

	type agentConfiguration struct {
	    bfgdata string 
		coordinationQMgr string
		coordinationQMgrHost string
		coordinationQMgrPort string
		coordinationQMgrChannel string
		connectionQMgr string
		connectionQMgrHost string
		connectionQMgrPort string
		connectionQMgrChannel string
		agentQMgr  string
		agentQMgrHost string
		agentQMgrPort string
		agentQMgrChannel string
		credentialsFile  string
		bt string //${PROTOCOL_FILE_SERVER_TYPE} 
		bh string //${SERVER_HOST_NAME}
		btz string //${SERVER_TIME_ZONE} 
		bm string // ${SERVER_PLATFORM} 
		bsl string // ${SERVER_LOCALE} 
		bfe string // ${SERVER_FILE_ENCODING} 
    }

    var configData agentConfiguration

    // unmarshall it
    err = json.Unmarshal(data, &configData)
    if err != nil {
        fmt.Println("error:", err)
    }
	
 // Set BFG_DATA environment variable
 os.Setenv("BFG_DATA",configData.bfgdata)

 cmdCoordPath, lookErr := exec.LookPath("fteSetupCoordination")
 if lookErr != nil {
    panic(lookErr)
    return
 }

 cmdCmdsPath, lookErr := exec.LookPath("fteSetupCommands")
 if lookErr != nil {
    panic(lookErr)
    return
 }

 cmdCrtAgntPath, lookErr:= exec.LookPath("fteCreateAgent")
 if lookErr != nil {
	panic(lookErr)
	return
 }

 cmdStrAgntPath, lookErr:= exec.LookPath("fteStartAgent")
 if lookErr != nil {
    panic(lookErr)
    return
 }

 cmdStopAgntPath, lookErr:= exec.LookPath("fteStopAgent")
 if lookErr != nil {
    panic(lookErr)
    return
 }

 fmt.Printf("Setting up coordination configuration %s for agent %s\n", configData.coordinationQMgr, configData.agentName)
 cmdSetupCoord := &exec.Cmd {
	Path: cmdCoordPath,
	Args: [] string {cmdCoordPath, "-coordinationQMgr", configData.coordinationQMgr, "-coordinationQMgrHost", configData.coordinationQMgrHost, 
	                 "-coordinationQMgrPort", configData.coordinationQMgrPort, "-coordinationQMgrChannel", configData.coordinationQMgrChannel, "-f"},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }

 if err := cmdSetupCoord.Run(); err != nil {
	fmt.Println("Error:", err);
	return
 }

 fmt.Printf("Setting up command  configuration %s for agent %s\n", configData.connectionQMgr, configData.agentName)
 cmdCmds := &exec.Cmd {
	Path: cmdCmdsPath,
	Args: [] string {cmdCmdsPath, "-p", configData.coordinationQMgr, -connectionQMgr", configData.connectionQMgr, "-connectionQMgrHost", configData.connectionQMgrHost, 
	                 "-connectionQMgrPort", configData.connectionQMgrPort, "-connectionQMgrChannel", configData.connectionQMgrChannel,"-f"},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }

 if err := cmdCmds.Run(); err != nil {
	fmt.Println("Error:", err);
	return
 }

 fmt.Printf("Creating agent %s\n", configData.agentName)
 cmdCrtAgnt := &exec.Cmd {
	Path: cmdCrtAgntPath,
	Args: [] string {cmdCrtAgntPath,"-p", configData.coordinationQMgr, "-agentName", configData.agentName, "-agentQMgr", configData.agentQMgr, "-agentQMgrHost", configData.agentQMgrHost, 
	                 "-agentQMgrPort", configData.agentQMgrPort, "-agentQMgrChannel", configData.agentQMgrChannel,"-credentialsFile", configData.credentialsFile, "-f"},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }

 if err := cmdCrtAgnt.Run(); err != nil {
	fmt.Println("Error:", err);
	return
 }

 fmt.Printf("Starting agent %s\n", os.Args[5])
 cmdStrAgnt := &exec.Cmd {
	Path: cmdStrAgntPath,
	Args: [] string {cmdStrAgntPath,"-p", configData.coordinationQMgr, configData.agentName},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }

 if err := cmdStrAgnt.Run(); err != nil {
	fmt.Println("Error:", err);
	return
 }

 fmt.Printf("Verifying status of agent %s\n", configData.agentName)
 listAgentPath, lookErr := exec.LookPath("fteListAgents")
 if lookErr != nil {
    panic(lookErr)
    return
 }

 var outb, errb bytes.Buffer
 cmdListAgents := &exec.Cmd {
	Path: listAgentPath,
	Args: [] string {listAgentPath, "-p", configData.coordinationQMgr, configData.agentName},
 }
 cmdListAgents.Stdout = &outb
 cmdListAgents.Stderr = &errb

 err := cmdListAgents.Run()
 if err != nil {
	fmt.Println("Error: ", err)
	return
 }

  var agentStatus string 
  agentStatus = outb.String()
  
  //fmt.Printf("Agent stauts %s\n", agentStatus)
  
  if strings.Contains(agentStatus,"READY") == true  {
	// Agent is READY
	fmt.Println("Agent has started. Starting to monitor status")
        // Setup channel to handle signals to stop agent
	sigs := make(chan os.Signal, 1)
        done := make(chan bool, 1)

        signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
        var stopAgent bool

        go func() {
          sig := <-sigs
	  //fmt.Println("Received singal")
          fmt.Println(sig)
	  stopAgent = true
          done <- true
        }()

	// Loop for ever or till asked to stop
	for {
	    if stopAgent {
		fmt.Printf("Stopping agent %s\n", configData.agentName)

		cmdStopAgnt := &exec.Cmd {
	          Path: cmdStopAgntPath,
	          Args: [] string {cmdStopAgntPath,"-p", configData.coordinationQMgr, configData.agentName, "-i"},
	          Stdout: os.Stdout,
	          Stderr: os.Stdout,
	        }

		err := cmdStopAgnt.Run()
                if err != nil {
	          fmt.Println("Error: ", err)
                }
	        return
            }

            var outb, errb bytes.Buffer
	    cmdListAgents := &exec.Cmd {
              Path: listAgentPath,
              Args: [] string {listAgentPath, "-p", configData.coordinationQMgr, configData.agentName},
            }
            cmdListAgents.Stdout = &outb
            cmdListAgents.Stderr = &errb

            err := cmdListAgents.Run()
            if err != nil {
              fmt.Println("Error: ", err)
              return
            }

            var agentStatus string
            agentStatus = outb.String()  
           if strings.Contains(agentStatus,"UNKNOWN") {
             fmt.Println("Agent status unknown. Ping the agent")
	         return
           } else {
	         // Sleep for 5 seconds
	         time.Sleep(5100 * time.Millisecond)
           }
          }
      } else {
         fmt.Println("Agent not started. Quitting")
         return
      }
}



