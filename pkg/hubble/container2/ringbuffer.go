// Copyright 2020 Authors of Hubble
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package container2

import (
	"sync"
	"time"

	v1 "github.com/cilium/cilium/pkg/hubble/api/v1"
)

// A RingBuffer buffers a fixed number of recent events. Writing events never
// blocks, but reads will drop events if the reader cannot keep up.
//
// Written events are given sequence numbers in the order in which they are
// received.
type RingBuffer struct {
	// rwMutex synchronizes access. It is included as a field rather than
	// embedded so that its methods are not exported.
	rwMutex sync.RWMutex

	// last is the sequence number of the last event in the buffer.
	last int

	// next is the sequence number of the next event.
	next int

	buffer  []*v1.Event
	readers map[chan<- *v1.Event]*ReaderStats

	// sent is the total number of events sent to followers.
	sent int

	// dropped is the total number of events dropped by followers.
	dropped int

	// rUnlockLockFunc is called between releasing a read lock and acquiring a
	// write lock. It is for testing purposes only.
	rUnlockLockFunc func()
}

// A ReaderCancelFunc cancels a reader and returns its statisitcs.
type ReaderCancelFunc func() ReaderStats

// A RingBufferOption sets an option on a RingBuffer.
type RingBufferOption func(*RingBuffer)

// A ReaderStats collects statistics on a reader.
type ReaderStats struct {
	Sent    int
	Dropped int
}

// A RingBufferStatus contains a snapshot of a ring buffer's state.
type RingBufferStatus struct {
	NumEvents       int
	SeenEvents      int
	OldestEventTime time.Time
	NewestEventTime time.Time
}

// WithCapacity sets the capacity.
func WithCapacity(capacity int) RingBufferOption {
	return func(b *RingBuffer) {
		b.buffer = make([]*v1.Event, capacity)
	}
}

// NewRingBuffer returns a new RingBuffer with the given options.
func NewRingBuffer(options ...RingBufferOption) *RingBuffer {
	b := &RingBuffer{
		readers: make(map[chan<- *v1.Event]*ReaderStats),
	}
	for _, o := range options {
		o(b)
	}
	return b
}

// Buffer copies of all the events in r's buffer at the moment of the function
// call into events. If events is nil or its capacity is less than the size of a
// buffer then a new slice is allocated. The returned slice can be re-used in
// later calls to Buffer.
func (b *RingBuffer) Buffer(events []*v1.Event) []*v1.Event {
	if len(b.buffer) == 0 {
		return events[:0]
	}

	b.rwMutex.RLock()
	if b.last == b.next {
		b.rwMutex.RUnlock()
		return nil
	}
	if cap(events) < len(b.buffer) {
		events = make([]*v1.Event, b.next-b.last, len(b.buffer))
	} else if len(events) < b.next-b.last {
		events = append(events, make([]*v1.Event, b.next-b.last-len(events))...)
	}
	headIndex := b.next % len(b.buffer)
	tailIndex := b.last % len(b.buffer)
	if headIndex > tailIndex {
		copy(events, b.buffer[tailIndex:headIndex])
	} else {
		copy(events[0:len(b.buffer)-headIndex], b.buffer[tailIndex:len(b.buffer)])
		copy(events[len(b.buffer)-headIndex:], b.buffer[0:headIndex])
	}
	b.rwMutex.RUnlock()
	return events
}

// ReadAll returns a channel that returns all events in b and then switches to
// follow mode and a cancellation function.
func (b *RingBuffer) ReadAll(capacity int) (<-chan *v1.Event, ReaderCancelFunc) {
	if len(b.buffer) == 0 {
		return b.ReadNew(capacity)
	}

	b.rwMutex.RLock()
	seq := b.last
	b.rwMutex.RUnlock()
	return b.readFrom(seq, capacity)
}

