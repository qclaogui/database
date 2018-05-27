package builder

import (
	"math/rand"
	"runtime"
	"strings"
	"time"
)

func itoa(val int) string { // do it here rather than with fmt to avoid dependency
	if val < 0 {
		return "-" + uitoa(uint(-val))
	}
	return uitoa(uint(val))
}

func uitoa(val uint) string {
	var buf [32]byte // big enough for int64
	i := len(buf) - 1
	for val >= 10 {
		buf[i] = byte(val%10 + '0')
		i--
		val /= 10
	}
	buf[i] = byte(val + '0')
	return string(buf[i:])
}

// NumAllocs num
func NumAllocs(fn func()) uint64 {
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(1))
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	fn()
	runtime.ReadMemStats(&m2)
	return m2.Mallocs - m1.Mallocs
}

const alphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandomString returns the String
func RandomString(strLen int) string {
	var s strings.Builder
	s.Grow(strLen)

	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < strLen; i++ {
		s.WriteByte(alphaNum[rand.Intn(len(alphaNum))])
	}
	return s.String()
}
