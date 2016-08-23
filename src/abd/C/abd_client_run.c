//  Asynchronous client-to-server (DEALER to ROUTER)
//
//  While this example runs in a single process, that is to make
//  it easier to start and stop the example. Each task has its own
//  context and conceptually acts as a separate process.

#include "czmq.h"
#include <stdio.h>
#include <stdlib.h>
#include <math.h>

#include <signal.h>
#include <unistd.h>
#include <fcntl.h>
#include "algo_utils.h"


#define WRITE_VALUE "WRITE_VALUE"
#define GET_TAG "GET_TAG"

#define S_NOTIFY_MSG " "
#define S_ERROR_MSG "Error while writing to self-pipe.\n"

static int s_interrupted=0;
static void s_signal_handler(int signal_value)
{
    s_interrupted=1;
}

static void s_catch_signals ()
{
    struct sigaction action;
    action.sa_handler = s_signal_handler;
    //  Doesn't matter if SA_RESTART set because self-pipe will wake up zmq_poll
    //  But setting to 0 will allow zmq_read to be interrupted.
    action.sa_flags = 0;
    sigemptyset (&action.sa_mask);
    sigaction (SIGINT, &action, NULL);
    sigaction (SIGTERM, &action, NULL);
}



//  This is our client task
//  It connects to the server, and then sends a request once per second
//  It collects responses as they arrive, and it prints them out. We will
//  run several client tasks in parallel, each with a different random ID.

static TAG get_max_tag_phase(char *obj_name, unsigned int op_num, 
                      zsock_t *sock_to_servers,  char **servers, 
                          unsigned int num_servers, char *port)
{

    // send out the messages to all servers

    char buf[400];
    char algorithm[64];
    char phase[64];
    char tag_str[64];
    char buf1[400];
    int round;

    zmq_pollitem_t items [] = { { sock_to_servers, 0, ZMQ_POLLIN, 0 } };

    int i; 
    zframe_t *obj_name_frame = zframe_new(obj_name, strlen(obj_name));
    zframe_t *algo = zframe_new("ABD", 3);
    zframe_t *phase_frame = zframe_new(GET_TAG, 7);
    zframe_t *op_num_frame = zframe_new((const void *)&op_num, sizeof(int));

    for(i=0; i < num_servers; i++) {
       zframe_send(&obj_name_frame, sock_to_servers, ZFRAME_REUSE + ZFRAME_MORE);
       zframe_send(&algo, sock_to_servers, ZFRAME_REUSE + ZFRAME_MORE);
       zframe_send(&phase_frame, sock_to_servers, ZFRAME_REUSE + ZFRAME_MORE);
       zframe_send(&op_num_frame, sock_to_servers, ZFRAME_REUSE);
       printf("     \tsending to server %d\n",i);
    }

    zframe_destroy(&obj_name_frame);
    zframe_destroy(&algo);
    zframe_destroy(&phase_frame);
    zframe_destroy(&op_num_frame);

//    zframe_destroy (&payloadf);
    unsigned int majority =  ceil((num_servers+1)/2);
//     zmq_pollitem_t items [] = { { sock_to_servers, 0, ZMQ_POLLIN, 0 } };
     unsigned int responses =0;
     zlist_t *tag_list = zlist_new();
     
     TAG *tag;
     while (true) {
        //  Tick once per second, pulling in arriving messages
            
           // zmq_pollitem_t items [] = { { sock_to_servers, 0, ZMQ_POLLIN, 0 } };
            printf("      \treceiving data\n");
            int rc = zmq_poll(items, 1, -1);
            if(rc < 0 ||  s_interrupted ) {
                printf("Interrupted!\n");
                exit(0);
            }
           // zclock_sleep(300); 
            if (items [0].revents & ZMQ_POLLIN) {
                zmsg_t *msg = zmsg_recv (sock_to_servers);

                //identity
/*
                zframe_t *identity = zmsg_pop(msg);
                _zframe_str(identity,buf);
                printf("identity  : %s\n",buf);
*/

                //object
                zframe_t *identity = zmsg_pop(msg);
                _zframe_str(identity,buf);
                printf("\t\tobject    : %s\n",buf);

                // algorithm
                identity = zmsg_pop(msg);
                _zframe_str(identity,buf);
                printf("\t\talgorithm : %s\n",buf);

                // phase
                identity = zmsg_pop(msg);
                _zframe_str(identity,phase);
                printf("\t\tphase     : %s\n",phase);

                // operation number
                identity = zmsg_pop(msg);
                _zframe_int(identity, &round);
                printf("\t\tOP_NUM    : %d\n", round);

                identity = zmsg_pop(msg);
                _zframe_str(identity, tag_str);
                printf("\t\tTAG    : %s\n", tag_str);
                
                zmsg_destroy (&msg);
                if(round==op_num && strcmp(phase, GET_TAG)==0) {
                   responses++;

                   // add tag to list                
                   tag = (TAG *)malloc(sizeof(TAG));
                   string_to_tag(tag_str, tag);
                   zlist_append(tag_list, (void *)tag);

                   if(responses >= majority) break;
                   //if(responses >= num_servers) break;
                }
                else{
                     printf("   OLD MESSAGES : %s  %d\n", phase, op_num);

                }
            }
     }
   //comute the max tag now and return 
     TAG max_tag = get_max_tag(tag_list);

     free_items_in_list(tag_list);
     zlist_destroy(&tag_list);
     return  max_tag;
}

