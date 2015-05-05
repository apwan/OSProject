/*
 * Test case 2.1: Stress test of paging system
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}

#define N 6
int a[N][N],b[N][N],c[N][N];

int i,j,k;
int main(int argc, char** argv)
{
	for(i=0;i<N;i++)
	for(j=0;j<N;j++)
		a[i][j]=b[i][j]=i+j+1;
	
	for(i=0;i<N;i++)
	for(j=0;j<N;j++)
		c[i][j]=0;
	
	for(i=0;i<N;i++)
	for(j=0;j<N;j++)
	{
		for(k=0;k<N;k++)
			c[i][j]+=a[i][k]*b[k][j];
	}
	equal(c[0][0],2870);
	equal(c[N-1][N-1],18070);
	return 0;
}
