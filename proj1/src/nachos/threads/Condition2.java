package nachos.threads;

import nachos.machine.*;

/**
 * An implementation of condition variables that disables interrupt()s for
 * synchronization.
 *
 * <p>
 * You must implement this.
 *
 * @see nachos.threads.Condition
 */
public class Condition2 {
    /**
     * Allocate a new condition variable.
     *
     * @param   conditionLock   the lock associated with this condition
     *                          variable. The current thread must hold this
     *                          lock whenever it uses <tt>sleep()</tt>,
     *                          <tt>wake()</tt>, or <tt>wakeAll()</tt>.
     */
    public Condition2(Lock conditionLock) {
        this.conditionLock = conditionLock;

        //edited by KuLokSun on 10/4/2015
        //  using the queue in scheduler to solve priority inversion problem
        waitQueue = ThreadedKernel.scheduler.newThreadQueue(true);
    }

    /**
     * Atomically release the associated lock and go to sleep on this condition
     * variable until another thread wakes it using <tt>wake()</tt>. The
     * current thread must hold the associated lock. The thread will
     * automatically reacquire the lock before <tt>sleep()</tt> returns.
     */
    public void sleep() {
        Lib.assertTrue(conditionLock.isHeldByCurrentThread());

        //edited by KuLokSun on 10/4/2015
        // get this thread from the lock
        KThread currentThread = KThread.currentThread();
        conditionLock.release();
        boolean intStatus = Machine.interrupt().disable();
        // add this thread to queue
        waitQueue.waitForAccess(currentThread);
        // this thread sleep
        KThread.sleep();
        Machine.interrupt().restore(intStatus);
        conditionLock.acquire();
    }

    /**
     * Wake up at most one thread sleeping on this condition variable. The
     * current thread must hold the associated lock.
     */
    public void wake() {
        Lib.assertTrue(conditionLock.isHeldByCurrentThread());

        //edited by KuLokSun on 10/4/2015
        boolean intStatus = Machine.interrupt().disable();
        KThread nextThread = waitQueue.nextThread();
        if(nextThread != null){
            nextThread.ready();
        }
        Machine.interrupt().restore(intStatus);;
    }

    /**
     * Wake up all threads sleeping on this condition variable. The current
     * thread must hold the associated lock.
     */
    public void wakeAll() {
        Lib.assertTrue(conditionLock.isHeldByCurrentThread());

        //edited by KuLokSun on 10/4/2015
        boolean intStatus = Machine.interrupt().disable();
        do{
            KThread nextThread = waitQueue.nextThread();
            if(nextThread != null){
                nextThread.ready();
            }else{
                break;
            }
        }while(true);
        Machine.interrupt().restore(intStatus);;
    }
    

    

