// Copyright (c) 2022 Nikita Chisnikov <chisnikov@gmail.com>
// Distributed under the MIT/X11 software license
package goba

import (
	"fmt"
	"math/bits"
	"sync/atomic"
	"unsafe"
)

var isLE bool

func init() {
	var x uint16 = 0xff00
	xb := *(*[2]byte)(unsafe.Pointer(&x))
	isLE = (xb[0] == 0x00)
}

type BitArray struct {
	left       int64 // left boundary
	right      int64 // right boundary
	length     int64 // length in bits
	concurrent bool
	data       []uint64
}

// New returns an instantiated BitArray struct.
//
// length in bits, concurrent for concurrent safe usage
func New(length int, concurrent bool) *BitArray {
	res := BitArray{
		length:     int64(length),
		concurrent: concurrent,
	}
	res.data = make([]uint64, (length+63)/64)
	return &res
}

// Length of BitArray in bits
func (s *BitArray) Len() int {
	if s.concurrent {
		return int(atomic.LoadInt64(&s.length))
	} else {
		return int(s.length)
	}
}

// Set bit at index
func (s *BitArray) Set(index int) {
	if s.concurrent {
		s.setAtomically(index)
	} else {
		s.set(index)
	}
}

func (s *BitArray) set(index int) {
	if s == nil || index >= int(s.length) || index < 0 {
		return
	}
	var i int64 = int64(index >> 6)
	s.data[i] |= (1 << (index & 0x3f))
	if s.right < i {
		s.right = i
	}
	if s.left > i {
		s.left = i
	}
}

func (s *BitArray) setAtomically(index int) {
	if s == nil || index >= int(atomic.LoadInt64(&s.length)) || index < 0 {
		return
	}
	var i int64 = int64(index >> 6)
	var v uint64 = atomic.LoadUint64(&s.data[i])
	atomic.StoreUint64(&s.data[i], v|(1<<(index&0x3f)))
	if atomic.LoadInt64(&s.right) < i {
		atomic.StoreInt64(&s.right, i)
	}
	if atomic.LoadInt64(&s.left) > i {
		atomic.StoreInt64(&s.left, i)
	}
}

// Set all bits to 1
func (s *BitArray) SetAll() {
	if s.concurrent {
		s.setAllAtomically()
	} else {
		s.setAll()
	}
}

func (s *BitArray) setAll() {
	if s == nil {
		return
	}
	for i := range s.data {
		if i < len(s.data)-1 {
			s.data[i] = 0xffffffffffffffff
		} else {
			s.data[i] = (1<<(s.length&0x3f) - 1)
		}
	}
	s.left = 0
	s.right = int64(len(s.data)) - 1
}

func (s *BitArray) setAllAtomically() {
	if s == nil {
		return
	}
	for i := range s.data {
		if i < len(s.data)-1 {
			atomic.StoreUint64(&s.data[i], 0xffffffffffffffff)
		} else {
			atomic.StoreUint64(&s.data[i], (1<<(atomic.LoadInt64(&s.length)&0x3f) - 1))
		}
	}
	atomic.StoreInt64(&s.left, 0)
	atomic.StoreInt64(&s.right, int64(len(s.data))-1)
}

// Remove bit at index
func (s *BitArray) Remove(index int) {
	if s.concurrent {
		s.removeAtomically(index)
	} else {
		s.remove(index)
	}
}

func (s *BitArray) remove(index int) {
	if s == nil || index >= int(s.length) || index < 0 {
		return
	}
	var i int64 = int64(index >> 6)
	s.data[i] &^= (1 << (index & 0x3f))
	if s.right < i {
		s.right = i
	}
	if s.left > i {
		s.left = i
	}
}

func (s *BitArray) removeAtomically(index int) {
	if s == nil || index >= int(atomic.LoadInt64(&s.length)) || index < 0 {
		return
	}
	var i int64 = int64(index >> 6)
	var v uint64 = atomic.LoadUint64(&s.data[i])
	atomic.StoreUint64(&s.data[i], v&^(1<<(index&0x3f)))
	if atomic.LoadInt64(&s.right) < i {
		atomic.StoreInt64(&s.right, i)
	}
	if atomic.LoadInt64(&s.left) > i {
		atomic.StoreInt64(&s.left, i)
	}
}

// Remove all bits
func (s *BitArray) RemoveAll() {
	if s.concurrent {
		s.removeAllAtomically()
	} else {
		s.removeAll()
	}
}

func (s *BitArray) removeAll() {
	if s == nil {
		return
	}
	for i := range s.data {
		s.data[i] = 0x0000000000000000
	}
	s.left = 0
	s.right = 0
}

