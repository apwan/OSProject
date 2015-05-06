/*
 * Test case 1.4: Read/write on the same time
 * Exit when mismatch
 * (will clean up files before exit)
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char buf[BUFSIZE],buf2[BUFSIZE];

int compare(char* buf,char* buf2, int size)
{
	int i;
	for(i=0;i<size;i++)
		if(buf[i]!=buf2[i])
			return -1;
	return 1;
}
char filename[20];
int fd,fd2,size,i,j,ret;
#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int main(int argc, char** argv)
{
	sprintf(filename,"test-full.tmp");
	fd=creat(filename);
	for(i=0;i<BUFSIZE;i++)buf[i]='A'+i%26;
	ret=write(fd,buf,BUFSIZE);
	check(ret);
	ret=close(fd);
	check(ret);
	
	fd=open(filename);
	fd2=open(filename);
	for(i=0;i<30;i++)
	{
		ret=read(fd,buf2+i*32,32);
		equal(ret,32);
		check(compare(buf+i*32,buf2+i*32,32));
		
		ret=write(fd2,"12345678901234567890123456789012",32);
		equal(ret,32);
	}
	close(fd);
	close(fd2);
	unlink(filename);
	return 0;
}
