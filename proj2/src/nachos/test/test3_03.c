/*
 * Test case 3.3: exit(0) twice. Join a children who exit() twice
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int pid,st;
int main(int argc, char** argv)
{
	if(argc>1000)
	{
		open("test_nonexistent_file.tmp");
		exit(0);
		exit(1);
	}
	else
	{
		pid=exec("test3_03.coff",1234,argv);
		equal(join(pid,&st),1);
		equal(st,0);
		exit(0);
	}
	return -233;
}
