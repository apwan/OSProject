/*
 * Test case 1.6: Cyclic reuse of fd; 1000*16 open/close fd pair, write should not write to wrong file.
 * Exit when any error occur
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
			//create file
			//don't close
			//wait some while, 
			//write
			//close
			//wait some while, read&check
			//unlink
			
			if(i>1)
			{
				sprintf(filename,"test_cyclifFD_%d_%d.tmp",i-2,j);
				check(fd=open(filename));
				sprint(buf2," ");//invalidate buffer
				equal(read(fd,buf2,20),20);
				check(close(fd));
				check(compare(buf+j,buf2,20));
				check(unlink(filename));
			}
			if(i>0)
			{
				//sprintf(filename,"test_cyclifFD_%d_%d.tmp",i-1,j);
				equal(write(fds[j],buf+j,20),20);
				check(close(fds[j]));
			}
			
			sprintf(filename,"test_cyclifFD_%d_%d.tmp",i,j);
			check(fds[j]=creat(filename));
		}
	}
	for(j=0;j<FD;j++)close(fds[j]);
	for(i=10000-2;i<10000;i++)
		for(j=0;j<FD;j++)
		{
			sprintf(filename,"test_cyclifFD_%d_%d.tmp",i,j);
			ret=unlink(filename);
			check(ret);
		}
	return 0;
}
