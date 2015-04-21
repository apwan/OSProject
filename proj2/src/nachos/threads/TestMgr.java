package nachos.threads;

import java.util.*;

public class TestMgr {
    
    public static List<String > testNames=new ArrayList<String>();
    public static List<Integer > testStatus=new ArrayList<Integer>();
    
    public static final int init=0,pass=1,fail=2;
    
    public static int addTest(String s)
    {
        testNames.add(s);
        int id=testNames.size()-1;
        testStatus.add(init);
        return id;
    }
    public static void finishTest(int id)
    {
        finishTest(id,true);
    }
    public static void finishTest(int id, boolean passing)
    {
        testStatus.set(id, passing? pass:fail);
    }
    
    public static void printAll()
    {
        int passed=0,failed=0,blocked=0,error=0;
        for(int i=0;i<testStatus.size();i++)
        {
            switch(testStatus.get(i))
            {
                case pass: passed++;break;
                case fail: failed++;break;
                case init: blocked++;break;
                default: error++;break;
            }
        }
            
        System.out.println("Total test run:"+testStatus.size() );
        System.out.println("Passed:"+passed );
        System.out.println("Failed:"+failed);
        System.out.println("blocked:"+blocked );

        System.out.println("===" );
        for(int i=0;i<testStatus.size();i++)
        {
            switch(testStatus.get(i))
            {
                case pass: continue;
                case fail: System.out.print("fail:");break;
                case init: System.out.print("unfinished case:");break;
                default: System.out.print("???:");
            }
            System.out.println(testNames.get(i));
        }
        System.out.println("===" );
    }
}
