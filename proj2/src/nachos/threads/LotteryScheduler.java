package nachos.threads;

import nachos.machine.*;

import java.util.Enumeration;
import java.util.Hashtable;
import java.util.Iterator;
import java.util.LinkedList;

/**
 * A scheduler that chooses threads using a lottery.
 *
 * <p>
 * A lottery scheduler associates a number of tickets with each thread. When a
 * thread needs to be dequeued, a random lottery is held, among all the tickets
 * of all the threads waiting to be dequeued. The thread that holds the winning
 * ticket is chosen.
 *
 * <p>
 * Note that a lottery scheduler must be able to handle a lot of tickets
 * (sometimes billions), so it is not acceptable to maintain state for every
 * ticket.
 *
 * <p>
 * A lottery scheduler must partially solve the priority inversion problem; in
 * particular, tickets must be transferred through locks, and through joins.
 * Unlike a priority scheduler, these tickets add (as opposed to just taking
 * the maximum).
 */
public class LotteryScheduler extends Scheduler {
    /**
     * Allocate a new lottery scheduler.
     */
    public LotteryScheduler() {
    }

    /**
     * Allocate a new lottery thread queue.
     *
     * @param   transferPriority        <tt>true</tt> if this queue should
     *                                  transfer tickets from waiting threads
     *                                  to the owning thread.
     * @return  a new lottery thread queue.
     */
    public ThreadQueue newThreadQueue(boolean transferPriority) {
        // implement me
        return new LotteryQueue(transferPriority);
    }

        public int getPriority(KThread thread) {
        Lib.assertTrue(Machine.interrupt().disabled());

        return getThreadState(thread).getPriority();
    }

    public int getEffectivePriority(KThread thread) {
        Lib.assertTrue(Machine.interrupt().disabled());

        return getThreadState(thread).getEffectivePriority();
    }

    public void setPriority(KThread thread, int priority) {
        Lib.assertTrue(Machine.interrupt().disabled());

        Lib.assertTrue(priority >= priorityMinimum &&
                priority <= priorityMaximum);

        getThreadState(thread).setPriority(priority);
    }

    public boolean increasePriority() {
        boolean intStatus = Machine.interrupt().disable();

        KThread thread = KThread.currentThread();

        int priority = getPriority(thread);
        if (priority == priorityMaximum)
            return false;

        setPriority(thread, priority+1);

        Machine.interrupt().restore(intStatus);
        return true;
    }

    public boolean decreasePriority() {
        boolean intStatus = Machine.interrupt().disable();

        KThread thread = KThread.currentThread();

        int priority = getPriority(thread);
        if (priority == priorityMinimum)
            return false;

        setPriority(thread, priority-1);

        Machine.interrupt().restore(intStatus);
        return true;
    }


    /**
     * Return the scheduling state of the specified thread.
     *
     * @param   thread  the thread whose scheduling state to return.
     * @return  the scheduling state of the specified thread.
     */
    protected ThreadState getThreadState(KThread thread) {
        if (thread.schedulingState == null)
            thread.schedulingState = new ThreadState(thread);

        return (ThreadState) thread.schedulingState;
    }

    /**
     * The default priority for a new thread. Do not change this value.
     */
    public static final int priorityDefault = 1;
    /**
     * The minimum priority that a thread can have. Do not change this value.
     */
    public static final int priorityMinimum = 1;
    /**
     * The maximum priority that a thread can have. Do not change this value.
     */
    public static final int priorityMaximum = Integer.MAX_VALUE;


    // Modified from the suggested solution, by Ku Lok Sun
    protected class LotteryQueue extends ThreadQueue{

        LotteryQueue() {
        }

        LotteryQueue(boolean transferPriority) {
            this.transferPriority = transferPriority;
            //waitQueue = new LinkedList<ThreadState>();
            resAccessing = null;
            counter = 0;
            donation = 0;
        }

