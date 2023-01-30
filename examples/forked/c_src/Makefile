CC=gcc
CFLAGS=-I. 

test_forks: test_forks.o 
	$(CC) -ggdb3 -std=c11 -Wall -Wextra -pedantic -pthread -o test_forks test_forks.o

.PHONY: clean

clean:
	rm -f *.o *~ core test_forks