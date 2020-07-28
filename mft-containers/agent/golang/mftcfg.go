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
)

func main () {
 // Set BFG_DATA environment variable so that we can run MFT commands.
 os.Setenv("BFG_DATA",os.Args[6])

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

 // Setup coordination configuration
 fmt.Printf("Setting up coordination configuration %s for agent %s\n", os.Args[1], os.Args[5])
 cmdSetupCoord := &exec.Cmd {
	Path: cmdCoordPath,
	Args: [] string {cmdCoordPath, "-coordinationQMgr", os.Args[1], "-coordinationQMgrHost", os.Args[2], "-coordinationQMgrPort", os.Args[3], "-coordinationQMgrChannel", os.Args[4], "-f"},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }
 
 // Execute the fteSetupCoordination command. Log an error an exit in case of any error.
 if err := cmdSetupCoord.Run(); err != nil {
	fmt.Println("fteSetupCoordination command failed. The error is: ", err);
	return
 }

 // Setup commands configuration
 fmt.Printf("Setting up commands configuration %s for agent %s\n", os.Args[1], os.Args[5])
 cmdCmds := &exec.Cmd {
	Path: cmdCmdsPath,
	Args: [] string {cmdCmdsPath, "-connectionQMgr", os.Args[1], "-connectionQMgrHost", os.Args[2], "-connectionQMgrPort", os.Args[3], "-connectionQMgrChannel", os.Args[4],"-f"},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }

 // Execute the fteSetupCommands command. Log an error an exit in case of any error.
 if err := cmdCmds.Run(); err != nil {
	fmt.Println("fteSetupCommands command failed. The errror is: ", err);
	return
 }

 // Create agent.
 fmt.Printf("Creating agent %s\n", os.Args[5])
 cmdCrtAgnt := &exec.Cmd {
	Path: cmdCrtAgntPath,
	Args: [] string {cmdCrtAgntPath,"-agentName", os.Args[5], "-agentQMgr", os.Args[1], "-agentQMgrHost", os.Args[2], "-agentQMgrPort", os.Args[3], "-agentQMgrChannel", os.Args[4],"-credentialsFile","/usr/local/bin/MQMFTCredentials.xml", "-f"},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }

 // Execute the fteCreateAgent command. Log an error an exit in case of any error.
 if err := cmdCrtAgnt.Run(); err != nil {
	fmt.Println("fteCreateAgent command failed. The error is: ", err);
	return
 }

 fmt.Printf("Starting agent %s\n", os.Args[5])
 cmdStrAgnt := &exec.Cmd {
	Path: cmdStrAgntPath,
	Args: [] string {cmdStrAgntPath,"-p", os.Args[1], os.Args[5]},
	Stdout: os.Stdout,
	Stderr: os.Stdout,
 }
 
 // Run fteStartAgent command. Log an error and exit in case of any error.
 if err := cmdStrAgnt.Run(); err != nil {
	fmt.Println("Error:", err);
	return
 }

 fmt.Printf("Verifying status of agent %s\n", os.Args[5])
 listAgentPath, lookErr := exec.LookPath("fteListAgents")
 if lookErr != nil {
    panic(lookErr)
    return
 }

 var outb, errb bytes.Buffer
 cmdListAgents := &exec.Cmd {
	Path: listAgentPath,
	Args: [] string {listAgentPath, "-p", os.Args[1], os.Args[5]},
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
		fmt.Printf("Stopping agent %s\n", os.Args[5])

		cmdStopAgnt := &exec.Cmd {
	          Path: cmdStopAgntPath,
	          Args: [] string {cmdStopAgntPath,"-p", os.Args[1], os.Args[5], "-i"},
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
              Args: [] string {listAgentPath, "-p", os.Args[1], os.Args[5]},
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



