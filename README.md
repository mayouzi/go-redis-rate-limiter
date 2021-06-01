# rate limiter with Redis

## Install

```shell
go get github.com/mayouzi/go-redis-rate-limiter
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	rate "github.com/mayouzi/redis_rate_limiter"
	"log"
)

type Resource struct {
	prefix string
}

func (r *Resource) Key(idf string) string {
	return fmt.Sprintf("%s%s", r.prefix, idf)
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       2,
	})

	key := "/user/profile"

	resource := &Resource{
		prefix: "_api_qps_",
	}
	limiter, err := rate.NewRateLimiter(client, 20, 1, resource)
	if err != nil {
		panic(err)
	}

	result, err := limiter.Usage(key)
	if err != nil {
		panic(err)
	}
	if result.Pass {
		log.Printf("pass, already number: %d", result.Usage)
    } else {
		log.Printf("need wailt: %v", result.Wait)
    }
}
```