package nachos.threads;
import nachos.ag.BoatGrader;

public class Boat
{
    static BoatGrader bg;

    static int
        adultsOnSource = 0,
        childrenOnSource = 0,
        childrenAboard = 0,
        boat = 0; // 0: on Oahu, 1: on Molokai
    static Lock
        arith = new Lock(),
        boarding = new Lock(),
        childrenUnload = new Lock();
    static Condition
        bd = new Condition(boarding),
        cu = new Condition(childrenUnload);
    
    public static void selfTest()
    {
        BoatGrader b = new BoatGrader();
        
        //System.out.println("\n ***Testing Boats with only 2 children***");
        //begin(0, 2, b);
        
        //System.out.println("\n ***Testing Boats with 2 children, 1 adult***");
        //begin(1, 2, b);
        
        //System.out.println("\n ***Testing Boats with 3 children, 3 adults***");
        //begin(3, 3, b);

        //System.out.println("\n ***Testing Boats with 2 children, 10 adults***");
        //begin(10, 2, b);

        //System.out.println("\n ***Testing Boats with 5 children, 5 adults***");
        //begin(5, 5, b);

        //System.out.println("\n ***Testing Boats with 10 children, 2 adults***");
        //begin(2, 10, b);

        //System.out.println("\n ***Testing Boats with 100 children, 100 adults***");
        //begin(100, 100, b);
          
    }

    public static void begin( int adults, int children, BoatGrader b )
    {
        // Store the externally generated autograder in a class
        // variable to be accessible by children.
        bg = b;

        // Instantiate global variables here
        
        // Create threads here. See section 3.4 of the Nachos for Java
        // Walkthrough linked from the projects page.

        /*
        Runnable r = new Runnable() {
            public void run() {
                SampleItinerary();
            }
        };
        KThread t = new KThread(r);
        t.setName("Sample Boat Thread");
        t.fork();
        */
        for(int i = 0; i < children; i++)
        {
            Runnable r = new Runnable() {
                public void run() {
                    ChildItinerary();
                }
            };
            KThread t = new KThread(r);
            t.setName("Child Thread No. " + i);
            t.fork();
        }
        for(int i = 0; i < adults; i++)
        {
            Runnable r = new Runnable() {
                public void run() {
                    AdultItinerary();
                }
            };
            KThread t = new KThread(r);
            t.setName("Adult Thread No. " + i);
            t.fork();
        }
        KThread.yield();
        while(childrenOnSource + adultsOnSource > 0)
        {
            //System.out.println(childrenOnSource + " " + adultsOnSource);
            KThread.yield();
        }
        return;
    }

    static void AdultItinerary()
    {
        /* This is where you should put your solutions. Make calls
           to the BoatGrader to show that it is synchronized. For
           example:
               bg.AdultRowToMolokai();
           indicates that an adult has rowed the boat across to Molokai
        */
        arith.acquire();
        {
            adultsOnSource++;
        }
        arith.release();
        KThread.yield();
        int state = 0; // 0: on Oahu, 1: on Molokai
        while(state == 0)
        {
            boarding.acquire();
            {
                while(childrenAboard > 0 || childrenOnSource > 1 || boat == 1)
                {
                    bd.sleep();
                }
                arith.acquire();
                {
                    adultsOnSource--;
                }
                arith.release();
                state = 1;
                boat = 1;
                bg.AdultRowToMolokai();
                bd.wakeAll();
            }
            boarding.release();
        }
        return;
    }

    static void ChildItinerary()
    {
        arith.acquire();
        {
            childrenOnSource++;
        }
        arith.release();
        KThread.yield();
        int state = 0; // 0: on Oahu, 1: on Molokai
        while(true)
        {
            int isDriver = 1;
            while(state == 1)
            {
                boarding.acquire();
                {
                    while(childrenAboard > 0 || boat == 0)
                    {
                        bd.sleep();
                    }
                    bg.ChildRowToOahu();
                    arith.acquire();
                    {
                        childrenOnSource++;
                    }
                    arith.release();
                    state = 0;
                    boat = 0;
                    bd.wakeAll();
                }
                boarding.release();
            }
            while(state == 0)
            {
                boarding.acquire();
                {
                    //System.out.println("!!!!! child boarding." + childrenAboard);
                    while(childrenAboard == 2 || boat == 1)
                    {
                        bd.sleep();
                    }
                    //System.out.println("!!!!!!! " + boat);
                    arith.acquire();
                    {
                        childrenAboard++;
                    }
                    arith.release();
                    isDriver = 1;
                    while(isDriver == 1 && childrenAboard == 1 && childrenOnSource + adultsOnSource > 1 || boat == 1)
                    {
                        if(childrenOnSource == 1 || boat == 1)
                        {
                            arith.acquire();
                            {
                                childrenAboard--;
                            }
                            arith.release();
                            bd.sleep();
                            arith.acquire();
                            {
                                childrenAboard++;
                            }
                            arith.release();
                        }
                        else
                        {
                            //System.out.println("!!!!! rider decided.");
                            isDriver = 0;
                            bd.wakeAll();
                            bd.sleep();
                        }
                    }
                    if(childrenOnSource == 1 && isDriver == 1)
                    {
                        --childrenAboard;
                    }
                    bd.wakeAll();
                }
                boarding.release();
                if(isDriver == 1)
                {
                    childrenUnload.acquire();
                    {
                        bg.ChildRowToMolokai();
                        arith.acquire();
                        {
                            childrenOnSource--;
                        }
                        arith.release();
                        state = 1;
                        cu.wakeAll();
                    }
                    childrenUnload.release();
                }
                else
                {
                    childrenUnload.acquire();
                    {
                        /*
                        while(childrenAboard > 1)
                        {
                            cu.sleep();
                        }
                        */
                        bg.ChildRideToMolokai();
                        bg.ChildRowToOahu(); // counters not modified
                        arith.acquire();
                        {
                            childrenAboard -= 2;
                        }
                        arith.release();
                    }
                    childrenUnload.release();
                }
            }
        }
    }

    static void SampleItinerary()
    {
        // Please note that this isn't a valid solution (you can't fit
        // all of them on the boat). Please also note that you may not
        // have a single thread calculate a solution and then just play
        // it back at the autograder -- you will be caught.
        System.out.println("\n ***Everyone piles on the boat and goes to Molokai***");
        bg.AdultRowToMolokai();
        bg.ChildRideToMolokai();
        bg.AdultRideToMolokai();
        bg.ChildRideToMolokai();
    }
    
}