    private static class CaseTester1 implements Runnable {
    	private int tcid;
    	CaseTester1()
    	{
    		tcid=TestMgr.addTest("Condition Case Test 1: DB RW");
    	}
    	private static class DB {
    		int violation=0;
    		int legitimate_user_count=0;
    		int reading=0;
    		int writing=0;
    		void sR(){reading++;if(writing>0)violation++;}
    		void eR(){reading--;if(reading<0)legitimate_user_count++;}
    		void sW(){writing++;if(writing>1)violation++;}
    		void eW(){writing--;if(writing!=0)legitimate_user_count++;}
    	}
    	private static DB db;
    	static Lock lock;
    	static Condition2 okToRead,okToWrite;
    	static int AR,AW,WR,WW;
    	static int done=0;
    	private static class CaseTester1Reader implements Runnable {
    		public void run() {
				lock.acquire();
				while ((AW + WW) > 0) {	// Is it safe to read?
				WR++;	// No. Writers exist
				okToRead.sleep();
				WR--;	// No longer waiting
				}
				AR++;		// Now we are active!
				lock.release();
					// Perform actual read-only access
				
				db.sR();
				System.out.println("start reading db...");
				for(int i=0;i<2;i++)KThread.currentThread().yield();
				db.eR();
				
					//AccessDatabase(ReadOnly);
					// Now, check out of system
				lock.acquire();
				AR--;		// No longer active
				if (AR == 0 && WW > 0)	// No other active readers
					okToWrite.wake();	// Wake up one writer
				lock.release();
				done++;
				System.out.println("Reader I'm done; now done:"+done);
    		}
    	}
    	private static class CaseTester1Writer implements Runnable {
    		public void run() {
				// First check self into system
				lock.acquire();
					while ((AW + AR) > 0) {	// Is it safe to write?
					WW++;	// No. Active users exist
					okToWrite.sleep();
					WW--;	// No longer waiting
				}
					AW++;		// Now we are active!
				lock.release();
					// Perform actual read/write access
				db.sW();
				System.out.println("start writing db...");
				for(int i=0;i<10;i++)KThread.currentThread().yield();
				db.eW();
				
					// Now, check out of system
				lock.acquire();
				AW--;		// No longer active
				if (WW > 0){	// Give priority to writers
					okToWrite.wake();	// Wake up one writer
				} else if (WR > 0) {	// Otherwise, wake reader
					okToRead.wakeAll();	// Wake all readers
				}	
				lock.release();
				done++;
				System.out.println("Writer I'm done; now done:"+done);
    		}
    	}
    	
    	public void run() {
    		db=new DB();
    		lock=new Lock();
    		okToRead=new Condition2(lock);
    		okToWrite=new Condition2(lock);
    		AR=AW=WR=WW=0;
    		final int N=10;
    		done=0;
    		Runnable[] a=new Runnable[N];
    		for(int i=0;i<N;i++)
    		{
    			if(i%4==3)
    				a[i]=new CaseTester1Writer();
    			else
    				a[i]=new CaseTester1Reader();
    			new KThread(a[i]).setName("ct1-tr"+i).fork();
    		}
    		
    		KThread.currentThread().yield();
    		do
    		{
    			System.out.println("check rw#:"+AR+","+AW+","+WR+","+WW+" done:"+done);
    			KThread.currentThread().yield();
    		}while(AR+AW+WR+WW>0 || done<N*0.7);
    		
    		TestMgr.finishTest(tcid, db.violation==0);
    		int tc2id=TestMgr.addTest("DB #user is consistent");
    		TestMgr.finishTest(tc2id, db.legitimate_user_count==0);
    		int tc3id=TestMgr.addTest("All DB user is finished.");
    		System.out.println("done vs n:"+done+":"+N);
    		TestMgr.finishTest(tc3id, done==N);
    	}
    }
    
    private static class CaseTester2 implements Runnable {
    	private int tcid;
    	CaseTester2()
    	{
    		tcid=TestMgr.addTest("Condition Case Test 2: Bank deposition");
    	}
    	static int done;
    	static Lock lock;
    	static Condition2 condition;
    	static int balance;
    	static int atomicBalance;
    	static int balanceViolation=0;
    	private static class CaseTester2Changer implements Runnable{
    		int amount;
    		CaseTester2Changer(int a){
    			amount=a;
    			if(a<0)
    				System.out.println("I'm a remover, amount"+(-a));
    			else
    				System.out.println("I'm an adder, amount"+(a));
    		}
    		public void run(){
    			lock.acquire();
    			if(amount<0)
    			{
    				while(balance+amount<0)
	    			{
	    				condition.sleep();
	    			}
	    		}
    			int tmp=balance;
    			//KThread.currentThread().yield();
    			int newbalance=tmp+amount;
    			System.out.println("oldbanalce:"+tmp+" -> new"+newbalance);
    			atomicBalance+=amount;
    			KThread.currentThread().yield();
//    			balance=amount; Seems Logic Error, It should be atomicBalance
    			balance = atomicBalance;
    			if(balance<0)balanceViolation++;
    			if(amount>0)condition.wakeAll();
    			lock.release();
    			System.out.println("adder "+amount+" has done. current balance:"+balance);
    			done++;
    		}
    	}
    	
