/*
 * Test case 3.5: fork and join forming a chain; after 1000 level, the final children exit(), then the greatest parent should exit normally, at last.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int i,j,pid,st;
int main(int argc, char** argv)
{
	if(argc<1000)
	{
		pid=exec("test3_05.coff",argc+1,argv);
		equal(join(pid,&st),1);
		equal(st,0);
	}
	else exit(0);
	return 0;
}