        public void waitForAccess(KThread thread) {
            Lib.assertTrue(Machine.interrupt().disabled());
            getThreadState(thread).waitForAccess(this , counter++);
        }

        public void acquire(KThread thread) {
            Lib.assertTrue(Machine.interrupt().disabled());
            if (!transferPriority)
                return;
            getThreadState(thread).acquire(this);
            resAccessing = getThreadState(thread);
        }   
        
        public KThread nextThread() {
            Lib.assertTrue(Machine.interrupt().disabled());
            
            /** nextThread acquires resource */
            ThreadState nextThreadState = pickNextThread();
            if (resAccessing != null) {
                resAccessing.release(this);
                resAccessing = null;
            }
            
            if (nextThreadState == null) 
                return null;
            
            waitQueue.remove(nextThreadState);
            nextThreadState.ready();
            
            donatePriority();
            acquire(nextThreadState.thread);
            
            return nextThreadState.thread;
        }

        /**
         * Return the next thread that <tt>nextThread()</tt> would return,
         * without modifying the state of this queue.
         *
         * @return  the next thread that <tt>nextThread()</tt> would
         *      return.
         */
        protected ThreadState pickNextThread() {
            // written by Ku Lok Sun
            int lotterySum = 0;

            ThreadState ts = null;
            Iterator it = waitQueue.iterator();
            while (it.hasNext()){
                ts = (ThreadState)it.next();
                if(priorityMaximum - ts.getEffectivePriority() < lotterySum){
                    lotterySum = priorityMaximum;
                    break;
                } else{
                    lotterySum += ts.getEffectivePriority();
                }
            }
            if (lotterySum <= 0){
                return null;
            }
            int ticket = Lib.random(lotterySum) + 1;
            
            it = waitQueue.iterator();
            while (it.hasNext() && ticket > 0) {
                ts = (ThreadState)it.next();
                ticket -= ts.getEffectivePriority();
            }
            return ts;
        }
        

        public void print() {
            Lib.assertTrue(Machine.interrupt().disabled());
            
            Iterator it = waitQueue.iterator();
            
            if (it.hasNext())
                System.out.println("===========");
            while (it.hasNext()) {
                ThreadState ts = (ThreadState)it.next();
                ts.print();
            }
        }
        
        protected void removeFromWaitQueue(ThreadState ts) {
            Lib.assertTrue(waitQueue.remove(ts));
        }
        
        protected void addToWaitQueue(ThreadState ts) {
            waitQueue.add(ts);
            donatePriority();
        }
        
        protected void donatePriority() {
            int newDonation = 0;
            
            if (transferPriority){
                for (int i = 0; i < waitQueue.size(); i++) {
                    ThreadState ts = waitQueue.get(i);
                    newDonation += ts.getEffectivePriority();
                }
            }
            
            if (newDonation == donation)
                return;

            donation = newDonation;
            if (this.resAccessing != null) {
                this.resAccessing.resources.put(this , donation);
                this.resAccessing.updatePriority();
            }
        }
        
        /**
         * <tt>true</tt> if this queue should transfer priority from waiting
         * threads to the owning thread.
         */
        public boolean transferPriority;
        
        /** the thread accessing to resources */
        protected ThreadState resAccessing;
        
        protected LinkedList<ThreadState> waitQueue = new LinkedList<ThreadState>();
        
        protected int counter;
        
        protected int donation;
    }

    /**
     * The scheduling state of a thread. This should include the thread's
     * priority, its effective priority, any objects it owns, and the queue
     * it's waiting for, if any.
     *
     * @see nachos.threads.KThread#schedulingState
     */
    protected class ThreadState {
        public ThreadState() {
        }
        /**
         * Allocate a new <tt>ThreadState</tt> object and associate it with the
         * specified thread.
         *
         * @param   thread  the thread this state belongs to.
         */
        public ThreadState(KThread thread) {
            this.thread = thread;
            resourceWaitQueue = null;
            setPriority(priorityDefault);
        }
        
