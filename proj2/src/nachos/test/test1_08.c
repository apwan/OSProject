/*
 * Test case 1.8: Write after close should cause exit(); unclosed handle should be closed after exit() or exception.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char buf[BUFSIZE],buf2[BUFSIZE];

char filename[20];
int fd,size,i,j,ret;
int main(int argc, char** argv)
{
	
	return 0;
}
