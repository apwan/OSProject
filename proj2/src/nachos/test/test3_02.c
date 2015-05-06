/*
 * Test case 3.2: Joining a children twice.
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
		if(argc>2000)return 234;
		exit(233);
	}
	else
	{
		pid=exec("test3_02.coff",1234,argv);
		equal(join(pid,&st),1);
		equal(st,233);
		equal(join(pid,&st),-1);
		
		pid=exec("test3_02.coff",2345,argv);
		equal(join(pid,&st),1);
		equal(st,234);		
	}
	return 0;
}
