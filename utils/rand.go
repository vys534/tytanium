package utils

import (
	"encoding/hex"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

// RandBytes gets a set of random bytes. Useful for random strings
// note: i have no idea where i found this code lol but whatever
func RandBytes(n int, c chan<- string, onExit func()) {
	defer onExit()
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	c <- *(*string)(unsafe.Pointer(&b))
}

func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func RandString(len int) string {
	var wg sync.WaitGroup
	randomStringChan := make(chan string, 1)
	go func() {
		wg.Add(1)
		RandBytes(len, randomStringChan, func() { wg.Done() })
	}()
	wg.Wait()
	o := <-randomStringChan
	return o
}
