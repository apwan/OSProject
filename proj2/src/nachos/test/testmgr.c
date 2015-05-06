/*
 * Test Manager
 * Execute all test cases, output statistics.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define BUFSIZE 1024

char buf[BUFSIZE];

char filename[50];
long fd,size,i,j,ret,pid,st;
#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}

#define N 20
int res[N];
char *cases[]={
"test1_01",
"test1_02",
"test1_03",
"test1_04",
"test1_05",
"test1_06",
"test1_07",
"test1_08",
"test2_01",
"test2_02",
"test2_03",
"test2_04",
"test3_01",
"test3_02",
"test3_03",
"test3_04",
"test3_05",
"test3_06",
"test3_07",
"test3_08"
};

int main(int argc, char** argv)
{
	int succ=0,fail=0;
	for(i=0;i<N;i++)
	{
		sprintf(filename,"%s.coff",cases[i]);
		printf("Running test case: %s\n",filename);
		pid=exec(filename,0,NULL);
		ret=join(pid,&res[i]);
		if(ret!=1)res[i]=ret;
		printf("Result: %d\n",res[i]);
		if(res[i]==0)succ++;
		else fail++;
	}
	printf("===\nTotal:%d\nSucc:%d\nFail:%d\n",N,succ,fail);
	for(i=0;i<N;i++)
	if(res[i]!=0)
	{
		printf("Failed:%s (%d)",cases[i],res[i]);
	}
	return 0;
}
