//u`ofe`{i`nlhofgdhbi`nsthvtmhg`o
//PART OF THE NACHOS. DON'T CHANGE CODE OF THIS LINE
package nachos.threads;

import nachos.machine.*;

import java.util.Enumeration;
import java.util.Hashtable;
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
     * @param	transferPriority	<tt>true</tt> if this queue should
     *					transfer priority from waiting threads
     *					to the owning thread.
     * @return	a new priority thread queue.
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
     * @param	thread	the thread whose scheduling state to return.
     * @return	the scheduling state of the specified thread.
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
    	PriorityQueue() {
    	}
    	
    	PriorityQueue(boolean transferPriority) {
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
    	 * @return	the next thread that <tt>nextThread()</tt> would
    	 *		return.
    	 */
    	protected ThreadState pickNextThread() {
    		ThreadState nextThreadState = null;
    		int maxPriority = -1;
    		int addTime = Integer.MAX_VALUE;
    		
    		for (int i = 0; i < waitQueue.size(); i++) {
    			ThreadState ts = waitQueue.get(i);
    			int tmpPriority = ts.getEffectivePriority();
    			if (tmpPriority > maxPriority) {
    				maxPriority = tmpPriority;
    				addTime = ts.addTime;
    				nextThreadState = ts;
    			}
    			else if (tmpPriority == maxPriority && addTime > ts.addTime) {
    				addTime = ts.addTime;
    				nextThreadState = ts;
    			}
    		}
    		return nextThreadState;
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

    	protected int getMaxPriority() {
    		int maxPriority = priorityMinimum;
    		if (transferPriority) {
    			for (int i = 0; i < waitQueue.size(); i++) {
    				ThreadState ts = waitQueue.get(i);
    				maxPriority = Math.max(maxPriority , ts.getEffectivePriority());
    			}
    		}
    		return maxPriority;
    	}
    	
    	protected void removeFromWaitQueue(ThreadState ts) {
    		Lib.assertTrue(waitQueue.remove(ts));
    	}
    	
    	protected void addToWaitQueue(ThreadState ts) {
    		waitQueue.add(ts);
    		donatePriority();
    	}
    	
    	protected void donatePriority() {
    		int newDonation = priorityMinimum;
    		
    		if (transferPriority)
    			newDonation = Math.max(newDonation , this.getMaxPriority());
    		
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
   	 * @see	nachos.threads.KThread#schedulingState
   	 */
   	protected class ThreadState {
   		public ThreadState() {
   		}
   		/**
   		 * Allocate a new <tt>ThreadState</tt> object and associate it with the
   		 * specified thread.
   		 *
   		 * @param	thread	the thread this state belongs to.
   		 */
   		public ThreadState(KThread thread) {
   			this.thread = thread;
   			//resources = new Hashtable<PriorityQueue , Integer>();
   			resourceWaitQueue = null;
   			setPriority(priorityDefault);
   		}
   		
   		/**
    	 * Return the priority of the associated thread.
    	 *
    	 * @return	the priority of the associated thread.
    	 */
    	public int getPriority() {
    		return originalPriority;
    	}

    	/**
    	 * Return the effective priority of the associated thread.
    	 *
    	 * @return	the effective priority of the associated thread.
    	 */
    	public int getEffectivePriority() {
    		Lib.assertTrue(priority >= originalPriority);
    		return priority;
    	}
  
    	/**
    	 * Set the priority of the associated thread to the specified value.
    	 *
    	 * @param	priority	the new priority.
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
    	 * @param	waitQueue	the queue that the associated thread is
    	 *				now waiting on.
    	 *
    	 * @see	nachos.threads.ThreadQueue#waitForAccess
    	 */
    	public void waitForAccess(PriorityQueue waitQueue , int time) {
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
    	 * @see	nachos.threads.ThreadQueue#acquire
    	 * @see	nachos.threads.ThreadQueue#nextThread
    	 */
    	public void acquire(PriorityQueue waitQueue) {	
    		resources.put(waitQueue , waitQueue.donation);
    		this.updatePriority();
    	}
    	
    	public void release(PriorityQueue waitQueue) {
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
    	
    	protected void removeFromResources(PriorityQueue waitQueue) {
    		Lib.assertTrue(resources.remove(waitQueue) != null);
    	}
    	
    	protected void addToResources(PriorityQueue waitQueue) {
    		resources.put(waitQueue , waitQueue.donation);
    		updatePriority();
    	}
    	
    	protected void updatePriority() {
    		int newEffectivePriority = originalPriority;
    		if (!resources.isEmpty()) {
    			for (Enumeration<PriorityQueue> queues = resources.keys(); 
    					queues.hasMoreElements(); ) {
    				PriorityQueue q = queues.nextElement();
    				newEffectivePriority = Math.max(newEffectivePriority , 
    					q.donation);
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
    	protected Hashtable<PriorityQueue , Integer> resources = new Hashtable<PriorityQueue , Integer>();
    	/** Waiting caching queue: all other threads waiting with the associated thread */
    	protected PriorityQueue resourceWaitQueue;
    }
}