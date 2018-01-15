CC=gcc -g -DDEBUG
CIFLAGS= -I$(UTILS) -I../ZMQ/include  
CFLAGS= -lm -L../ZMQ/zmqlibs -lzmq  -L../ZMQ/czmqlibs/lib -lczmq -Wimplicit-function-declaration -fPIC -Wall

all:simple_client.c simple_server.c
	$(CC) simple_client.c -o simple_client $(CIFLAGS) $(CFLAGS)
	$(CC) simple_server.c -o simple_server $(CIFLAGS) $(CFLAGS)

simple_client:simple_client.c
	$(CC) simple_client.c -o $@ $(CIFLAGS) $(CFLAGS)

simple_server:simple_server.c
	$(CC) simple_server.c -o $@ $(CIFLAGS) $(CFLAGS)

.PHONY:clean
clean:
	rm simple_client simple_server

