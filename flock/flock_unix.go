// Copyright 2018 GRAIL, Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Package flock implements a simple POSIX file-based advisory lock.
package flock

import (
	"sync"
	"syscall"

	"github.com/grailbio/base/log"
)

type T struct {
	name string
	fd   int
	mu   sync.Mutex
}

// New creates an object that locks the given path.
func New(path string) *T {
	return &T{name: path}
}

// Lock locks the file. Iff Lock() returns nil, the caller must call Unlock()
// later.
func (f *T) Lock() error {
	f.mu.Lock() // Serialize the lock within one process.

	var err error
	f.fd, err = syscall.Open(f.name, syscall.O_CREAT|syscall.O_RDWR, 0777)
	if err != nil {
		f.mu.Unlock()
		return err
	}
	err = syscall.Flock(f.fd, syscall.LOCK_EX|syscall.LOCK_NB)
	for err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
		log.Printf("waiting for lock %s", f.name)
		err = syscall.Flock(f.fd, syscall.LOCK_EX)
	}
	if err != nil {
		f.mu.Unlock()
	}
	return err
}

// Lock unlocks the file.
func (f *T) Unlock() error {
	err := syscall.Flock(f.fd, syscall.LOCK_UN)
	if err := syscall.Close(f.fd); err != nil {
		log.Error.Printf("close %s: %v", f.name, err)
	}
	f.mu.Unlock()
	return err
}