//  This is our client task
//  It connects to the server, and then sends a request once per second
//  It collects responses as they arrive, and it prints them out. We will
//  run several client tasks in parallel, each with a different random ID.

//   write_value_phase(obj_name, writer_id,  op_num, sock_to_servers, servers, num_servers, port, payload, size, max_tag);

static TAG write_value_phase(  
                      char *obj_name,
                      char *writer_id, 
                      unsigned int op_num, 
                      zsock_t *sock_to_servers,  
                      char **servers, 
                      unsigned int num_servers, 
                      char *port, byte *payload, 
                      int size, 
                      TAG max_tag
                   )
{

    // send out the messages to all servers

    char buf[400];
    char algorithm[64];
    char phase[64];
    char tag_str[64];
    char buf1[400];
    int round;

    zmq_pollitem_t items [] = { { sock_to_servers, 0, ZMQ_POLLIN, 0 } };

    int i; 
    zframe_t *obj_name_frame = zframe_new(obj_name, strlen(obj_name));
    zframe_t *algo = zframe_new("ABD", 3);
    zframe_t *phase_frame = zframe_new(WRITE_VALUE, 11);
    zframe_t *op_num_frame = zframe_new((const void *)&op_num, sizeof(int));

    TAG new_tag;
    new_tag.z = max_tag.z + 1;
    strcpy(new_tag.id, writer_id);
    tag_to_string(new_tag, tag_str); 
    zframe_t *tag_frame = zframe_new(tag_str, strlen(tag_str));

    zframe_t *payload_frame = zframe_new(payload, size);

    for(i=0; i < num_servers; i++) {
       zframe_send(&obj_name_frame, sock_to_servers, ZFRAME_REUSE + ZFRAME_MORE);
       zframe_send(&algo, sock_to_servers, ZFRAME_REUSE + ZFRAME_MORE);
       zframe_send(&phase_frame, sock_to_servers, ZFRAME_REUSE + ZFRAME_MORE);
       zframe_send(&op_num_frame, sock_to_servers, ZFRAME_REUSE+ ZFRAME_MORE);
       zframe_send(&tag_frame, sock_to_servers, ZFRAME_REUSE+ ZFRAME_MORE);
       zframe_send(&payload_frame, sock_to_servers, ZFRAME_REUSE);
       printf("     \tsending to server %d\n",i);
    }

    zframe_destroy(&obj_name_frame);
    zframe_destroy(&algo);
    zframe_destroy(&phase_frame);
    zframe_destroy(&op_num_frame);
    zframe_destroy(&tag_frame);
    zframe_destroy (&payload_frame);

    unsigned int majority =  ceil((num_servers+1)/2);
     unsigned int responses =0;
     int j =0;
     zlist_t *tag_list = zlist_new();
     
     TAG *tag;
     while (true) {
        //  Tick once per second, pulling in arriving messages
            
           // zmq_pollitem_t items [] = { { sock_to_servers, 0, ZMQ_POLLIN, 0 } };
            printf("      \twaiting ack\n");
            int rc = zmq_poll(items, 1, -1);
            if(rc < 0 ||  s_interrupted ) {
                printf("Interrupted!\n");
                exit(0);
            }
           // zclock_sleep(300); 
           // zclock_sleep(300); 
            if (items [0].revents & ZMQ_POLLIN) {
                zmsg_t *msg = zmsg_recv (sock_to_servers);

                 //identity variable is abused
                //object
                zframe_t *identity = zmsg_pop(msg);
                _zframe_str(identity,buf);
                printf("\t\tobject    : %s\n",buf);

                // algorithm
                identity = zmsg_pop(msg);
                _zframe_str(identity,buf);
                printf("\t\talgorithm : %s\n",buf);

                // phase
                identity = zmsg_pop(msg);
                _zframe_str(identity,phase);
                printf("\t\tphase     : %s\n",phase);

                // operation number
                identity = zmsg_pop(msg);
                _zframe_int(identity, &round);
                printf("\t\tOP_NUM    : %d\n", round);

                // tag string
                identity = zmsg_pop(msg);
                _zframe_str(identity, tag_str);
                printf("\t\tTAG STRING    : %s\n", tag_str);

/*
                // status
                identity = zmsg_pop(msg);
                _zframe_str(identity, tag_str);
                printf("\t\tSTATUS    : %s\n", tag_str);
*/


                zmsg_destroy (&msg);
                if(round==op_num && strcmp(phase, WRITE_VALUE)==0) {
                   responses++;

                   // add tag to list                
                   tag = (TAG *)malloc(sizeof(TAG));
                   string_to_tag(tag_str, tag);
                   zlist_append(tag_list, (void *)tag);

                   if(responses >= majority) break;
                   //if(responses >= num_servers) break;
                }
                else{
                     printf("   OLD MESSAGES : %s  %d\n", phase, op_num);

                }
            }
     }
     return ;
}


