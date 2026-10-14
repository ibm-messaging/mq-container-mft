/*
© Copyright IBM Corporation 2020, 2026

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
 * Readiness probe for IBM MQ Managed File Transfer agents running in containers.
 *
 * Exit codes:
 *   0 — Agent is ready          (AgentStatus=STARTED, queueManagerStatus=AVAILABLE)
 *   1 — Agent is not ready      (AgentStatus is not STARTED)
 *   2 — IPC error               (MFT_AGENT_CONFIG_FILE not set, or socket communication failed)
 *   3 — Config file error       (file unreadable or parse error)
 *   4 — Agent not found         (agent name not present in the config file)
 *   5 — QM not connected        (Agent is alive/STARTED but queueManagerStatus is not AVAILABLE)
 */
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/user"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ibm-messaging/mq-container-mft/pkg/utils"
	"github.com/tidwall/gjson"
)

// EYE_CATCHER is the 4-byte magic number that starts every FTEI frame.
const EYE_CATCHER = "FTEI"

// GET_AGENT_STATUS is the 4-byte command word for a "get agent status" request.
// The agent replies with a comma-separated key=value payload.
const GET_AGENT_STATUS = "GETA"

const AGENT_REDY_EXIT_CODE_0 = 0 // ready
const AGENT_REDY_EXIT_CODE_1 = 1 // not ready
const AGENT_REDY_EXIT_CODE_2 = 2 // IPC / env error
const AGENT_REDY_EXIT_CODE_3 = 3 // config file error
const AGENT_REDY_EXIT_CODE_4 = 4 // agent not found in config
const AGENT_REDY_EXIT_CODE_5 = 5 // agent alive but queue manager not connected

func readConfigFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func main() {
	agentNameEnv, agentNameEnvSet := os.LookupEnv("MFT_AGENT_NAME")
	if !agentNameEnvSet {
		msg := utils.AGENT_REDY_ENV_AGENT_NAME_NOT_SET_3001
		fmt.Fprintln(os.Stderr, msg)
		utils.WriteProbeLog("READINESS", "FAILED", msg)
		os.Exit(AGENT_REDY_EXIT_CODE_1)
	}

	bfgConfigFilePath, bfgConfigFilePathSet := os.LookupEnv("MFT_AGENT_CONFIG_FILE")
	if !bfgConfigFilePathSet {
		msg := fmt.Sprintf(utils.MFT_ENV_AGNT_CFG_FILE_NOT_SPECIFIED, utils.MFT_DEFAULT_CONFIG_JSON)
		fmt.Fprintln(os.Stderr, msg)
		utils.WriteProbeLog("READINESS", "INFO", msg)
		bfgConfigFilePath = utils.MFT_DEFAULT_CONFIG_JSON
	} else {
		bfgConfigFilePath = strings.TrimSpace(bfgConfigFilePath)
	}

	agentConfig, e := readConfigFile(bfgConfigFilePath)
	if e != nil {
		msg := fmt.Sprintf(utils.AGENT_REDY_ENV_CFG_FILE_READ_3003, bfgConfigFilePath, e)
		fmt.Fprintln(os.Stderr, msg)
		utils.WriteProbeLog("READINESS", "FAILED", msg)
		os.Exit(AGENT_REDY_EXIT_CODE_3)
	}

	// Find the queue manager name for this agent in the config.
	// Config schema: {"agents":[{"name":"AGENT1","qmgrName":"qm1",...},...]}
	var agentQueueManager string
	configurationFound := false
	agentsJson := gjson.Get(agentConfig, "agents").Array()
	for _, agent := range agentsJson {
		if strings.Contains(gjson.Get(agent.String(), "name").String(), agentNameEnv) {
			agentQueueManager = gjson.Get(agent.String(), "qmgrName").String()
			configurationFound = true
			break
		}
	}

	if !configurationFound {
		msg := fmt.Sprintf(utils.AGENT_REDY_NOT_IN_CFG_3006, agentNameEnv, bfgConfigFilePath)
		fmt.Fprintln(os.Stderr, msg)
		utils.WriteProbeLog("READINESS", "FAILED", msg)
		os.Exit(AGENT_REDY_EXIT_CODE_4)
	}

	agentStatus, qmStatus, err := getAgentStatus(agentNameEnv, agentQueueManager)
	if err != nil {
		msg := fmt.Sprintf(utils.AGENT_REDY_IPC_ERROR_3007, agentNameEnv, err)
		fmt.Fprintln(os.Stderr, msg)
		utils.WriteProbeLog("READINESS", "FAILED", msg)
		os.Exit(AGENT_REDY_EXIT_CODE_2)
	}

	detail := fmt.Sprintf("AgentStatus=%s, queueManagerStatus=%s", agentStatus, qmStatus)

	if agentStatus == "STARTED" && qmStatus == "AVAILABLE" {
		utils.WriteProbeLog("READINESS", "PASSED", fmt.Sprintf("agent %s is ready: %s", agentNameEnv, detail))
		os.Exit(AGENT_REDY_EXIT_CODE_0)
	}

	if agentStatus == "STARTED" && qmStatus != "AVAILABLE" {
		msg := fmt.Sprintf(utils.AGENT_REDY_QM_NOT_CONNECTED_3004, agentNameEnv, detail)
		fmt.Fprintln(os.Stderr, msg)
		utils.WriteProbeLog("READINESS", "FAILED", msg)
		os.Exit(AGENT_REDY_EXIT_CODE_5)
	}

	msg := fmt.Sprintf(utils.AGENT_REDY_NOT_READY_3005, agentNameEnv, detail)
	fmt.Fprintln(os.Stderr, msg)
	utils.WriteProbeLog("READINESS", "FAILED", msg)
	os.Exit(AGENT_REDY_EXIT_CODE_1)
}

