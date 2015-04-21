package nachos.threads;

import nachos.machine.*;

// Library for priority queue
import java.util.Comparator;  
import java.util.PriorityQueue;  
import java.util.Random;

/**
 * Uses the hardware timer to provide preemption, and to allow threads to sleep
 * until a certain time.
 */
public class Alarm {
    

    private static class CaseTester0 implements Runnable{
        private int tcid;
        CaseTester0()
        {
            tcid=TestMgr.addTest("alarm Case Test 1: sleep&wake");
        }
        public void run() {
            long time=Machine.timer().getTime();
            Alarm.getInstance().waitUntil(2);
            long time2=Machine.timer().getTime();
            TestMgr.finishTest(tcid, time2-time>=2);
        }
    }
    
    private static class CaseTester1 implements Runnable{
        private int tcid;
        CaseTester1()
        {
            tcid=TestMgr.addTest("alarm Case Test 1: sleepsort");
        }
        private static class CaseTester1sort implements Runnable{
            int val;
            CaseTester1sort(int d){val=d;}
            public void run() {
                Alarm.getInstance().waitUntil(val);
                glo[pos++]=val;
            }
        }
        static int glo[];
        static int pos;
        public void run() {
            int N=10;
            int Max=20;
            int a[]=new int[N];
            glo=new int[N];
            pos=0;
            Random rand = new Random();
            for(int i=0;i<N;i++)
                a[i]=rand.nextInt(Max);
            for(int i=0;i<N;i++)
                new KThread(new CaseTester1sort(a[i])).fork();
            
            for(int i=0;i<N-1;i++)
                for(int j=i;j<N;j++)
                    if(a[i]>a[j])
                    {
                        int t=a[i];a[i]=a[j];a[j]=t;
                    }

            Alarm.getInstance().waitUntil(Max);
            boolean cond=true;
            for(int i=0;i<N;i++)
                if(a[i]!=glo[i])cond=false;
            TestMgr.finishTest(tcid, cond);
        }        
    }
    
    public static void selfTest(){
        System.out.println("Alarm self test started");
        KThread k;
        k=new KThread(new CaseTester0());
        k.setName("alarm CT0").fork();
        k.join();
        System.out.println("Alarm self test finished");
    }
    /**
     * Only allow a unique instance
     * 
     */
    private static Alarm instance=null;
    public static Alarm getInstance()
    {
        if(instance==null)
            instance=new Alarm();
        return instance;
    }
    /**
     * Allocate a new Alarm. Set the machine's timer interrupt handler to this
     * alarm's callback.
     *
     * <p><b>Note</b>: Nachos will not function correctly with more than one
     * alarm.
     */
    public Alarm() {
        Machine.timer().setInterruptHandler(new Runnable() {
            public void run() { timerInterrupt(); }
        });
        // create a comparator for priority queue
        Comparator<Record> comparator = new Comparator<Record>() {  
            public int compare(Record a, Record b) {  
                if (a.wakeTime < b.wakeTime ){
                    return -1;
                } else if (a.wakeTime == b.wakeTime && a.sleepTime <= b.sleepTime) {
                    return -1;
                   
                }
                return 1;
            }  
        };
        // create priority queue, 128 is just a hint, it will auto resize
        if (waitQueue == null){
            waitQueue = new PriorityQueue<Record>(128, comparator); 
        }
    }

    // A private class store (thread, time)for priority queue
    private class Record {
        public Record(KThread kthread, long sleepTime, long wakeTime){
            this.thread = kthread;
            this.sleepTime = sleepTime;
            this.wakeTime = wakeTime;
        }
        public KThread thread;
        private long sleepTime;
        public long wakeTime;
    }

    /**
     * The timer interrupt handler. This is called by the machine's timer
     * periodically (approximately every 500 clock ticks). Causes the current
     * thread to yield, forcing a context switch if there is another thread
     * that should be run.
     */
    public void timerInterrupt() {

        //We don't need to disable interrupt
        
        
        long currentTime = Machine.timer().getTime();
        Record record;
        do{
            record = waitQueue.peek();
            if (record == null || record.wakeTime > currentTime){
                break;
            }
            waitQueue.remove();
            record.thread.ready();
            Lib.debug('t', "thread "+record.thread.getName()+" wake up");
        }while(true);
        
    }

    /**
     * Put the current thread to sleep for at least <i>x</i> ticks,
     * waking it up in the timer interrupt handler. The thread must be
     * woken up (placed in the scheduler ready set) during the first timer
     * interrupt where
     *
     * <p><blockquote>
     * (current time) >= (WaitUntil called time)+(x)
     * </blockquote>
     *
     * @param   x       the minimum number of clock ticks to wait.
     *
     * @see     nachos.machine.Timer#getTime()
     */
    public void waitUntil(long x) {
        // for now, cheat just to get something working (busy waiting is bad)
        // long wakeTime = Machine.timer().getTime() + x;
        // while (wakeTime > Machine.timer().getTime())
        //     KThread.yield();
        
        boolean intStatus = Machine.interrupt().disable();

        // create object which is going to be added into queue
        KThread currentThread = KThread.currentThread();
        long sleepTime = Machine.timer().getTime();
        long wakeTime = sleepTime + x;
        Record p = new Record(currentThread, sleepTime, wakeTime);
        Lib.debug('t', "alarm set by thread "+ KThread.currentThread().getName()+", timing "+x);
        
        // add into queue        
        waitQueue.add(p);
        // *** Tutor ask why we do not sleep on conditional revariable.
        KThread.sleep();

        Machine.interrupt().restore(intStatus);
    }

    private PriorityQueue<Record> waitQueue;
}
