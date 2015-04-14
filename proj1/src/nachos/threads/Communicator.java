package nachos.threads;

import java.util.Random;

import nachos.machine.*;

/**
 * A <i>communicator</i> allows threads to synchronously exchange 32-bit
 * messages. Multiple threads can be waiting to <i>speak</i>,
 * and multiple threads can be waiting to <i>listen</i>. But there should never
 * be a time when both a speaker and a listener are waiting, because the two
 * threads can be paired off at this point.
 */
public class Communicator {
	static int sent;
	static int rcvd;
	static int rcv_acc;
	static Communicator CTcomm; 
	private static class CaseTestHelperRecv implements Runnable{
		public void run() {
			int a=CTcomm.listen();
			rcvd++;
			rcv_acc+=a;
		}
	}
	private static class CaseTestHelperSend implements Runnable{
		private int msg;
		CaseTestHelperSend(int a){msg=a;}
		public void run() {
			CTcomm.speak(msg);
			sent++;
		}
	}
	private static class CaseTester0 implements Runnable{
    	private int tcid;
    	CaseTester0()
    	{
    		tcid=TestMgr.addTest("Communicator Case Test 1: 1 on 1 (no deadlock either direction)");
    	}
		public void run() {
			CTcomm=new Communicator();

			KThread t1=new KThread(new CaseTestHelperSend(5));
			t1.fork();
			KThread.currentThread().yield();
			KThread t2=new KThread(new CaseTestHelperRecv());
			t2.fork();
			t2.join();t1.join();
			
			
			KThread t3=new KThread(new CaseTestHelperRecv());
			t3.fork();
			KThread.currentThread().yield();
			KThread t4=new KThread(new CaseTestHelperSend(5));
			t4.fork();
			t4.join();t3.join();
			
			TestMgr.finishTest(tcid);
		}
	}
	
	private static class CaseTester1 implements Runnable{
    	private int tcid;
    	CaseTester1()
    	{
    		tcid=TestMgr.addTest("Communicator Case Test 1: 5 on 2");
    	}
		public void run() {
			CTcomm=new Communicator();
			sent=rcvd=0;
			rcv_acc=0;
			int snd_acc=0;
			KThread senders[]=new KThread[5];
			for(int i=0;i<5;i++)
			{
				senders[i]=new KThread(new CaseTestHelperSend(i));
				senders[i].setName("ComCT1#"+i).fork();
				snd_acc+=i;
			}
			KThread t1=new KThread(new CaseTestHelperRecv());
			t1.fork();
			KThread t2=new KThread(new CaseTestHelperRecv());
			t2.fork();
			t1.join();t2.join();
			
			TestMgr.finishTest(tcid, sent==2 && rcvd==2);
		}
	}
	
	private static class CaseTester2 implements Runnable{
    	private int tcid;
    	CaseTester2()
    	{
    		tcid=TestMgr.addTest("Communicator Case Test 2: 2 on 5");
    	}
		public void run() {
			CTcomm=new Communicator();
			sent=rcvd=0;
			rcv_acc=0;
			KThread senders[]=new KThread[5];
			for(int i=0;i<5;i++)
			{
				senders[i]=new KThread(new CaseTestHelperRecv());
				senders[i].setName("ComCT2#"+i).fork();
			}
			KThread t1=new KThread(new CaseTestHelperSend(3));
			t1.fork();
			KThread t2=new KThread(new CaseTestHelperSend(6));
			t2.fork();
			t1.join();t2.join();
			
			TestMgr.finishTest(tcid, sent==2 && rcvd==2 && rcv_acc==3+6);
		}
	}

	private static class CaseTester3 implements Runnable{
    	private int tcid;
    	CaseTester3()
    	{
    		tcid=TestMgr.addTest("Communicator Case Test 3: 50 on 50");
    	}
		public void run() {
			CTcomm=new Communicator();
			sent=rcvd=0;
			rcv_acc=0;
			int snd_acc=0;
			int cnts=0,cntr=0;
			Random rnd=new Random();
			KThread actors[]=new KThread[100];
			boolean hasForked[]=new boolean[100];
			for(int i=0;i<100;i++)
			{
				int val=rnd.nextInt(100);
				if(i%2==0)
				{
					actors[i]=new KThread(new CaseTestHelperSend(val));
					actors[i].setName("ComCT3 snd#"+i);
					snd_acc+=val;
					
				}
				else
				{
					actors[i]=new KThread(new CaseTestHelperRecv());
					actors[i].setName("ComCT3 rcv#"+i);
				}
			}
			for(int i=0;i<103;i++)
			{
				int pos=(i*130)%103;
				if(pos>=100)continue;
				actors[pos].fork();
			}
			
			for(int i=0;i<100;i++)
				actors[i].join();
			
			TestMgr.finishTest(tcid, rcv_acc == snd_acc && sent==rcvd && rcvd==50);
		}
		private int Min(int a,int b) {
			return a<b?a:b;
		}
	}
	
	
	public static void selfTest(){
		System.out.println("Communicator self test started");
    	KThread k;
    	k=new KThread(new CaseTester0());
    	k.setName("Communicator CT0").fork();
    	k.join();//cannot join due to the bug of joining finished thread.
    	k=new KThread(new CaseTester1());
    	k.setName("Communicator CT1").fork();
    	k.join();
    	k=new KThread(new CaseTester2());
    	k.setName("Communicator CT2").fork();
    	k.join();
    	k=new KThread(new CaseTester3());
    	k.setName("Communicator CT3").fork();
    	k.join();
		System.out.println("Communicator self test finished");
	}
	
	
    /**
     * Allocate a new communicator.
     */
    public Communicator() {
        lock = new Lock();
        sendCV = new Condition2(lock);
        receiveCV = new Condition2(lock);
    }

    /**
     * Wait for a thread to listen through this communicator, and then transfer
     * <i>word</i> to the listener.
     *
     * <p>
     * Does not return until this thread is paired up with a listening thread.
     * Exactly one listener should receive <i>word</i>.
     *
     * @param   word    the integer to transfer.
     */
    public void speak(int word) {
        lock.acquire();
        // ++senderCount; // seems useless
        while(wordReady || receiverCount == 0){
            receiveCV.wakeAll();
            sendCV.sleep();
        }
        
        wordBuffer = word;
        wordReady = true;
        // --senderCount; // seems useless
        // Added by KuLokSun
        // wake all receiver to get the word
        receiveCV.wakeAll();
        // end add
        lock.release();
    }

    /**
     * Wait for a thread to speak through this communicator, and then return
     * the <i>word</i> that thread passed to <tt>speak()</tt>.
     *
     * @return  the integer transferred.
     */    
    public int listen() {
        lock.acquire();
        ++receiverCount;
        while(!wordReady){
            sendCV.wakeAll();
            receiveCV.sleep();
        }
        int ret = wordBuffer;
        wordReady = false;
        
        --receiverCount;
        // Added by KuLokSun
        // wake all receiver and sender for next round
        receiveCV.wakeAll(); // make recevierCount > 0
        sendCV.wakeAll(); // let sender send the message
        // end add
        lock.release();

        return ret;
    }
    
    Lock lock = null;
    private Condition2 sendCV = null, 
            receiveCV = null;
    private int senderCount = 0, receiverCount = 0;
    // we have no public method to know whether there are senders/listeners
    
    private int wordBuffer;
    private boolean wordReady = false;
}
