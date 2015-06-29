from random import *

# gap: time to sleep after start/stop/Endblock, numOp: number of instructions, name: output name
def gen(gap, numOp, keyRange, valueRange, name):
    with open(name, 'w') as f:
        op = []
        for i in range(numOp):
            r = randint(1, 4)
            # 1: Get, 2: Insert, 3: Update, 4: Remove
            if r == 1:
                op.append('Get %d' % (randint(1, keyRange)))
            elif r == 2:
                op.append('Put %d %d' % (randint(1, keyRange), randint(1, valueRange)))
            elif r == 3:
                op.append('Update %d %d' % (randint(1, keyRange), randint(1, valueRange)))
            elif r == 4:
                op.append('Delete %d' % (randint(1, keyRange)))

        primary = 1
        backup = 1
        inBlock = 0

        '''
        print >> f, 'Exec bin/%s_server -p' % ('stop' if primary else 'start')
        print >> f, 'Switch primary'
        print >> f, 'Sleep %d' % (gap)
        primary ^= 1

        print >> f, 'Exec bin/%s_server -b' % ('stop' if backup else 'start')
        print >> f, 'Switch backup'
        print >> f, 'Sleep %d' % (gap)
        backup ^= 1
        '''

        print >> f, 'Sleep %d' % (gap)
        print >> f, 'Switch %d' % (1)
        
        sd = randint(numOp / 3, numOp / 3 * 2)

        for i in range(numOp):
            '''
            if randint(1, 20) == 0 and not inBlock: # switch primary with probability 5% and not in a block
                print >> f, 'Exec bin/%s_server -p' % ('stop' if primary else 'start')
                print >> f, 'Switch primary'
                print >> f, 'Sleep %d' % (gap)
                primary ^= 1
            if randint(1, 20) == 0 and not inBlock: # switch backup with probability 5% and not in a block
                print >> f, 'Exec bin/%s_server -b' % ('stop' if backup else 'start')
                print >> f, 'Switch backup'
                print >> f, 'Sleep %d' % (gap)
                backup ^= 1
            '''
            if i == sd:
                print >> f, 'stop_server %d' % (randint(1, 3))

            if randint(1, 20) == 1:
                print >> f, 'Switch %d' % (randint(1, 3))
            if randint(1, 20) == 1: # enter/leave a block with probability %5
                print >> f, '%s' % ('Endblock' if inBlock else 'Block')
                if inBlock:
                    print >> f, 'Sleep %d' % (gap)
                inBlock ^= 1
            print >> f, op[i]

        if inBlock:
            print >> f, '%s' % ('Endblock' if inBlock else 'Block')
            print >> f, 'Sleep %d' % (gap)

        print >> f, 'Sleep %d' % (gap)

        '''
        if not primary:
            print >> f, 'Exec bin/%s_server -p' % ('stop' if primary else 'start')
            print >> f, 'Switch primary'
            print >> f, 'Sleep %d' % (gap)
        if not backup:
            print >> f, 'Exec bin/%s_server -b' % ('stop' if backup else 'start')
            print >> f, 'Switch backup'
            print >> f, 'Sleep %d' % (gap)
        '''

# seed(42)
for i in range(2, 10):
    gen(500, 100, 30, 100, str(i) + '.test')