// ReadCurrent returns a channel that returns all events in b and a cancellation
// function.
func (b *RingBuffer) ReadCurrent(capacity int) (<-chan *v1.Event, ReaderCancelFunc) {
	ch := make(chan *v1.Event, capacity)

	// If there is no buffer then we are done.
	if len(b.buffer) == 0 {
		close(ch)
		return ch, func() ReaderStats {
			return ReaderStats{}
		}
	}

	// Acquire a read lock.
	b.rwMutex.RLock()

	// Record the sequence numbers of the current events in the buffer.
	seq := b.last
	last := b.next

	// If the buffer is empty then we are done.
	if seq == last {
		b.rwMutex.RUnlock()
		close(ch)
		return ch, func() ReaderStats {
			return ReaderStats{}
		}
	}

	// Copy the first event from the buffer so that it cannot be overwritten.
	event := b.buffer[seq%len(b.buffer)]

	// Start a goroutine to send events from the ring buffer to the channel.
	// Once all events in the ring buffer have been sent, close the channel.
	var readerStats ReaderStats
	readerReadyCh := make(chan struct{})
	doneCh := make(chan struct{})
	go func() {
		defer close(ch)
		close(readerReadyCh)

		// Release the read lock.
		b.rwMutex.RUnlock()

		for {
			// Send the event to the reader or wait for cancellation.
			select {
			case <-doneCh:
				return
			case ch <- event:
				readerStats.Sent++
				seq++

				// If we have caught up to the last event then we are done.
				if seq >= last {
					return
				}

				// Take a read lock.
				b.rwMutex.RLock()

				// If the reader was slow then we might have dropped events from
				// the ring buffer. Record the number of dropped events and
				// advance to the oldest event in the ring buffer.
				if seq < b.last {
					// If the last event we want is no longer in the ring buffer
					// then we are done.
					if last < b.last {
						b.rwMutex.RUnlock()
						readerStats.Dropped += last - seq
						return
					}
					// Otherwise record the number of dropped events and advance
					// to the last event in the ring buffer.
					readerStats.Dropped += b.last - seq
					seq = b.last
				}

				// Copy the next event from the ring buffer so that it cannot be
				// overwritten.
				event = b.buffer[seq%len(b.buffer)]

				// Release the read lock.
				b.rwMutex.RUnlock()
			}
		}
	}()

	// Wait for the reader to be ready. This ensures that the reader goroutine
	// has started and that the reader has had the opportunity to read at an
	// event from the buffer.
	<-readerReadyCh

	return ch, func() ReaderStats {
		close(doneCh)
		return readerStats
	}
}

// ReadNew returns a channel with the given capacity that sends events written
// to b and a cancellation function. capacity should be zero (unbuffered) except
// in special circumstances (testing). Events will be dropped if the reader of
// the returned channel cannot keep up.
//
// FIXME how to make capacity only available to test code?
func (b *RingBuffer) ReadNew(capacity int) (<-chan *v1.Event, ReaderCancelFunc) {
	ch := make(chan *v1.Event, capacity)
	b.rwMutex.Lock()
	b.readers[ch] = &ReaderStats{}
	b.rwMutex.Unlock()

	return ch, func() ReaderStats {
		b.rwMutex.Lock()
		readerStats := *b.readers[ch]
		delete(b.readers, ch)
		b.rwMutex.Unlock()
		close(ch)
		return readerStats
	}
}

// ReadSince returns a channel with capacity that returns all events since t and
// a cancellation function. t is assumed to be in the past. If t is more recent
// than the last event in the buffer then all new events are returned.
func (b *RingBuffer) ReadSince(t time.Time, capacity int) (<-chan *v1.Event, ReaderCancelFunc) {
	if len(b.buffer) == 0 {
		return b.ReadNew(capacity)
	}

	b.rwMutex.RLock()
	// If there are events in the buffer then scan backwards to find the first
	// event before t and then return events after that event.
	// FIXME replace this linear search with binary search
	// FIXME can improve search by assuming that events are roughly evenly distributed
	for seq := b.next - 1; seq >= b.last; seq-- {
		et := eventTime(b.buffer[seq%len(b.buffer)])
		if !et.IsZero() && et.Before(t) {
			b.rwMutex.RUnlock()
			return b.readFrom(seq+1, capacity)
		}
	}
	b.rwMutex.RUnlock()
	return b.readFrom(0, capacity)
}