        /**
         * Return the priority of the associated thread.
         *
         * @return  the priority of the associated thread.
         */
        public int getPriority() {
            return originalPriority;
        }

        /**
         * Return the effective priority of the associated thread.
         *
         * @return  the effective priority of the associated thread.
         */
        public int getEffectivePriority() {
            Lib.assertTrue(priority >= originalPriority);
            return priority;
        }
  
        /**
         * Set the priority of the associated thread to the specified value.
         *
         * @param   priority    the new priority.
         */
        public void setPriority(int priority) {
            if (this.originalPriority == priority)
                return;
            this.originalPriority = priority;
                
            updatePriority();
        }

        /**
         * Called when <tt>waitForAccess(thread)</tt> (where <tt>thread</tt> is
         * the associated thread) is invoked on the specified priority queue.
         * The associated thread is therefore waiting for access to the
         * resource guarded by <tt>waitQueue</tt>. This method is only called
         * if the associated thread cannot immediately obtain access.
         *
         * @param   waitQueue   the queue that the associated thread is
         *              now waiting on.
         *
         * @see nachos.threads.ThreadQueue#waitForAccess
         */
        public void waitForAccess(LotteryQueue waitQueue , int time) {
            /** add in waitQueue */
            Lib.assertTrue(!waitQueue.waitQueue.contains(this));
            Lib.assertTrue(this.resourceWaitQueue == null);

            this.addTime = time;
            this.resourceWaitQueue = waitQueue;
            
            waitQueue.addToWaitQueue(this);
        }
        
        /**
         * Called when the associated thread has acquired access to whatever is
         * guarded by <tt>waitQueue</tt>. This can occur either as a result of
         * <tt>acquire(thread)</tt> being invoked on <tt>waitQueue</tt> (where
         * <tt>thread</tt> is the associated thread), or as a result of
         * <tt>nextThread()</tt> being invoked on <tt>waitQueue</tt>.
         *
         * @see nachos.threads.ThreadQueue#acquire
         * @see nachos.threads.ThreadQueue#nextThread
         */
        public void acquire(LotteryQueue waitQueue) {  
            resources.put(waitQueue , waitQueue.donation);
            this.updatePriority();
        }
        
        public void release(LotteryQueue waitQueue) {
            resources.remove(waitQueue);
            this.updatePriority();
        }
        
        public void ready() {
            Lib.assertTrue(this.resourceWaitQueue != null);
            this.resourceWaitQueue = null;
        }
        
        public void print() {
            System.out.println(thread.getName() + " has priority " + priority);
        }
        
        protected void removeFromResources(LotteryQueue waitQueue) {
            Lib.assertTrue(resources.remove(waitQueue) != null);
        }
        
        protected void addToResources(LotteryQueue waitQueue) {
            resources.put(waitQueue , waitQueue.donation);
            updatePriority();
        }
        
        protected void updatePriority() {
            int newEffectivePriority = originalPriority;
            if (!resources.isEmpty()) {
                for (Enumeration<LotteryQueue> queues = resources.keys(); 
                        queues.hasMoreElements(); ) {
                    LotteryQueue q = queues.nextElement();
                    newEffectivePriority += q.donation;
                }
            }
            if (newEffectivePriority == priority)
                return;
            
            priority = newEffectivePriority;
            if (resourceWaitQueue != null)
                resourceWaitQueue.donatePriority();
        }
        
        /** The thread with which this object is associated. */
        protected KThread thread;
        /** The priority of the associated thread. */
        protected int priority;
        protected int originalPriority;
        protected int addTime;
        /** All resources holding by the associated thread */
        protected Hashtable<LotteryQueue , Integer> resources = new Hashtable<LotteryQueue , Integer>();
        /** Waiting caching queue: all other threads waiting with the associated thread */
        protected LotteryQueue resourceWaitQueue;
    }
}
