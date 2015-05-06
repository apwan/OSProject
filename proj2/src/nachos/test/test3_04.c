/*
 * Test case 3.4: exec() then exit parent; children should run normally.
*/

#include "syscall.h"
#include "stdio.h"
#include "stdlib.h"

#define check(ret) {if(ret<0)return -1;}
#define equal(ret,exp) {if(ret!=exp)return -1;}
int pid,fd,i;
char buf[20];
int main(int argc, char** argv)
{
	if(argc>2000)
	{
		fd=open("test_3_4_footprint.tmp");
		open("test_3_4_nonexistent1.tmp");
		open("test_3_4_nonexistent2.tmp");
		open("test_3_4_nonexistent3.tmp");
		open("test_3_4_nonexistent4.tmp");
		open("test_3_4_nonexistent5.tmp");
		open("test_3_4_nonexistent6.tmp");
		open("test_3_4_nonexistent7.tmp");
		buf[0]='1';
		write(fd,buf,20);
		close(fd);
		open("test_3_4_nonexistent8.tmp");
		open("test_3_4_nonexistent9.tmp");
	}
	else if(argc>1000)
	{
		pid=exec("test3_02.coff",2234,argv);
		exit(0);
	}
	else
	{
		fd=creat("test_3_4_footprint.tmp");
		buf[0]='0';
		write(fd,buf,1);
		close(fd);
		
		pid=exec("test3_02.coff",1234,argv);
		i=0;
		while(1)
		{
			fd=open("test_3_4_footprint.tmp");
			read(fd,buf,1);
			close(fd);
			if(buf[0]=='1')exit(0);
			if(i++>1000)exit(-1);
		}
	}
	return -1;
}
