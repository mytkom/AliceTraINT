package applog

import (
	"log"
	"time"
)

func ExternalCall(service, operation, details string, fn func() error) error {
	log.Printf("EXTERNAL %s %s started %s", service, operation, details)
	start := time.Now()

	err := fn()
	if err != nil {
		log.Printf("EXTERNAL %s %s failed in %v %s: %v", service, operation, time.Since(start), details, err)
		return err
	}

	log.Printf("EXTERNAL %s %s completed in %v %s", service, operation, time.Since(start), details)
	return nil
}

func ExternalCallResult[T any](service, operation, details string, fn func() (T, error)) (T, error) {
	log.Printf("EXTERNAL %s %s started %s", service, operation, details)
	start := time.Now()

	result, err := fn()
	if err != nil {
		log.Printf("EXTERNAL %s %s failed in %v %s: %v", service, operation, time.Since(start), details, err)
		return result, err
	}

	log.Printf("EXTERNAL %s %s completed in %v %s", service, operation, time.Since(start), details)
	return result, nil
}

func ExternalCacheHit(service, operation, details string) {
	log.Printf("EXTERNAL %s %s cache hit %s", service, operation, details)
}

func Fallback(primaryOperation string, err error, fallbackOperation string) {
	log.Printf("FALLBACK %s failed (%v), using %s", primaryOperation, err, fallbackOperation)
}
