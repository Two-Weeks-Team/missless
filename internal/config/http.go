package config

import (
	"net/http"
	"time"
)

// SharedHTTPClient is a global HTTP client singleton.
// Never create per-request http.Client instances.
var SharedHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     90 * time.Second,
	},
}
