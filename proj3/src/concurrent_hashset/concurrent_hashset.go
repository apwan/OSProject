package mymap

import (
    "sync"
    "time"
)

var empty string

type MyMap struct {
    p, q, r, lim int
    // p: first level moduli, q: second level moduli,
    // r: hash moduli, lim: memory pool
    e []entry
}

type entry struct {
    tot counter
    q, lim int
    k, v []string
    h, l []int
}

type counter struct {
    items int
    sync.RWMutex
}

func (m *MyMap) hash(s string) int {
    h := 0
    for i := range s {
        h = (h * 2 + int(s[i])) % m.r
    }
    return h
}

func _new(_p, _q, _r, _lim int) (m *MyMap) {
    var t MyMap
    m = &t
    m.p, m.q, m.r, m.lim = _p, _q, _r, _lim
    m.e = make([]entry, m.p)
    empty = string(m.hash(string(int(time.Now().Unix()) % m.p)))
    for i := range m.e {
        build(&m.e[i], m.q, m.lim)
    }
    return
}

func New() *MyMap {
    return _new(1103, 10007, 1000000007, 10000)
}

func build(e *entry, _q, _lim int) {
    e.q, e.lim = _q, _lim
    e.k = make([]string, e.lim)
    e.v = make([]string, e.lim)
    e.h = make([]int, e.q)
    e.l = make([]int, e.lim)
}

func (m *MyMap) Insert(s, t string) int {
    hv := m.hash(s)
    pv, qv := hv % m.p, hv % m.q
    e := &m.e[pv]
    e.tot.RLock()
    var f int
    for i := e.h[qv]; i != 0; i = e.l[i] {
        if e.k[i] == s {
            e.tot.RUnlock()
            return -1
        } else {
            if e.k[i] == empty {
                f = i
            }
        }
    }
    e.tot.RUnlock()
    e.tot.Lock()
    if f != 0 {
        e.k[f], e.v[f] = s, t
    } else {
        e.tot.items++
        tt := e.tot.items
        e.k[tt], e.v[tt] = s, t
        e.l[tt] = e.h[qv]
        e.h[qv] = tt
    }
    e.tot.Unlock()
    return 0
}

func (m *MyMap) Update(s, t string) (r string, ok int) {
    hv := m.hash(s)
    pv, qv := hv % m.p, hv % m.q
    e := &m.e[pv]
    e.tot.RLock()
    for i := e.h[qv]; i != 0; i = e.l[i] {
        if e.k[i] == s {
        ok = 0
            e.tot.RUnlock()
            e.tot.Lock()
            r = e.v[i]
            e.v[i] = t
            e.tot.Unlock()
            return
        }
    }
    r, ok = "", -1
    e.tot.RUnlock()
    return
}

func (m *MyMap) Remove(s string) (r string, ok int) {
    hv := m.hash(s)
    pv, qv := hv % m.p, hv % m.q
    e := &m.e[pv]
    e.tot.RLock()
    for i := e.h[qv]; i != 0; i = e.l[i] {
        if e.k[i] == s {
            e.tot.RUnlock()
            e.tot.Lock()
            e.k[i] = empty
            r, ok = e.v[i], 0
            e.tot.Unlock()
            return
        }
    }
    e.tot.RUnlock()
    s = ""
    ok = -1
    return
}

func (m *MyMap) Get(s string) (r string, ok int) {
    hv := m.hash(s)
    pv, qv := hv % m.p, hv % m.q
    e := &m.e[pv]
    e.tot.RLock()
    for i := e.h[qv]; i != 0; i = e.l[i] {
        if e.k[i] == s {
            r, ok = e.v[i], 0
            e.tot.RUnlock()
            return
        }
    }
    r, ok = "", -1
    e.tot.RUnlock()
    return
}

func (m *MyMap) Count() int {
    cnt := 0
    for i := 0; i < m.p; i++ {
        e := &m.e[i]
        e.tot.RLock()
        for j := 0; j < m.q; j++ {
            for k := e.h[j]; k != 0; k = e.l[k] {
                if e.k[k] != empty {
                    cnt++
                }
            }
        }
        e.tot.RUnlock()
    }
    return cnt
}

// Marshall: return a pointer to the built-in map version of the entire table
func (m *MyMap) Marshall() *map[string]string {
    r := make(map[string]string)
    for i := 0; i < m.p; i++ {
        e := &m.e[i]
        e.tot.RLock()
        for j := 0; j < m.q; j++ {
            for k := e.h[j]; k != 0; k = e.l[k] {
                if e.k[k] != empty {
                    r[e.k[k]] = e.v[k]
                }
            }
        }
        e.tot.RUnlock()
    }
    return &r
}

func (m *MyMap) Clear() {
    for i := 0; i < m.p; i++ {
        e := &m.e[i]
        e.tot.Lock()
        for j := 0; j < m.q; j++ {
            for k := e.h[j]; k != 0; k = e.l[k] {
                e.k[k] = empty
            }
        }
        e.tot.Unlock()
    }
}
