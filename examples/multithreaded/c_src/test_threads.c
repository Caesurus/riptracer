#include <stdio.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <stdlib.h>

// Global variable:
int i = 2;

void* foo(){
  // Print value received as argument:
  printf("Value recevied as argument in starting routine: ");
  for(int i=0; i < 1; i++){ 
    printf(".");
    fflush(stdout);
    sleep(1);
  }
  // Return reference to global variable:
  pthread_exit(&i);
}


void forkexample()
{
    // child process because return value zero
    if (fork() == 0)
        printf("Hello from Child!\n");
  
    // parent process because return value non-zero.
    else
        printf("Hello from Parent!\n");
}

int main(void){
  // Declare variable for thread's ID:
  pthread_t id;
  pthread_t id2;
  pthread_t id3;

  int j = 1;
  pthread_create(&id, NULL, foo, &j);
  pthread_create(&id2, NULL, foo, &j);
  pthread_create(&id3, NULL, foo, &j);

  forkexample();

  system("echo hello from a system call");

  int* ptr;

  // Wait for foo() and retrieve value in ptr;
  pthread_join(id, (void**)&ptr);
  pthread_join(id2, (void**)&ptr);
  pthread_join(id3, (void**)&ptr);
  printf("Value recevied by parent from child: ");
  printf("%i\n", *ptr);
}

