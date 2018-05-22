# Simple logrus wrapper
[![CircleCI](https://circleci.com/gh/codeamp/logger.svg?style=svg)](https://circleci.com/gh/codeamp/logger)


### Example Usage
```
package main

import (
	log "github.com/codeamp/logger"
)

func main() {
  log.InfoWithFields("hello", log.Fields{
    "world": "earth",
  })

  log.Println("Hello World")
  
  log.Debug("Hello World")
}
```

[GoDoc](https://godoc.org/github.com/codeamp/logger)
