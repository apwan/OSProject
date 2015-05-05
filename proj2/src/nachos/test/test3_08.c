/*
 * Test case 3.8: Illegal EXEC
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int i,j,k;
int main(int argc, char** argv)
{
	equal(exec("test-nonexistent-file-ASDFASD.coff",1234,argv),-1);	
	equal(exec("test3_08.c",2345,argv),-1);	
	return 0;
}
