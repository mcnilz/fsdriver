package main

import (
	"log"
)

// logger is a minimal facade to allow swapping later (e.g., zap/logrus).
// For now, it delegates to the standard library log package.

type logger struct{}

func (l logger) Info(msg string, kv ...any) {
	log.Println(append([]any{"level", "INFO", "msg", msg}, kv...)...)
}

func (l logger) Error(msg string, kv ...any) {
	log.Println(append([]any{"level", "ERROR", "msg", msg}, kv...)...)
}

var logx = logger{}
