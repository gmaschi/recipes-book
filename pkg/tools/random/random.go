package random

import (
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// String returns a random string of n characters
func String(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

// Email returns a random email
func Email() string {
	return String(10) + "@example.com"
}

// StringSlice creates a slice of length n containing random strings
func StringSlice(n int) []string {
	ss := make([]string, 0, n)
	for i := 0; i < n; i++ {
		ss = append(ss, String(10))
	}
	return ss
}