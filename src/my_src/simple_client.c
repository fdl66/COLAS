
#include <czmq.h>
#include <zmq.h>
#include <czmq_library.h>

typedef struct _client_Args {
    char client_id[100];
    char servers_str[100];
    char port[10];
    char port1[10];
} ClientArgs;

unsigned int count_num_servers(char *servers_str) {
    int count = 0;
    if(strlen(servers_str)==0) return 0;
    count++;
    char *p = servers_str;
    while(*p !='\0') {
        if( *p=='_') count++;
        p++;
    }
    return count;
}

char **create_server_names(char *servers_str) {
    unsigned int num_servers =  count_num_servers(servers_str);
    char **servers = (char **)malloc(num_servers*sizeof(char *));
    char *p, *q;
    p = servers_str;
    int i = 0;
    while( *p!='\0') {
        servers[i] = (char *)malloc(50*sizeof(char));
        q = servers[i];
        while(*p !='_' && *p !='\0') {
            *q++ = *p++;
        }
        *q = '\0';
        if( *p == '\0') break;
        p++;
        i++;
    }
    return servers;
}

//accepted Create a destructor for this memory
void  destroy_server_names(char **servers, int num_servers) {
    int i =0;
    for(i=0; i < num_servers; i++) {
        free(servers[i]);
    }
    free(servers);

}


char *create_destination(char *server, char *port) {
    int size = 0;
    size += strlen(server);
    size += strlen(port);

    char *dest = (char *)malloc( (size + 8)*sizeof(char));
    assert(dest!=0);
    sprintf(dest, "tcp://%s%s", server, port);
    printf("%s\n", dest);	
    return dest;
}



void send_request_to_servers(void *sock_to_servers, int num_servers, char *names[],  int n, ...) {
    int i =0, j;  
    // generate frames
    va_list valist;
    va_start(valist, n);
    void **values = (void **)malloc(n*sizeof(void *));
    zframe_t **frames = (zframe_t **)malloc(n*sizeof(zframe_t *));
    assert(values!=NULL);
    assert(frames!=NULL);
    for(i=0; i < n; i++ ) {
        if( strcmp(names[i], "OPNUM")==0)   {
            values[i] = (void *)va_arg(valist, unsigned  int *);
            frames[i]= zframe_new( (const void *)values[i], sizeof(unsigned int));
        }else {
            values[i] = va_arg(valist, char *);
            frames[i]= zframe_new(values[i], strlen((char *)values[i]));
        }
    }
    va_end(valist);

    //send frames to servers
    int rc;
    for(i=0; i < num_servers; i++) {
        printf("\t\t\tsend to server : %d\n", i);
        for(j=0; j < n-1; j++) {//n-1
                rc = zframe_send(&frames[j], sock_to_servers, ZFRAME_REUSE +  ZFRAME_MORE);
                if( rc < 0) {
                    printf("ERROR: %d\n", rc);
                    exit(EXIT_FAILURE);
                }
                assert(rc!=-1);
        }
	rc = zframe_send(&frames[j], sock_to_servers, ZFRAME_REUSE + ZFRAME_DONTWAIT);
        if( rc < 0) {
            printf("ERROR: %d\n", rc);
            exit(EXIT_FAILURE);
        }

    }
    printf("\n");
	
    if( values!=NULL) 		free(values);
    for(i=0; i < n; i++ )
		zframe_destroy(&frames[i]);
    if( frames!=NULL) 		free(frames);
}


void client_task( zsock_t *sock_to_servers, unsigned int num_servers ) {

   // zmq_pollitem_t items [] = { { sock_to_servers, 0, ZMQ_POLLIN, 0 } };
    //while (true) {
        printf("\t\tsend request....\n");
	 	
	int op_num = randof(100);
	char *request_types[] = {"OBJECT", "ALGORITHM", "PHASE", "OPNUM"};
	send_request_to_servers(sock_to_servers, num_servers, request_types,  4, "obj_name", "ABD", "GET_TAG", &op_num) ;
	zclock_sleep(op_num*10);
	/*	
        int rc = zmq_poll(items, 1, -1);
        if(rc < 0 ||  s_interrupted ) {
            printf("Interrupted!\n");
            exit(EXIT_FAILURE);
        }
        printf("\t\treceived data\n");

        if (items [0].revents & ZMQ_POLLIN) {
            zmsg_t *msg = zmsg_recv (sock_to_servers);
            zmsg_destroy (&msg);
        }
     */   
    //}
}



void *setup_comm_with_servers(ClientArgs *client_args) {
    int j;
    int num_servers = count_num_servers(client_args->servers_str);
    char **servers = create_server_names(client_args->servers_str);

    zctx_t *ctx  = zctx_new();
    void *sock_to_servers = zsocket_new(ctx, ZMQ_DEALER);
    assert (sock_to_servers);
    zsocket_set_identity(sock_to_servers,  client_args->client_id);
    //connect to all servers
    for(j=0; j < num_servers; j++) {
        char *destination = create_destination(servers[j], client_args->port);
        int rc = zsocket_connect(sock_to_servers, (const char *)destination);
        assert(rc==0);
        free(destination);
    }
   destroy_server_names(servers, num_servers);
    return sock_to_servers;
}

int  main(int argc, char **argv) {
    ClientArgs client ;
    printf("./simple_client client_id servers");
    strcpy(client.client_id, argv[1]);
    strcpy(client.servers_str, argv[2]);
    client.port[0] = '\0';
    client.port1 [0]= '\0';

    int num_servers = count_num_servers(client.servers_str);
    zsock_t *sock =  setup_comm_with_servers(&client);
    	
    int opnum;
    for( opnum=0; opnum< 1; opnum++) {
	    client_task(sock, num_servers);
    }	
   
}



