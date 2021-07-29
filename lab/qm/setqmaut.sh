# Authorize agents to connect with identity context. 
setmqaut -m MQMFT -t qmgr -p app +connect +inq +setid

# Authorize agents to publish on SYSTEM.FTE queue and topic.
setmqaut -m MQMFT -n SYSTEM.FTE -t q -p app +all
setmqaut -m MQMFT -n SYSTEM.FTE -t topic -p app +all
setmqaut -m MQMFT -n SYSTEM.DEFAULT.MODEL.QUEUE -t q -p app +put +get +dsp

# Set authorities on SRCAGENT agent queues.
setmqaut -m MQMFT -n SYSTEM.FTE.COMMAND.SRCAGENT -t q -p app +get +setid +browse +put
setmqaut -m MQMFT -n SYSTEM.FTE.DATA.SRCAGENT -t q -p app +put +get
setmqaut -m MQMFT -n SYSTEM.FTE.EVENT.SRCAGENT -t q -p app +put +get
setmqaut -m MQMFT -n SYSTEM.FTE.REPLY.SRCAGENT -t q -p app +put +get
setmqaut -m MQMFT -n SYSTEM.FTE.STATE.SRCAGENT -t q -p app +put +get +inq +browse
setmqaut -m MQMFT -n SYSTEM.FTE.HA.SRCAGENT -t q -p app +put +get +inq

# Set authorities on DESTAGENT agent queues.
setmqaut -m MQMFT -n SYSTEM.FTE.COMMAND.DESTAGENT -t q -p app +get +setid +browse +put
setmqaut -m MQMFT -n SYSTEM.FTE.DATA.DESTAGENT -t q -p app +put +get
setmqaut -m MQMFT -n SYSTEM.FTE.EVENT.DESTAGENT -t q -p app +put +get
setmqaut -m MQMFT -n SYSTEM.FTE.REPLY.DESTAGENT -t q -p app +put +get
setmqaut -m MQMFT -n SYSTEM.FTE.STATE.DESTAGENT -t q -p app +put +get +inq +browse
setmqaut -m MQMFT -n SYSTEM.FTE.HA.DESTAGENT -t q -p app +put +get