static bool abd_write(char *obj_name, char *writer_id, unsigned int op_num,  byte *payload, unsigned int size, char **servers, 
                          unsigned int num_servers, char *port)
{

    int j;
    zctx_t *ctx  = zctx_new();
    void *sock_to_servers = zsocket_new(ctx, ZMQ_DEALER);
    zctx_set_linger(ctx, 0);
    assert (sock_to_servers);

    zsocket_set_identity(sock_to_servers,  writer_id);
    for(j=0; j < num_servers; j++) {    
       char *destination = create_destination(servers[j], port);
       int rc = zsocket_connect(sock_to_servers, destination);
       assert(rc==0);
       free(destination);
    }

  //for( i=0; i < 50; i++) {
   // printf("WRITE %d\n", i);
   printf("     MAX_TAG\n");
   TAG max_tag=  get_max_tag_phase(obj_name,  op_num, sock_to_servers, servers, num_servers, port);

   printf("\tmax tag (%d,%s)\n\n", max_tag.z, max_tag.id);

   printf("     WRITE_VALUE\n");
   write_value_phase(obj_name, writer_id,  op_num, sock_to_servers, servers, num_servers, port, payload, size, max_tag);
  //}

    zsocket_destroy(ctx, sock_to_servers);
    zctx_destroy(&ctx);
  //}


    return true;
}

static bool abd_read(byte *payload, unsigned int size, char **servers, 
                          unsigned int num_servers, char *port)
{

    //client_task(payload, size, servers, num_servers, port);

    return true;
}
//  The main thread simply starts several clients and a server, and then
//  waits for the server to finish.
#undef MAIN
#ifdef MAIN




int main (void)
{
   int i ; 
   
   byte *payload = (byte *)malloc(100000000*sizeof(byte));
   unsigned int size = 100000000*sizeof(byte);

/*
   char *servers[]= {
                     "tcp://172.17.0.7", "tcp://172.17.0.5", 
                     "tcp://172.17.0.4", "tcp://172.17.0.6",
                     "tcp://172.17.0.3"
                   };

*/


   char *servers[] = {
"tcp://172.17.0.22", "tcp://172.17.0.21", "tcp://172.17.0.18", "tcp://172.17.0.17", "tcp://172.17.0.20", "tcp://172.17.0.16", "tcp://172.17.0.19", "tcp://172.17.0.15", "tcp://172.17.0.14", "tcp://172.17.0.13", "tcp://172.17.0.12", "tcp://172.17.0.11", "tcp://172.17.0.10", "tcp://172.17.0.9", "tcp://172.17.0.7", "tcp://172.17.0.8", "tcp://172.17.0.6", "tcp://172.17.0.5", "tcp://172.17.0.4", "tcp://172.17.0.3"
                     };

/*
   char *servers[]= {
                     "tcp://172.17.0.2"
                     };
*/


   unsigned int num_servers = 20;
   char port[]= {"5570"};

   char writer_id[] = { "writer_1"};
   char obj_name[] = {"object"};

   unsigned int op_num;
   s_catch_signals();

  for( i=0; i < 5; i++) {
    printf("\nWRITE %d\n", i);
    abd_write(obj_name, writer_id, i,  payload, size, servers, num_servers, port);
  }

   return 0;
}

#endif
