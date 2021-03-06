package utils

import (
	"math/rand"
)

func GetRandomIntRange(len uint64) uint64 {
	return uint64(rand.Intn(int(len)))
}

func GetRandomStrings(len uint64) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rs := ""
	var i uint64
	for i = 0; i < len; i++ {
		r := rand.Intn(62)
		if r == 0 {
			r = 1
		}
		rs += str[r-1 : r]
	}
	return rs
}

// 32bit 随机
func RandInt32(min, max int32) int32 {
	if min >= max {
		return max
	}
	return rand.Int31n(max-min) + min
}
