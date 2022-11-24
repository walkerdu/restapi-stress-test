package internal

import (
	"html/template"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var funcsMap = template.FuncMap{
	"rand":     rand_,
	"uuid":     uuid_,
	"date":     date,
	"randDate": randDate,
	"now":      now,
}

// return [min, max)
func rand_(min, max int64) int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(max-min) + min
}

func uuid_() string {
	return uuid.New().String()
}

func date(fmt string) string {
	return time.Now().Format(fmt)
}

func randDate(fmt string) string {
	return time.Unix(rand.Int63n(time.Now().Unix()-94608000)+94608000, 0).Format(fmt)
}

func now() int64 {
	return time.Now().Unix()
}
