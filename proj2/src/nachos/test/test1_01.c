/*
 * Test case 1.1: filling up the disk
 * Exit when mismatch
 * (will clean up files before exit)
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char buf[BUFSIZE],buf2[BUFSIZE];

int compare(int size)
{
	int i;
	for(i=0;i<size;i++)
		if(buf[i]!=buf2[i])
			return 0;
	return 1;
}
char filename[20];
int fd,size,i,j,ret;
int main(int argc, char** argv)
{
	sprintf(filename,"test-full.tmp");
	fd=creat(filename);
	for(i=0;i<BUFSIZE;i++)buf[i]='A'+i%26;
	
	for(i=0;i<10000;i++)
	{
		size=i%BUFSIZE;
		ret=write(fd,buf,size);
		if(ret<size)break;//disk filled.
	}
	ret=close(fd);
	if(ret<0)return -1;
	
	fd=open(filename);
	for(j=0;j<i;j++)
	{
		size=j%BUFSIZE;
		sprintf(buf2," ");//touch the buffer
		ret=read(fd,buf2,size);
		if(ret!=size)return -1;
		if(!compare(size))return -1;
	}
	close(fd);
	unlink(filename);
	return 0;
}