// Status returns the status of b.
func (b *RingBuffer) Status() RingBufferStatus {
	b.rwMutex.RLock()
	s := RingBufferStatus{
		NumEvents:  b.next - b.last,
		SeenEvents: b.next,
	}
	for i := b.last; i < b.next; i++ {
		if t := eventTime(b.buffer[i%len(b.buffer)]); !t.IsZero() {
			s.OldestEventTime = t
			break
		}
	}
	for i := b.next - 1; i >= b.last; i-- {
		if t := eventTime(b.buffer[i%len(b.buffer)]); !t.IsZero() {
			s.NewestEventTime = t
			break
		}
	}
	b.rwMutex.RUnlock()
	return s
}

// Write writes event to r.
func (b *RingBuffer) Write(event *v1.Event) {
	b.rwMutex.Lock()
	if len(b.buffer) > 0 {
		b.buffer[b.next%len(b.buffer)] = event
	}
	b.next++
	if b.last < b.next-len(b.buffer) {
		b.last = b.next - len(b.buffer)
	}
	for ch, readerStats := range b.readers {
		select {
		case ch <- event:
			b.sent++
			readerStats.Sent++
		default:
			b.dropped++
			readerStats.Dropped++
		}
	}
	b.rwMutex.Unlock()
}

// readFrom returns a channel with the given capacity that returns events from r
// from seq onwards and a cancellation function.
func (b *RingBuffer) readFrom(seq, capacity int) (<-chan *v1.Event, ReaderCancelFunc) {
	ch := make(chan *v1.Event, capacity)

	// Start a goroutine to send events from the ring buffer to the channel.
	// Once all events in the ring buffer have been sent, switch to follow mode.
	var readerStats ReaderStats
	readerReadyCh := make(chan struct{})
	doneCh := make(chan struct{})
	go func() {
		for {
			// Take a read lock.
			b.rwMutex.RLock()

			// Signal that the reader is ready after taking the read lock for
			// the first time.
			if readerReadyCh != nil {
				close(readerReadyCh)
				readerReadyCh = nil
			}

			// If we have caught up with the most recent event then switch to
			// follow mode.
			if seq == b.next {
				// Release the read lock and acquire the write lock.
				b.rwMutex.RUnlock()
				// FIXME find a way to eliminate this comparison in non-test
				// code
				if b.rUnlockLockFunc != nil {
					b.rUnlockLockFunc()
				}
				b.rwMutex.Lock()
				// Retry the test in case the state changed while the mutex was
				// unlocked.
				if seq == b.next {
					b.readers[ch] = &readerStats
					b.rwMutex.Unlock()
					return
				}
				// Otherwise, release the write lock and re-acquire a read lock.
				b.rwMutex.Unlock()
				b.rwMutex.RLock()
			}

			// If the reader was slow then we might have dropped events from the
			// ring buffer. Record the number of dropped events and advance to
			// the oldest event in the ring buffer.
			if seq < b.last {
				readerStats.Dropped += b.last - seq
				seq = b.last
			}

			// Copy the next event from the ring buffer so that it cannot be
			// overwritten.
			event := b.buffer[seq%len(b.buffer)]

			// Release the read lock.
			b.rwMutex.RUnlock()

			// Send the event to the reader or wait for cancellation.
			select {
			case <-doneCh:
				return
			case ch <- event:
				readerStats.Sent++
				seq++
			}
		}
	}()

	// Wait for the reader to be ready. This ensures that the reader goroutine
	// has started and that the reader has had the opportunity to read at an
	// event from the buffer.
	<-readerReadyCh

	return ch, func() ReaderStats {
		b.rwMutex.Lock()
		delete(b.readers, ch)
		b.rwMutex.Unlock()
		close(doneCh)
		return readerStats
	}
}

// eventTime returns the time of event.
func eventTime(event *v1.Event) time.Time {
	// FIXME verify this
	// FIXME do we need tp extract different times from different events?
	return event.Timestamp.AsTime()
}

// min returns the minimum of a and b.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
