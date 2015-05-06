/*
 * Test case 3.6: A children is joined and then caused exception; parent should continue normally.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int i,j,k,st,pid;
int main(int argc, char** argv)
{
	if(argc>1000)
	{
		open("test_nonexistent_file.tmp");
		i=9;
		j=8;
		i-=j;
		i-=1;
		j/=i;
		exit(0);
	}
	else
	{
		pid=exec("test3_06.coff",1234,argv);
		equal(join(pid,&st),0);//unhandled exception:0
		exit(0);
	}
	return 0;
}
