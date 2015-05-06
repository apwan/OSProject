/*
 * Test case 1.7: Different program opens different file.
 * NOTE: this test case requires using EXEC/JOIN syscall.
 * (will clean up files before exit)
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024
#define INSTCNT 1000 
char buf[BUFSIZE],buf2[BUFSIZE];
int pids[INSTCNT];

int compare(char* buf, char* buf2, int size)
{
	int i;
	for(i=0;i<size;i++)
		if(buf[i]!=buf2[i])
			return 0;
	return 1;
}
char filename[20];
int fd,size,i,j,ret,st;
#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int main(int argc, char** argv)
{
	for(i=0;i<BUFSIZE;i++)buf[i]='A'+i%26;
	
	if(argc<INSTCNT)//i'm manager
	{
		//writer argc: INSTCNT+id
		//reader argc: INSTCNT*2+id
		for(i=0;i<INSTCNT;i++)
			check(pids[i]=exec("test1_07.coff",INSTCNT+i,argv));
		ret=0;
		for(i=0;i<INSTCNT;i++)
			ret+=join(pids[i],&st);
		if(ret<=0)return ret;
		
		for(i=0;i<INSTCNT;i++)
			check(pids[i]=exec("test1_07.coff",INSTCNT*2+i,argv));
		ret=0;
		for(i=0;i<INSTCNT;i++)
			ret+=join(pids[i],&st);
		if(ret<=0)return ret;
		
		for(i=0;i<INSTCNT;i++)
		{
			sprintf(filename,"test-rwi%d.tmp",i);
			check(unlink(filename));
		}
	}
	else if(argc<2*INSTCNT)//i'm writer
	{
		sprintf(filename,"test-rwi%d.tmp",argc%INSTCNT);
		check(fd=open(filename));
		equal(write(fd,buf+(i%20),20),20);
		check(close(fd));
	}
	else //i'm reader
	{
		sprintf(filename,"test-rwi%d.tmp",argc%INSTCNT);
		check(fd=open(filename));
		check(read(fd,buf2,20));
		check(compare(buf2,buf+(i%20),20));
		check(close(fd));
	}
	return 0;
}
