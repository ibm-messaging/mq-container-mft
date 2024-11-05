#!/bin/bash

# © Copyright IBM Corporation 2024
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set +ex

setmqaut -m MFTQM -n SYSTEM.FTE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE -t topic -p app +all
setmqaut -m MFTQM -n SYSTEM.DEFAULT.MODEL.QUEUE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHADM1.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHAGT1.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHMON1.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHOPS1.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHSCH1.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHTRN1.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.COMMAND.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.DATA.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.EVENT.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.REPLY.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.STATE.SRC -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHADM1.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHAGT1.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHMON1.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHOPS1.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHSCH1.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.AUTHTRN1.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.COMMAND.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.DATA.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.EVENT.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.REPLY.BRIDGE -t q -p app +all
setmqaut -m MFTQM -n SYSTEM.FTE.STATE.BRIDGE -t q -p app +all
