# Token Session

Session based on Token in http header

## Basic idea

1. After authentication, server should generate a unique token `$token`, and deliver to the client

2. Client should make requests with http header `X-USER-TOKEN: $token`

3. Server will load from specified `Store` identified by `$token`

## Usage


```
import (
    "fmt"
    "github.com/felix021/tokensession"
    "github.com/felix021/tokensession/redis"
)

func main() {
    store, err := redisstore.NewRedisStore(10, "tcp", "127.0.0.1:6379", "<PASSWORD>")
    if err != nil {
        panic(err)
    }

    token := "XXXXXXXX"
    session := tokensession.NewTokenSession(token, store)
    err = session.Load()
    if err != nil {
        panic(err)
    }

    fmt.Println(session.MustGetString("foo", "bar"))

    session.Set("foo", "baz")
    fmt.Println(session.Values)

    fmt.Println(session.MustGetInt("foo", -1))

    session.Save()
}
```