// IPCMessage represents one FTEI framed message on the wire.
// Layout (all integers big-endian, 4 bytes each):
//
//	bytes  0– 3  eyeCatcher          "FTEI"
//	bytes  4– 7  messageHeaderLength  8
//	bytes  8–11  messageId            random uint32, echoed back in the reply
//	bytes 12–15  status               0 = request / success
//	bytes 16–19  messageTextLength    byte count of the text that follows
//	bytes 20+    messageText          command word or reply payload
type IPCMessage struct {
	eyeCatcher          []byte
	messageHeaderLength uint32
	messageId           uint32
	status              uint32
	messageTextLength   uint32
	messageText         []byte
}

// readReply runs in a goroutine and performs a single blocking read from r.
// The goroutine pattern lets the caller write the request and wait for the
// reply concurrently without blocking the main goroutine.
func readReply(r io.Reader, done chan bool, readBuf chan []byte) {
	buf := make([]byte, 2048)
	dataSize, err := r.Read(buf)
	if err != nil {
		// EOF with data means the agent closed the connection after writing —
		// that is still a valid reply.
		if err == io.EOF && dataSize > 0 {
			readBuf <- buf[:dataSize]
			done <- true
		} else {
			done <- false
		}
		return
	}
	if dataSize > 0 {
		readBuf <- buf[:dataSize]
		done <- true
	} else {
		done <- false
	}
}

