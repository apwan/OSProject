package nachos.threads;

import nachos.machine.*;

// Library for priority queue
import java.util.Comparator;  
import java.util.PriorityQueue;  

/**
 * Uses the hardware timer to provide preemption, and to allow threads to sleep
 * until a certain time.
 */
public class Alarm {
	public static void selfTest(){
		System.out.println("Alarm self test started");
		
		System.out.println("Alarm self test finished");
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
        
        Record top = waitQueue.peek();
        if (top == null){
        	return;
        }
        long currentTime = Machine.timer().getTime();
        while(top.wakeTime > currentTime){
            Record record = waitQueue.poll();
            record.thread.ready();
            Lib.debug('t', "thread "+record.thread.getName()+" wake up");
        }
        
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
