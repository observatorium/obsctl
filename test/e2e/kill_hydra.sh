#!/bin/bash

PID=$(pgrep hydra)
echo $PID
kill -9 $PID

exit 0