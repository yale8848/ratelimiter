// Create by Yale 2019/4/23 17:18
package ratelimiter

import (
	"fmt"
	"github.com/robfig/cron"
	"golang.org/x/time/rate"
	"sync"
	"time"
)

var mapRateLimiter = make(map[string]*RateLimiter)

func AddRateLimiter(key string, rateLimiter *RateLimiter) {
	mapRateLimiter[key] = rateLimiter
}
func GetRateLimiter(key string) *RateLimiter {
	return mapRateLimiter[key]
}

type TokenLimiter struct {
	AllowFailMsg string
	tokenBucket  *rate.Limiter
}

func NewTokenLimiter(allowFailMsg string, tokenBucket *rate.Limiter) *TokenLimiter {
	return &TokenLimiter{AllowFailMsg: allowFailMsg, tokenBucket: tokenBucket}
}

type Limiter struct {
	tokenLimiter *TokenLimiter
	countLimiter []*CountLimiter
}

func NewLimiter(tokenBucket *TokenLimiter, countLimiters ...*CountLimiter) *Limiter {
	cl := make([]*CountLimiter, 0)
	for _, v := range countLimiters {
		if v.max > 0 || len(v.cron) > 0 {
			_, err := cron.Parse(v.cron)
			if err == nil {
				cl = append(cl, v)
			}
		}
	}
	return &Limiter{tokenLimiter: tokenBucket, countLimiter: countLimiters}
}
func (limiter *Limiter) AllowTokenLimiter() (bool, string) {
	if limiter.tokenLimiter != nil && limiter.tokenLimiter.tokenBucket != nil {
		if !limiter.tokenLimiter.tokenBucket.Allow() {
			return false, limiter.tokenLimiter.AllowFailMsg
		}
	}
	return true, ""
}
func (limiter *Limiter) AllowCountLimiter() (bool, string) {
	for i, _ := range limiter.countLimiter {
		limiter.countLimiter[i].Increase()
		if !limiter.countLimiter[i].Allow() {
			return false, limiter.countLimiter[i].allowFailMsg
		}
	}
	return true, ""
}
func (limiter *Limiter) Allow() (bool, string) {
	r, s := limiter.AllowTokenLimiter()
	if !r {
		return r, s
	}
	return limiter.AllowCountLimiter()
}

type CountLimiter struct {
	max          uint64
	count        uint64
	cron         string
	lock         sync.Mutex
	allowFailMsg string
}

func NewCountLimiter(max uint64, cron, allowFailMsg string) *CountLimiter {
	return &CountLimiter{max: max, cron: cron, allowFailMsg: allowFailMsg}
}
func (countLimiter *CountLimiter) StartCount() {
	fmt.Println(time.Now())
	c := cron.New()
	_ = c.AddFunc(countLimiter.cron, func() {
		fmt.Println(time.Now())
		countLimiter.lock.Lock()
		defer countLimiter.lock.Unlock()
		countLimiter.count = 0
	})
	c.Start()
}
func (countLimiter *CountLimiter) Increase() {
	countLimiter.lock.Lock()
	defer countLimiter.lock.Unlock()
	countLimiter.count++
}
func (countLimiter *CountLimiter) Allow() bool {
	return countLimiter.count <= countLimiter.max
}

type RateLimiter struct {
	limiter *Limiter
	data    map[string]*Limiter
	lock    sync.Mutex
}

func NewRateLimiter(limiter *Limiter) *RateLimiter {
	return &RateLimiter{
		data:    make(map[string]*Limiter),
		limiter: limiter,
	}
}
func (rl *RateLimiter) copyLimiter() *Limiter {
	lim := &TokenLimiter{}
	if rl.limiter.tokenLimiter != nil && rl.limiter.tokenLimiter.tokenBucket != nil {
		lim.tokenBucket = rate.NewLimiter(rl.limiter.tokenLimiter.tokenBucket.Limit(),
			rl.limiter.tokenLimiter.tokenBucket.Burst())
		lim.AllowFailMsg = rl.limiter.tokenLimiter.AllowFailMsg
	} else {
		lim = nil
	}
	countLimiters := make([]*CountLimiter, 0)

	for _, v := range rl.limiter.countLimiter {
		c := &CountLimiter{max: v.max, allowFailMsg: v.allowFailMsg, cron: v.cron}
		countLimiters = append(countLimiters, c)
		c.StartCount()
	}

	return &Limiter{
		tokenLimiter: lim,
		countLimiter: countLimiters,
	}
}
func (rl *RateLimiter) Get(key string) *Limiter {

	if rl.limiter == nil {
		return nil
	}
	rl.lock.Lock()
	defer rl.lock.Unlock()
	if v, ok := rl.data[key]; ok {
		return v
	}
	lt := rl.copyLimiter()
	rl.data[key] = lt
	return lt
}
