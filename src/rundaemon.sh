#!/bin/bash

export LD_LIBRARY_PATH=/usr/local/lib:/home/docker/COLAS/src/abd:/home/docker/COLAS/src/soda:/home/docker/COLAS/src/codes:/home/docker/COLAS/src/sodaw

make rebuild

args=("$@")
ptype=${args[0]}
filesize=${args[1]}

if [ -z "$filesize" ]; then
	./process --process-type ${ptype}
else
	/process --process-type ${ptype} --init-file-size ${filesize}
fi
