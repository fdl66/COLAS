
#include <czmq.h>
#include <zmq.h>
#include <czmq_library.h>

//  .split server task
//  This is our server task.
//  It uses the multithreaded server model to deal requests out to a pool
//  of workers and route replies back to clients. One worker can handle
//  one request at a time but one client can talk to multiple workers at
//  once.
#define BUFSIZE 100

typedef struct _Server_Args {
    char *init_data;
    unsigned int init_data_size;

    //char *server_id;
    char server_id[BUFSIZE];
    char servers_str[BUFSIZE];
    char port[BUFSIZE];
    char port1[BUFSIZE];
    void *sock_to_servers;
    int num_servers;
    int symbol_size;
    unsigned int coding_algorithm; // 0 if full-vector and 1 is reed-solomon
    unsigned int K;
    unsigned int N;
} Server_Args;

Server_Args *server_args ;
int server_network_is_connected  = 0;

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





static void server_worker (void *args, zctx_t *ctx);

void *server_task (Server_Args *server_args) {
    //  Frontend socket talks to clients over TCP
    zctx_t *ctx = zctx_new ();
    void *frontend = zsocket_new(ctx, ZMQ_ROUTER);

    char str[20];
    strcpy(str, "tcp://*:");
    strcat(str, server_args->port);
    printf("%s\n", str);

    zsocket_bind(frontend, str);

    //  Backend socket talks to workers over inproc
    void *backend = zsocket_new (ctx, ZMQ_DEALER);
    zsocket_bind (backend, "inproc://backend");

    //  Launch pool of worker threads, precise number is not critical
    //   for (thread_nbr = 0; thread_nbr < 5; thread_nbr++)
    zthread_fork (ctx, server_worker, server_args);

    //  Connect backend to frontend via a proxy
    zmq_proxy (frontend, backend, NULL);

    printf("back\n");	
    zsocket_destroy(ctx, frontend);
    zsocket_destroy(ctx, backend);	
    zctx_destroy (&ctx);
    return NULL;
}


void create_metadata_sending_sockets() {
    int num_servers = count_num_servers(server_args->servers_str);
    char **servers = create_server_names(server_args->servers_str);

    zctx_t *ctx  = zctx_new();
    void *sock_to_servers = zsocket_new(ctx, ZMQ_DEALER);
    zctx_set_linger(ctx, 0);
    assert (sock_to_servers);

    zsocket_set_identity(sock_to_servers,  server_args->server_id);

    int j;
    for(j=0; j < num_servers; j++) {
        char *destination = create_destination(servers[j], "");
        int rc = zsocket_connect(sock_to_servers, destination);
        assert(rc==0);
        free(destination);
    }

    destroy_server_names(servers, num_servers);
    server_args->sock_to_servers = sock_to_servers;
}




static void
server_worker (void *_server_args, zctx_t *ctx) {
    void *worker = zsocket_new (ctx, ZMQ_DEALER);
    zsocket_connect(worker, "inproc://backend");

    int64_t affinity = 5000;
    int rc = zmq_setsockopt(socket, ZMQ_SNDBUF, &affinity, sizeof affinity);
    rc = zmq_setsockopt(socket, ZMQ_RCVBUF, &affinity, sizeof affinity);

    zmq_pollitem_t items[] = { { worker, 0, ZMQ_POLLIN, 0}};

    while (true) {
        printf("\twaiting for message\n");
        int rc = zmq_poll(items, 1, -1);
        if( rc < 0 ) {
            exit(EXIT_FAILURE);
        }

        if (items[0].revents & ZMQ_POLLIN) {
            printf("\treceived message\n");
            zmsg_t *msg = zmsg_recv(worker);

            printf("recv a message. \n");
            if(server_network_is_connected ==0){
                printf("Try to create server networks...\n");
                create_metadata_sending_sockets();
                server_network_is_connected = 1;
            }
		
            printf("Try to send to other servers...\n");
            int op_num = randof(100);
            char *request_types[] = {"OBJECT", "ALGORITHM", "PHASE", "OPNUM"};
            send_request_to_servers(server_args->sock_to_servers, 1, request_types,  4, "obj_name", "ABD", "GET_TAG", &op_num) ;

            zclock_sleep(60);
            zmsg_destroy(&msg);
        }
    }
    printf("done\n");
}



int server_process( char* self_id, char *self_port, char* other_servers ) {
   server_args = (Server_Args *) malloc(sizeof(Server_Args));	
   strcpy(server_args->server_id, self_id);
   strcpy(server_args->port, self_port);
   strcpy(server_args->servers_str, other_servers);
    zthread_new(server_task, (void *)server_args);
    printf("Starting the server [id: %s]\n", server_args->server_id);

    while(true) {
        zclock_sleep(60*5*1000);
    }
    free(server_args);
    return 0;
}



int main(int argc, char *argv[] ){
	printf("./simple_server self_id self_port other_servers");
	server_process(argv[1], argv[2], argv[3]);
}


