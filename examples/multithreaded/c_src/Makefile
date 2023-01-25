CC=gcc
CFLAGS=-I. 

test_threads: test_threads.o 
	$(CC) -ggdb3 -std=c11 -Wall -Wextra -pedantic -pthread -o test_threads test_threads.o

.PHONY: clean

clean:
	rm -f *.o *~ core test_threads