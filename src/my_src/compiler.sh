gcc -g -DDEBUG -fPIC -o ${1} ${1}.c -DASLIBRARY  -lm -L../ZMQ/zmqlibs -lzmq  -L../ZMQ/czmqlibs -lczmq -Wimplicit-function-declaration -Wall
