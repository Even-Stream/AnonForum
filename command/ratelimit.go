package main

import (
    "sync"

    "golang.org/x/time/rate"
)

// IPRateLimiter .
type IPRateLimiter struct {
    ips         map[string][]*rate.Limiter
    mu          *sync.RWMutex
    retR        rate.Limit
    retB        int
    postR       rate.Limit
    postB       int
    themeR      rate.Limit
    themeB      int
    admR        rate.Limit
    admB        int
    loginR      rate.Limit
    loginB      int
    vidR        rate.Limit
    vidB        int
	userR       rate.Limit
	userB       int
}

// NewIPRateLimiter .
func NewIPRateLimiter(r []rate.Limit, b []int) *IPRateLimiter {
    i := &IPRateLimiter{
        ips:        make(map[string][]*rate.Limiter),
        mu:         &sync.RWMutex{},
        retR:       r[0],
        retB:       b[0],
        postR:      r[1],
        postB:      b[1],
        themeR:     r[2],
        themeB:     b[2],
        admR:       r[3],
        admB:       b[3],
        loginR:     r[4],
        loginB:     b[4],
        vidR:       r[5],
        vidB:       b[5],
		userR:      r[6],
		userB:      b[6],
    }

    return i
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *IPRateLimiter) AddIP(ip string, sel int) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()

    retLimiter := rate.NewLimiter(i.retR, i.retB)
    postLimiter := rate.NewLimiter(i.postR, i.postB)
    themeLimiter := rate.NewLimiter(i.themeR, i.themeB)
    admLimiter := rate.NewLimiter(i.admR, i.admB)
    loginLimiter := rate.NewLimiter(i.loginR, i.loginB)
    vidLimiter := rate.NewLimiter(i.vidR, i.vidB)
	userLimiter := rate.NewLimiter(i.userR, i.userB)
    limiters := []*rate.Limiter{retLimiter, postLimiter, themeLimiter, admLimiter, loginLimiter, vidLimiter, userLimiter}

    i.ips[ip] = limiters

    return limiters[sel]
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise calls AddIP to add IP address to the map
func (i *IPRateLimiter) GetLimiter(ip string, sel int) *rate.Limiter {
    
    i.mu.Lock()
    limiters, exists := i.ips[ip]

    if !exists {
        
        i.mu.Unlock()

        return i.AddIP(ip, sel)
    }


    i.mu.Unlock()


    return limiters[sel]
}
