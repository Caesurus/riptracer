#include <stdio.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <stdlib.h>
#include <sys/wait.h>

void forkexample()
{
    int pid = fork();

    if (pid == 0){
        // child process because return value zero  
        printf("Hello from Child %d! \n",getpid());
        for(int i=0; i < 1; i++){ 
          printf(".");
          fflush(stdout);
          sleep(1);
        }
    }else{
        // parent process because return value non-zero.
        printf("Hello from Parent %d!, child: %d\n", getpid(), pid);
        waitpid(pid, NULL, 0);
    }
}

int main(void){
    // Declare variable for thread's ID:
    forkexample();
}

