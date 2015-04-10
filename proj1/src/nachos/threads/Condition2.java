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
        Machine.interrupt().disable();
        // add this thread to queue
        waitQueue.waitForAccess(currentThread);
        // this thread sleep
        KThread.sleep();
        Machine.interrupt().enable();
        conditionLock.acquire();
    }

    /**
     * Wake up at most one thread sleeping on this condition variable. The
     * current thread must hold the associated lock.
     */
    public void wake() {
        Lib.assertTrue(conditionLock.isHeldByCurrentThread());

        //edited by KuLokSun on 10/4/2015
        Machine.interrupt().disable();
        KThread nextThread = waitQueue.nextThread();
        if(nextThread != null){
            nextThread.ready();
        }
        Machine.interrupt().enable();
    }

    /**
     * Wake up all threads sleeping on this condition variable. The current
     * thread must hold the associated lock.
     */
    public void wakeAll() {
        Lib.assertTrue(conditionLock.isHeldByCurrentThread());

        //edited by KuLokSun on 10/4/2015
        Machine.interrupt().disable();
        do{
            KThread nextThread = waitQueue.nextThread();
            if(nextThread != null){
                nextThread.ready();
            }else{
                break;
            }
        }while(true);
        Machine.interrupt().enable();
    }

    private Lock conditionLock;

    //edited by KuLokSun on 10/4/2015
    // Using Thread Queue in schedular
    private ThreadQueue waitQueue;
}
