package nachos.threads;

import nachos.machine.*;

/**
 * A <i>communicator</i> allows threads to synchronously exchange 32-bit
 * messages. Multiple threads can be waiting to <i>speak</i>,
 * and multiple threads can be waiting to <i>listen</i>. But there should never
 * be a time when both a speaker and a listener are waiting, because the two
 * threads can be paired off at this point.
 */
public class Communicator {
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
    	++senderCount;
    	while(wordReady || receiverCount == 0){
    		receiveCV.wakeAll();
    		sendCV.sleep();
    	}
    	
    	wordBuffer = word;
    	wordReady = true;
    	--senderCount;
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
    	
    	lock.release();
    	--receiverCount;
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
