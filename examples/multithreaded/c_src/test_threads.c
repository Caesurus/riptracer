#include <stdio.h>
#include <string.h>
#include <pthread.h>
#include <unistd.h>
#include <stdlib.h>

// Global variable:
int i = 2;

int *foo(int input)
{
  // Print value received as argument:
  printf("Value received as argument in starting routine: %d\n", input);
  for (int i = 0; i < 1; i++)
  {
    printf(".");
    fflush(stdout);
    sleep(1);
  }
  i = input;
  // Return reference to global variable:
  pthread_exit(&i);
}

int foo2(int input)
{
  return input;
}

int main(void)
{
  // Declare variable for thread's ID:
  pthread_t id;
  pthread_t id2;
  pthread_t id3;

  pthread_create(&id, NULL, foo, 1);
  pthread_create(&id2, NULL, foo, 2);
  pthread_create(&id3, NULL, foo, 3);

  system("echo hello from a system call");

  int *ptr;

  // Wait for foo() and retrieve value in ptr;
  pthread_join(id3, (void **)&ptr);
  printf("Value recevied by parent from child 3: %i\n", *ptr);
  pthread_join(id2, (void **)&ptr);
  printf("Value recevied by parent from child 2: %i\n", *ptr);
  pthread_join(id, (void **)&ptr);
  printf("Value recevied by parent from child 1: %i\n", *ptr);

  foo2(11);
  foo2(11);
}
