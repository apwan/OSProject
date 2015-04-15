package nachos.threads;

import nachos.machine.*;

import java.util.TreeSet;
import java.util.HashSet;
import java.util.Iterator;
import java.util.LinkedList;

/**
 * A scheduler that chooses threads based on their priorities.
 *
 * <p>
 * A priority scheduler associates a priority with each thread. The next thread
 * to be dequeued is always a thread with priority no less than any other
 * waiting thread's priority. Like a round-robin scheduler, the thread that is
 * dequeued is, among all the threads of the same (highest) priority, the
 * thread that has been waiting longest.
 *
 * <p>
 * Essentially, a priority scheduler gives access in a round-robin fassion to
 * all the highest-priority threads, and ignores all other threads. This has
 * the potential to
 * starve a thread if there's always a thread waiting with higher priority.
 *
 * <p>
 * A priority scheduler must partially solve the priority inversion problem; in
 * particular, priority must be donated through locks, and through joins.
 */
public class PriorityScheduler extends Scheduler {
    /**
     * Allocate a new priority scheduler.
     */
    public PriorityScheduler() {
    }
    
    /**
     * Allocate a new priority thread queue.
     *
     * @param   transferPriority        <tt>true</tt> if this queue should
     *                                  transfer priority from waiting threads
     *                                  to the owning thread.
     * @return  a new priority thread queue.
     */
    public ThreadQueue newThreadQueue(boolean transferPriority) {
        return new PriorityQueue(transferPriority);
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
     * The default priority for a new thread. Do not change this value.
     */
    public static final int priorityDefault = 1;
    /**
     * The minimum priority that a thread can have. Do not change this value.
     */
    public static final int priorityMinimum = 0;
    /**
     * The maximum priority that a thread can have. Do not change this value.
     */
    public static final int priorityMaximum = 7;    

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
     * A <tt>ThreadQueue</tt> that sorts threads by priority.
     */
    protected class PriorityQueue extends ThreadQueue {
        PriorityQueue(boolean transferPriority) {
            this.transferPriority = transferPriority;
        }

        public void waitForAccess(KThread thread) {
            Lib.assertTrue(Machine.interrupt().disabled());
            getThreadState(thread).waitForAccess(this);
        }

        public void acquire(KThread thread) {
            Lib.assertTrue(Machine.interrupt().disabled());
            ThreadState state = getThreadState(thread);
            // change ownership
            if(this.transferPriority && this.owner != null){
            	
            	this.owner.resQueues.remove(this);
            }
            this.owner = state;
            state.acquire(this);
        }

        public KThread nextThread() {
            Lib.assertTrue(Machine.interrupt().disabled());
            // implement me
            if(waitQueue.isEmpty()) return null;
            // change ownership
            if(this.transferPriority && this.owner != null){
            	this.owner.resQueues.remove(this);
            }
            
            KThread next = pickNextThread();
            if(next != null){
            	Lib.debug('t', "idle??");
            	waitQueue.remove(next);
            	getThreadState(next).acquire(this);
            }
            return next;
        }

        /**
         * Return the next thread that <tt>nextThread()</tt> would return,
         * without modifying the state of this queue.
         *
         * @return      the next thread that <tt>nextThread()</tt> would
         *              return.
         */
        protected KThread pickNextThread() {
            // implement me
        	Lib.debug('t', "pick next: transferPriority: "+ this.transferPriority);
        	Iterator<KThread> it=waitQueue.iterator();
        	KThread next = it.next();
        	while(it.hasNext()){
        		
        		KThread thread = it.next();
        		if(getThreadState(next).getEffectivePriority()>getThreadState(thread).getEffectivePriority()){
        			next = thread;
        		}
        	}
        	
        	
            return next;
        }
        
        public int getEffectivePriority() {

            
            // return minimum priority
            if (transferPriority == false) {
            return priorityMinimum;
            }
            // return max in the waitQueue
            if (dirty) {
                effective = priorityMinimum; 
                for (Iterator<KThread> it = waitQueue.iterator(); it.hasNext();) {  
                    int tmp = getThreadState(it.next()).getEffectivePriority();
                    if ( tmp > effective) { 
                        effective = tmp;
                    }
                }
                dirty = false;
            }

            return effective;
        }
        
        public void setDirty() {
            if (!this.transferPriority) {
                return;
            }
            dirty = true;

            if (this.owner != null) {
                this.owner.setDirty();
            }
        }
        
        public void print() {
            Lib.assertTrue(Machine.interrupt().disabled());
            // implement me (if you want)
        }

        /**
         * <tt>true</tt> if this queue should transfer priority from waiting
         * threads to the owning thread.
         */
        public boolean transferPriority;
        
        private ThreadState owner = null;
        private LinkedList<KThread> waitQueue = new LinkedList<KThread>();
        private boolean dirty = false;
        private int effective;
    }

    /**
     * The scheduling state of a thread. This should include the thread's
     * priority, its effective priority, any objects it owns, and the queue
     * it's waiting for, if any.
     *
     * @see     nachos.threads.KThread#schedulingState
     */
    protected class ThreadState {
        /**
         * Allocate a new <tt>ThreadState</tt> object and associate it with the
         * specified thread.
         *
         * @param       thread  the thread this state belongs to.
         */
        public ThreadState(KThread thread) {
            this.thread = thread;
            
            setPriority(priorityDefault);
        }

        /**
         * Return the priority of the associated thread.
         *
         * @return      the priority of the associated thread.
         */
        public int getPriority() {
            return priority;
        }

        /**
         * Return the effective priority of the associated thread.
         *
         * @return      the effective priority of the associated thread.
         */
        public int getEffectivePriority() {
            // implement me
        	int ret = this.priority;
        	if(dirty){
        		
        	}
            return ret;
            
        }

        /**
         * Set the priority of the associated thread to the specified value.
         *
         * @param       priority        the new priority.
         */
        public void setPriority(int priority) {
            if (this.priority == priority)
                return;
            
            this.priority = priority;
            
            // implement me
            setDirty();
        }
        
        public void setDirty() {
            if (dirty) {
                return;
            }
            dirty = true;

            PriorityQueue tmp = (PriorityQueue)waiting;
            if (tmp != null) {
                tmp.setDirty();
            }

        }


        /**
         * Called when <tt>waitForAccess(thread)</tt> (where <tt>thread</tt> is
         * the associated thread) is invoked on the specified priority queue.
         * The associated thread is therefore waiting for access to the
         * resource guarded by <tt>waitQueue</tt>. This method is only called
         * if the associated thread cannot immediately obtain access.
         *
         * @param       waitQueue       the queue that the associated thread is
         *                              now waiting on.
         *
         * @see nachos.threads.ThreadQueue#waitForAccess
         */
        public void waitForAccess(PriorityQueue waitQueue) {
            // implement me
        	waitQueue.waitQueue.add(this.thread);
        	waitQueue.setDirty();
        	waiting = waitQueue;
        	
        	// remove if already in resQueues
        	if(resQueues.indexOf(waitQueue)>-1){
        		resQueues.remove(waitQueue);
        		waitQueue.owner = null;
        	}
        	
        	
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
        public void acquire(PriorityQueue waitQueue) {
            // implement me
        	
        	resQueues.add(waitQueue);
        	// clear if just waiting
        	if(waitQueue == waiting){
        		waiting = null;
        	}
        	setDirty();
        }       

        /** The thread with which this object is associated. */    
        protected KThread thread;
        /** The priority of the associated thread. */
        protected int priority;
        protected int effective;
        private boolean dirty = false;
        protected ThreadQueue waiting;
        // queues wait for this resource
        protected LinkedList<ThreadQueue> resQueues = new LinkedList<ThreadQueue>();
        
        
    }
}
