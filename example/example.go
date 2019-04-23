// Create by Yale 2019/4/23 17:37
package example

import (
	"fmt"
	"github.com/yale8848/ratelimiter"
	"golang.org/x/time/rate"
	"time"
)

func single() {
	limiter := ratelimiter.NewLimiter(ratelimiter.NewTokenLimiter("to fast in second", rate.NewLimiter(5, 1)),
		ratelimiter.NewCountLimiter(100, "@every 1m", "to fast in minute"))
	rateLimiter := ratelimiter.NewRateLimiter(limiter)
	count := 0
	for {
		r, msg := rateLimiter.Get("ip").Allow()
		if r {
			count++
		}
		fmt.Printf("allowed:%t; allwFailMsg:%s; allowedCount:%d\r\n", r, msg, count)
		time.Sleep(100 * time.Millisecond)
	}
}

func globe() {
	limiter := ratelimiter.NewLimiter(ratelimiter.NewTokenLimiter("to fast in second", rate.NewLimiter(5, 1)),
		ratelimiter.NewCountLimiter(100, "@every 1m", "to fast in minute"),
		ratelimiter.NewCountLimiter(3000, "@hourly", "to fast in hour"),
		ratelimiter.NewCountLimiter(8000, "@daily", "to fast in day"))
	rateLimiter := ratelimiter.NewRateLimiter(limiter)
	ratelimiter.AddRateLimiter("method1", rateLimiter)
	count := 0
	for {
		r, msg := ratelimiter.GetRateLimiter("method1").Get("ip").Allow()
		if r {
			count++
		}
		fmt.Printf("allowed:%t; allwFailMsg:%s; allowedCount:%d\r\n", r, msg, count)
		time.Sleep(100 * time.Millisecond)
	}
}
