/*
 * Test case 3.7: Run multiple processes. Let them call exit(0) in random order and check whether the machine terminates after the last process exits.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int i,j,k,fd,st;
int pids[20];
int ranInt()
{
	int i,r=0;
	for(i=0;i<16;i++)
		r+=rand()<<i;
	return r;
}

int main(int argc, char** argv)
{
	if(argc>1000)
	{
		k=ranInt()%100;
		for(i=0;i<=k+1;i++)
		{
			fd=open("test_3_07_delayer.tmp");//will sleep some while
			close(fd);
		}
		exit(argc%1000);
	}
	else
	{
		creat("test_3_07_delayer.tmp");
		for(i=0;i<1000;i++)
			pids[i]=exec("test3_07.coff",1000+i,argv);
		for(i=0;i<1000;i++)
		{		
			equal(join(pids[i],&st),1);
			equal(st,i);
		}
		
		
		unlink("test_3_07_delayer.tmp");
		exit(0);
	}
	return 0;
}
