/*
 * Test case 1.5: Cyclic reuse of disk; shall not full if file is unlinked properly. 100000*16 file creat/unlink.
 * Exit when creat error
 * (will clean up files before exit)
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char buf[BUFSIZE],buf2[BUFSIZE];

char filename[50];
#define FD 13
int fds[FD];
long fd,size,i,j,ret;
#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int main(int argc, char** argv)
{
	for(i=0;i<10000;i++)
	{
		for(j=0;j<FD;j++)
		{
			if(i>0)
			{
				ret=close(fds[j]);
				check(ret);
			}
			sprintf(filename,"test_cyclifFD_%d_%d.tmp",i,j);
			fds[j]=open(filename);
			check(fds[j]);		
			ret=write(fds[j],buf+j,BUFSIZE-j);
			equal(ret,BUFSIZE-j);
			ret=unlink(filename);
			check(ret);
		}
	}
	for(j=0;j<FD;j++)
	{
		ret=close(fds[j]);
		check(ret);
	}
	return 0;
}