func (s *BitArray) removeAllAtomically() {
	if s == nil {
		return
	}
	for i := range s.data {
		atomic.StoreUint64(&s.data[i], 0x0000000000000000)
	}
	atomic.StoreInt64(&s.left, 0)
	atomic.StoreInt64(&s.right, 0)
}

// Get bit value at index
// 1 - true, 0 - false
func (s *BitArray) Get(index int) bool {
	if s.concurrent {
		return s.getAtomically(index)
	} else {
		return s.get(index)
	}
}

func (s *BitArray) get(index int) bool {
	if index >= int(s.length) || index < 0 {
		return false
	}
	return ((s.data[index>>6] >> ((index) & 0x3f)) & 1) == 1
}

func (s *BitArray) getAtomically(index int) bool {
	if index >= int(atomic.LoadInt64(&s.length)) || index < 0 {
		return false
	}
	return ((atomic.LoadUint64(&s.data[index>>6]) >> ((index) & 0x3f)) & 1) == 1
}

// Count of nonzero bits
func (s *BitArray) Count() int {
	if s.concurrent {
		return s.count12Atomically()
	} else {
		return s.count12()
	}
}

func (s *BitArray) count12() int {
	var cnt uint64
	for _, v := range s.data {
		if v > 0 {
			v -= (v >> 1) & 0x5555555555555555
			v = (v & 0x3333333333333333) + ((v >> 2) & 0x3333333333333333)
			v = (v + (v >> 4)) & 0x0f0f0f0f0f0f0f0f
			cnt += (v * 0x0101010101010101) >> 56
		}
	}
	return int(cnt)
}

func (s *BitArray) count12Atomically() int {
	var cnt uint64
	for i := range s.data {
		v := atomic.LoadUint64(&s.data[i])
		if v > 0 {
			v -= (v >> 1) & 0x5555555555555555
			v = (v & 0x3333333333333333) + ((v >> 2) & 0x3333333333333333)
			v = (v + (v >> 4)) & 0x0f0f0f0f0f0f0f0f
			cnt += (v * 0x0101010101010101) >> 56
		}
	}
	return int(cnt)
}

func (s *BitArray) count17() int {
	var cnt uint64
	for _, v := range s.data {
		if v > 0 {
			v -= (v >> 1) & 0x5555555555555555
			v = (v & 0x3333333333333333) + ((v >> 2) & 0x3333333333333333)
			v = (v + (v >> 4)) & 0x0f0f0f0f0f0f0f0f
			v += v >> 8
			v += v >> 16
			v += v >> 32
			cnt += v & 0x7f
		}
	}
	return int(cnt)
}

func (s *BitArray) count17Atomically() int {
	var cnt uint64
	for i := range s.data {
		v := atomic.LoadUint64(&s.data[i])
		if v > 0 {
			v -= (v >> 1) & 0x5555555555555555
			v = (v & 0x3333333333333333) + ((v >> 2) & 0x3333333333333333)
			v = (v + (v >> 4)) & 0x0f0f0f0f0f0f0f0f
			v += v >> 8
			v += v >> 16
			v += v >> 32
			cnt += v & 0x7f
		}
	}
	return int(cnt)
}

func (s *BitArray) sprint() string {
	var res string
	if isLE {
		for i := range s.data {
			res = fmt.Sprintf("%s[%064b]", res, bits.Reverse64(s.data[i]))
		}
	} else {
		for i := range s.data {
			res = fmt.Sprintf("%s[%064b]", res, s.data[i])
		}
	}
	return res
}

// Return union of BitArrays
func (s *BitArray) UnifyWith(ba *BitArray) *BitArray {
	if s.concurrent || ba.concurrent {
		return s.unifyWithAtomically(ba)
	} else {
		return s.unifyWith(ba)
	}
}

func (s *BitArray) unifyWith(ba *BitArray) *BitArray {
	var res *BitArray
	if len(s.data) >= len(ba.data) {
		res = New(int(s.length), s.concurrent)
		copy(res.data, s.data)
		for i := range ba.data {
			res.data[i] |= ba.data[i]
		}
		if res.length < ba.length {
			res.length = ba.length
		}
	} else {
		res = New(int(ba.length), s.concurrent)
		copy(res.data, ba.data)
		for i := range s.data {
			res.data[i] |= s.data[i]
		}
		res.length = ba.length
	}
	if ba.left < s.left {
		res.left = ba.left
	} else {
		res.left = s.left
	}
	if ba.right > s.right {
		res.right = ba.right
	} else {
		res.right = s.right
	}
	return res
}

