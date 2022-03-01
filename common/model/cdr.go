package model

type CdrRedis struct {
	LastPushedAt string `json:"last_pushed_at"`
	FailedCount  int    `json:"failed_count"`
}
