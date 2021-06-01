package redis_rate_limiter

import (
	"context"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const (
	SCRIPT = `
		local current
		local limitCount    = tonumber(ARGV[1])
		local expireSecond  = tonumber(ARGV[2])
		local mKey          = KEYS[1]
		
		local incr = function (key, expire)
			local v = redis.call('INCR', key)
			if v == 1 then
				redis.call("EXPIRE", key, expire)
			end
			return v
		end
		
		local yetBeyonds = function (key, maxV, expire)
			local t = tonumber(redis.call("PTTL", key))
			local v
			if t > 0 then
				v = t/1000.0
			else
				v = expire*1.0
			end
				return tostring(v)
		end
		
		current = tonumber(redis.call("GET", mKey))
		if current then
			if current >= limitCount then
				return (yetBeyonds(mKey, limitCount, expireSecond))
			else
				return (incr(mKey, expireSecond))
			end
		else
			return (incr(mKey, expireSecond))
		end
	`
)

// Resource rate resource
type Resource interface {
	Key(string)string
}

// Result usage result
type Result struct {
	Wait time.Duration
	Usage int64
	Pass bool
}

// RateLimiter is a Redis-backed rate limiter.
type RateLimiter struct {
	client *redis.Client
	max int64
	resource Resource
	expire int64
	_scriptSha string
	_script string
}

// Usage usage with idf
func (rl *RateLimiter) Usage(idf string) (*Result, error) {

	key := rl.resource.Key(idf)
	ctx := context.TODO()

	rawResult, err := rl.client.EvalSha(
		ctx, rl._scriptSha, []string{key}, rl.max, rl.expire).Result()
	if err != nil {
		rawResult, err = rl.client.Eval(
			ctx, rl._script, []string{key}, rl.max, rl.expire).Result()
		if err != nil {
			return nil, err
		}
	}

	result := &Result{}

	switch rawResult.(type) {
	case int64:
		result.Pass = true
		result.Usage = rawResult.(int64)
	case string:
		n, er := strconv.ParseFloat(rawResult.(string), 32)
		if er != nil {
			return nil, er
		}
		result.Wait = time.Duration(n)
	}
	return result, nil
}

// NewRateLimiter create RateLimiter
func NewRateLimiter(client *redis.Client, max int64, expire int64, resource Resource) (*RateLimiter, error) {

	script := redis.NewScript(SCRIPT)
	ctx := context.TODO()

	sha, err := script.Load(ctx, client).Result()
	if err != nil {
		return nil, err
	}

	return &RateLimiter{
		client: client,
		max: max,
		expire: expire,
		resource: resource,
		_scriptSha: sha,
		_script: SCRIPT,
	}, nil
}
