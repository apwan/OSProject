/*
 * Test case 3.1: Put some statement after exit(0), make sure exit(0) function normally.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int i,j,k;
int main(int argc, char** argv)
{
	i=2;
	i-=2;
	exit(i);
	i-=1;
	exit(i);	
	return -233;
}