// getAgentStatus connects to the agent's Unix domain socket, sends a GETA
// request, and returns (agentStatus, queueManagerStatus, err).
func getAgentStatus(agentName string, agentQMName string) (string, string, error) {
	// Socket path: $BFG_DATA/mqft/ipc/<username>/<AGENTNAME>@<QMGRNAME>
	// In containers running as a random UID with no /etc/passwd entry,
	// user.Current() fails. Fall back to the numeric UID string, which is
	// what the agent JVM uses as the directory name in that case.
	userName := strconv.Itoa(os.Getuid())
	if u, err := user.Current(); err == nil {
		userName = u.Username
	}
	pipePath := os.Getenv("BFG_DATA") + "/mqft/ipc/" + userName + "/" + agentName + "@" + agentQMName

	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	messageId := uint32(r1.Intn(10000))

	var req IPCMessage
	req.eyeCatcher = []byte(EYE_CATCHER)
	req.messageHeaderLength = 8
	req.messageId = messageId
	req.status = 0
	req.messageTextLength = 4
	req.messageText = []byte(GET_AGENT_STATUS)

	// Serialise the frame into a flat byte slice (big-endian integers).
	var requestBuffer []byte
	buf := make([]byte, 4)
	requestBuffer = append(requestBuffer, req.eyeCatcher...)
	binary.BigEndian.PutUint32(buf, req.messageHeaderLength)
	requestBuffer = append(requestBuffer, buf...)
	binary.BigEndian.PutUint32(buf, req.messageId)
	requestBuffer = append(requestBuffer, buf...)
	binary.BigEndian.PutUint32(buf, req.status)
	requestBuffer = append(requestBuffer, buf...)
	binary.BigEndian.PutUint32(buf, req.messageTextLength)
	requestBuffer = append(requestBuffer, buf...)
	requestBuffer = append(requestBuffer, req.messageText...)

	conn, err := net.Dial("unix", pipePath)
	if err != nil {
		return "", "", fmt.Errorf("dial unix %s: %w", pipePath, err)
	}
	defer conn.Close()

	// Start the reader goroutine before writing so it is ready to receive
	// the reply the moment it arrives.
	done := make(chan bool, 1)
	readBuf := make(chan []byte, 1)
	go readReply(conn, done, readBuf)

	if _, err = conn.Write(requestBuffer); err != nil {
		return "", "", fmt.Errorf("write to unix %s (msgID %d): %w", pipePath, messageId, err)
	}

	if ok := <-done; !ok {
		return "", "", fmt.Errorf("no reply received from unix %s (msgID %d)", pipePath, messageId)
	}

	reply := <-readBuf

	// Validate the reply frame.
	// The agent controls messageHeaderLength in its reply, so all subsequent
	// offsets are derived from it rather than hard-coded — this avoids the
	// message-ID mismatch failures seen in production where the agent used a
	// different header length than the 8 we send in requests.
	if len(reply) < 8 {
		return "", "", fmt.Errorf("unix %s: reply too short (%d bytes)", pipePath, len(reply))
	}

	var rep IPCMessage
	rep.eyeCatcher = make([]byte, 4)
	copy(rep.eyeCatcher, reply[0:4])
	if !reflect.DeepEqual(rep.eyeCatcher, []byte(EYE_CATCHER)) {
		return "", "", fmt.Errorf("unix %s: FTEI eye catcher mismatch: got %q want %q",
			pipePath, string(rep.eyeCatcher), EYE_CATCHER)
	}

	tempBuf := make([]byte, 4)
	copy(tempBuf, reply[4:8])
	replyHdrLen := binary.BigEndian.Uint32(tempBuf)

	// Header fields occupy replyHdrLen bytes starting at byte 8;
	// messageTextLength sits immediately after at byte 8+replyHdrLen.
	textLenOffset := 8 + replyHdrLen
	if uint32(len(reply)) < textLenOffset+4 {
		return "", "", fmt.Errorf("unix %s: reply too short for header (len=%d hdrLen=%d)",
			pipePath, len(reply), replyHdrLen)
	}

	// status is the second uint32 in the header fields (absolute offset 12).
	if replyHdrLen >= 8 {
		copy(tempBuf, reply[12:16])
		replyStatus := binary.BigEndian.Uint32(tempBuf)
		if replyStatus != 0 {
			return "", "", fmt.Errorf("unix %s: reply status non-zero: %d", pipePath, replyStatus)
		}
	}

	copy(tempBuf, reply[textLenOffset:textLenOffset+4])
	textLen := binary.BigEndian.Uint32(tempBuf)
	if textLen == 0 {
		return "", "", fmt.Errorf("unix %s: reply body is empty", pipePath)
	}
	// Guard against a malformed length field corrupting the slice.
	const maxTextLen uint32 = 2028
	if textLen > maxTextLen {
		textLen = maxTextLen
	}

	textOffset := textLenOffset + 4

	// Parse the GETA reply payload: comma-separated key=value string, e.g.:
	//   "agentType=STANDARD,...,AgentStatus=STARTED,queueManagerStatus=AVAILABLE,..."
	rawPayload := string(reply[textOffset : textOffset+textLen])

	var agentStatus string
	var qmStatus string
	for _, token := range strings.Split(rawPayload, ",") {
		kv := strings.SplitN(token, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if key == "AgentStatus" {
			agentStatus = val
		} else if key == "queueManagerStatus" {
			qmStatus = val
		}
	}

	return agentStatus, qmStatus, nil
}
