/*
 * Test case 1.8: Write after close should cause exit(); unclosed handle should be closed after exit() or exception.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char buf[BUFSIZE],buf2[BUFSIZE];

char filename[20];
int fd,size,i,j,ret,pid,st;
int main(int argc, char** argv)
{
	if(argc>1000)
	{
		fd=open("test_1_08.tmp");
		close(fd);
		write(fd,buf,20);
		exit(0);
	}
	else
	{
		creat("test_1_08.tmp");
		unlink("test_1_08.tmp");
		
		pid=exec("test1_08.coff",1001,argv);
		if(join(pid,&st)!=0)exit(-1);//exception
		exit(0);
	}
	return 0;
}
