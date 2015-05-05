/*
 * Test case 1.9: Unlink the file while another process opened and is reading the file.
 * (will clean up files before exit)
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