    	public void run()
    	{
    		lock=new Lock();
    		condition=new Condition2(lock);
    		done=0;
    		//CaseTester2Changer[] a=new CaseTester2Changer[20];
    		int n=0, i=0;
    		//a[i++]=
    		new KThread(new CaseTester2Changer(50)).setName("bank client#"+(i++)).fork();
    		new KThread(new CaseTester2Changer(-20)).setName("bank client#"+(i++)).fork();
    		new KThread(new CaseTester2Changer(-20)).setName("bank client#"+(i++)).fork();
    		new KThread(new CaseTester2Changer(-20)).setName("bank client#"+(i++)).fork();
    		new KThread(new CaseTester2Changer(50)).setName("bank client#"+(i++)).fork();
    		//new KThread(new CaseTester2Changer(50)).setName("bank client#"+(i++)).fork();
    		n=i;int step=0;
    		do{
    			System.out.println("Current finished adder #"+done+", step"+step);
    			KThread.currentThread().yield();
    			step++;
    		}while(done!=n && step<n*50);
    		TestMgr.finishTest(tcid);
    		tcid=TestMgr.addTest("Bank: Atomic Balance == Actual Operated Balance");
    		TestMgr.finishTest(tcid, atomicBalance==balance);
    		tcid=TestMgr.addTest("Bank: balance<0 violation");
    		TestMgr.finishTest(tcid, balanceViolation==0);
    	}
    }

	private static class CaseTester3 implements Runnable {
    	private int tcid;
    	CaseTester3()
    	{
    		tcid=TestMgr.addTest("Condition Case Test 3: Overslept");
    	}
    	static int done;
    	static Lock lock;
    	static Condition2 condition, altcond;
    	static int ok=0;
    	static final int N=5;
    	private static class CaseTester3Waiter implements Runnable{
    		public void run(){
    			lock.acquire();
    			while(ok==0)
    			{
    				condition.sleep();
    			}
    			KThread.currentThread().yield();
    			done++;        		
    			altcond.wake();
    			while(ok==1)
    			{
    				condition.sleep();
    			}
    			lock.release();
    			done++;
    		}
    	}
    	private static class CaseTester3Allower implements Runnable{
    		public void run(){
    			lock.acquire();
    			KThread.currentThread().yield();
    			ok++;
    			condition.wakeAll();
    			while(done<N)
    			{
    				altcond.sleep();
    			}
    			ok++;
    			condition.wakeAll();
    			lock.release();
    		}
    	}
	
    	public void run()
    	{
    		lock=new Lock();
    		condition=new Condition2(lock);altcond=new Condition2(lock);
    		done=0;
    		for(int i=0;i<N;i++)
    			new KThread(new CaseTester3Waiter()).setName("waiter#"+(i)).fork();
    		new KThread(new CaseTester3Allower()).setName("upd##").fork();
    		
    		//new KThread(new CaseTester2Changer(50)).setName("bank client#"+(i++)).fork();
    		int step=0;
    		do{
    			System.out.println("Current finished waiter stage#"+done+", step"+step);
    			KThread.currentThread().yield();
    			step++;
    		}while(done!=2*N && step<N*50);
    		TestMgr.finishTest(tcid,done==2*N);
    	}
    }
    
    

    
    public static void selfTest(){
    	//Lock lock = new Lock();
    	//Condition2 condition2 = new Condition2(lock);
    	// implement the test case!
    	
    	KThread k;
    	k=new KThread(new CaseTester1());
    	k.setName("condition CT1").fork();
    	k.join();
    	k=new KThread(new CaseTester2());
    	k.setName("condition CT2").fork();
    	k.join();
    	k=new KThread(new CaseTester3());
    	k.setName("condition CT3").fork();
    	k.join();
    	
    }

    private Lock conditionLock;

    //edited by KuLokSun on 10/4/2015
    // Using Thread Queue in schedular
    private ThreadQueue waitQueue;
}