func (s *BitArray) unifyWithAtomically(ba *BitArray) *BitArray {
	var res *BitArray
	if len(s.data) >= len(ba.data) {
		res = New(int(atomic.LoadInt64(&s.length)), s.concurrent)
		copy(res.data, s.data)
		for i := range ba.data {
			res.data[i] |= atomic.LoadUint64(&ba.data[i])
		}
		if res.length < atomic.LoadInt64(&ba.length) {
			res.length = atomic.LoadInt64(&ba.length)
		}
	} else {
		res = New(int(atomic.LoadInt64(&ba.length)), s.concurrent)
		copy(res.data, ba.data)
		for i := range s.data {
			res.data[i] |= atomic.LoadUint64(&s.data[i])
		}
		res.length = atomic.LoadInt64(&ba.length)
	}
	if atomic.LoadInt64(&ba.left) < atomic.LoadInt64(&s.left) {
		res.left = atomic.LoadInt64(&ba.left)
	} else {
		res.left = atomic.LoadInt64(&s.left)
	}
	if atomic.LoadInt64(&ba.right) > atomic.LoadInt64(&s.right) {
		res.right = atomic.LoadInt64(&ba.right)
	} else {
		res.right = atomic.LoadInt64(&s.right)
	}
	return res
}

// Return intersection of BitArrays
func (s *BitArray) IntersectWith(ba *BitArray) *BitArray {
	if s.concurrent {
		return s.intersectWithAtomically(ba)
	} else {
		return s.intersectWith(ba)
	}
}

func (s *BitArray) intersectWith(ba *BitArray) *BitArray {
	if s == nil || ba == nil {
		return nil
	}
	var res *BitArray
	if s.length < ba.length {
		res = New(int(s.length), s.concurrent)
	} else {
		res = New(int(ba.length), s.concurrent)
	}
	var left, right int64
	if s.left < ba.left {
		left = ba.left
	} else {
		left = s.left
	}
	if s.right > ba.right {
		right = ba.right
	} else {
		right = s.right
	}
	for i := left; i <= right && i < int64(len(res.data)); i++ {
		res.data[i] = s.data[i] & ba.data[i]
	}
	res.left = left
	res.right = right

	return res
}

func (s *BitArray) intersectWithAtomically(ba *BitArray) *BitArray {
	if s == nil || ba == nil {
		return nil
	}
	var res *BitArray
	if s.length < ba.length {
		res = New(int(atomic.LoadInt64(&s.length)), s.concurrent)
	} else {
		res = New(int(atomic.LoadInt64(&ba.length)), s.concurrent)
	}
	var left, right int64
	if atomic.LoadInt64(&s.left) < atomic.LoadInt64(&ba.left) {
		left = atomic.LoadInt64(&ba.left)
	} else {
		left = atomic.LoadInt64(&s.left)
	}
	if atomic.LoadInt64(&s.right) > atomic.LoadInt64(&ba.right) {
		right = atomic.LoadInt64(&ba.right)
	} else {
		right = atomic.LoadInt64(&s.right)
	}
	for i := left; i <= right && i < int64(len(res.data)); i++ {
		res.data[i] = atomic.LoadUint64(&s.data[i]) & atomic.LoadUint64(&ba.data[i])
	}
	res.left = left
	res.right = right

	return res
}

// Check for intersection with BitArray
func (s *BitArray) HasIntersectionWith(ba *BitArray) bool {
	return s.hasIntersectionWith(ba)
}

func (s *BitArray) hasIntersectionWith(ba *BitArray) bool {
	if s == nil || ba == nil ||
		atomic.LoadInt64(&s.left) > atomic.LoadInt64(&ba.right) ||
		atomic.LoadInt64(&s.right) < atomic.LoadInt64(&ba.left) {
		return false
	}
	var left, right int64
	if atomic.LoadInt64(&s.left) < atomic.LoadInt64(&ba.left) {
		left = atomic.LoadInt64(&ba.left)
	} else {
		left = atomic.LoadInt64(&s.left)
	}
	if atomic.LoadInt64(&s.right) > atomic.LoadInt64(&ba.right) {
		right = atomic.LoadInt64(&ba.right)
	} else {
		right = atomic.LoadInt64(&s.right)
	}
	for i := left; i <= right && i < int64(len(s.data)) && i < int64(len(ba.data)); i++ {
		if atomic.LoadUint64(&s.data[i])&atomic.LoadUint64(&ba.data[i]) != 0 {
			return true
		}
	}

	return false
}
