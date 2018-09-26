package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiters struct {
	noLimitIPs map[string]struct{} // concurrent read safe after init.
	visitors   map[string]*rate.Limiter
	sync.RWMutex
}

func (ls *limiters) tryAddVisitor(ip string) *rate.Limiter {
	ls.Lock()
	defer ls.Unlock()
	limiter, exists := ls.visitors[ip]
	if exists {
		return limiter
	}
	limit := rate.Every(time.Minute / time.Duration(requestsPerMinuteLimit))
	limiter = rate.NewLimiter(limit, requestsPerMinuteLimit/10)
	log.Println("Added new visitor:", ip, " limit ", fmt.Sprint(requestsPerMinuteLimit))
	ls.visitors[ip] = limiter
	return limiter
}

func (ls *limiters) getVisitor(ip string) *rate.Limiter {
	ls.RLock()
	limiter, exists := ls.visitors[ip]
	ls.RUnlock()
	if !exists {
		return ls.tryAddVisitor(ip)
	}
	return limiter
}

func (ls *limiters) AllowLimit(r ModifiedRequest) bool {
	if _, ok := ls.noLimitIPs[r.RemoteAddr]; ok {
		return true
	}
	limiter := ls.getVisitor(r.RemoteAddr)
	if limiter.Allow() == false {
		return false
	}
	return true
}
