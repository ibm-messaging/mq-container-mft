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

 // Get the path of MFT ftePingAgent command. 
 cmdPingAgntPath, lookErr:= exec.LookPath("ftePingAgent")
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
 cmdListAgentPath, lookErr := exec.LookPath("fteListAgents")
 if lookErr != nil {
    panic(lookErr)
    return
 }
 
 // Prepare fteListAgents command for execution
 var outb, errb bytes.Buffer
 cmdListAgents := &exec.Cmd {
	Path: cmdListAgentPath,
	Args: [] string {cmdListAgentPath, "-p", os.Args[1], os.Args[5]},
 }
 cmdListAgents.Stdout = &outb
 cmdListAgents.Stderr = &errb

 // Execute and get the output of the command into a byte buffer.
 err := cmdListAgents.Run()
 if err != nil {
	fmt.Println("Error: ", err)
	return
 }

 // Now parse the output of fteListAgents command and take
 // appropriate actions.
 var agentStatus string 
 agentStatus = outb.String()
 
  // Create a go routine to read the agent output0.log file
  done := make(chan bool, 1)
  go func() {
   var agentLogPath string
   agentLogPath = os.Args[6] + "/mqft/logs/" + os.Args[1] + "/agents/" + os.Args[5] + "/logs/output0.log"
   f, err := os.Open(agentLogPath)
   if err != nil {
     fmt.Printf("error opening file: %v\n",err)
     return
    }
    r := bufio.NewReader(f)
    for {
       s, e := Readln(r)
       for e == nil {
         fmt.Println(s)
         s,e = Readln(r)
       }
    }
   }()


 // If agent status is READY, then we are good. 
 if strings.Contains(agentStatus,"READY") == true  {
	// Agent is READY, so start monitoring the status.
	// If the status becomes unknown, this monitoring program terminates
	// thus container also ends.
	fmt.Println("Agent has started. Starting to monitor status")
    // Setup channel to handle signals to stop agent
	sigs := make(chan os.Signal, 1)
      //done := make(chan bool, 1)
      // Notify monitor program when SIGINT or SIGTERM is 
	  // issued to container.
      signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
      var stopAgent bool
      // Handler for 
      go func() {
       sig := <-sigs
	   fmt.Printf("Received singal %s\n:", sig)
	   stopAgent = true
           done <- true
     }()

    // Determine the sleep time.
	sleepTime, err := strconv.Atoi(os.Args[7])
	if err != nil {
	  // There was some error when getting the sleep time. Assume 300 seconds
	  // as the default sleep time
	  sleepTime = 300
	}

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
	     fmt.Println("An error occured when running fteStopAgent command. The error is: ", err)
       }
	   return
     } // End of stopAgent processing

     // Keep running fteListAgents at specified interval.
     var outb, errb bytes.Buffer
	 cmdListAgents := &exec.Cmd {
       Path: cmdListAgentPath,
       Args: [] string {cmdListAgentPath, "-p", os.Args[1], os.Args[5]},
     }
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
	    Args: [] string {cmdPingAgntPath, "-p", os.Args[1], os.Args[5]},
       }
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
	     fmt.Printf("Agent %s did not respond to ping. Monitor exiting\n", os.Args[5])
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



