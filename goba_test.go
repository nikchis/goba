// Copyright (c) 2022 Nikita Chisnikov <chisnikov@gmail.com>
// Distributed under the MIT/X11 software license
package goba

import "testing"

func TestBitArraySetGetRemove(t *testing.T) {
	ba := New(128, false)

	ba.Set(0)
	ba.Set(1)
	ba.Set(63)
	ba.Set(64)
	ba.Set(125)
	ba.Set(127)

	t.Log(ba.sprint())

	if !ba.Get(0) {
		t.Fatalf("failed on test case 1")
	}
	if !ba.Get(63) {
		t.Fatalf("failed on test case 2")
	}
	if !ba.Get(64) {
		t.Fatalf("failed on test case 3")
	}
	if !ba.Get(125) {
		t.Fatalf("failed on test case 4")
	}
	if ba.Get(126) {
		t.Fatalf("failed on test case 5")
	}
	if ba.Get(128) {
		t.Fatalf("failed on test case 6")
	}
	if ba.Count() != 6 {
		t.Fatalf("failed on test case 7")
	}
	ba.Remove(125)
	if ba.Get(125) {
		t.Fatalf("failed on test case 8")
	}
	ba.Remove(126)
	if ba.Get(126) {
		t.Fatalf("failed on test case 9")
	}
}

func TestBitArraySetAllRemoveAll(t *testing.T) {
	ba := New(67, false)

	ba.SetAll()
	t.Log(ba.sprint())

	if ba.Count() != 67 {
		t.Fatalf("failed on test case 1")
	}

	ba.RemoveAll()
	if ba.Count() != 0 {
		t.Fatalf("failed on test case 2")
	}
}

func TestBitArraySetGetConcurrent(t *testing.T) {
	ba := New(128, true)

	ba.Set(0)
	ba.Set(1)
	ba.Set(63)
	ba.Set(64)
	ba.Set(125)
	ba.Set(127)

	t.Log(ba.sprint())

	if !ba.Get(0) {
		t.Fatalf("failed on test case 1")
	}
	if !ba.Get(63) {
		t.Fatalf("failed on test case 2")
	}
	if !ba.Get(64) {
		t.Fatalf("failed on test case 3")
	}
	if !ba.Get(125) {
		t.Fatalf("failed on test case 4")
	}
	if ba.Get(126) {
		t.Fatalf("failed on test case 5")
	}
	if ba.Get(128) {
		t.Fatalf("failed on test case 6")
	}
	if ba.Count() != 6 {
		t.Fatalf("failed on test case 7")
	}
}

func TestBitArraySetAllRemoveAllConcurent(t *testing.T) {
	ba := New(67, true)

	ba.SetAll()
	t.Log(ba.sprint())

	if ba.Count() != 67 {
		t.Fatalf("failed on test case 1")
	}

	ba.RemoveAll()
	if ba.Count() != 0 {
		t.Fatalf("failed on test case 2")
	}
}

func BenchmarkSet(b *testing.B) {
	ba := New(1<<31, false)
	for i := 0; i < b.N; i++ {
		if i < ba.Len() {
			ba.Set(i)
		}
	}
}

func BenchmarkGet(b *testing.B) {
	ba := New(1<<31, false)
	ba.SetAll()
	for i := 0; i < b.N; i++ {
		if i < ba.Len() {
			ba.Get(i)
		}
	}
}

func BenchmarkCount(b *testing.B) {
	ba := New(1<<10, false)
	ba.Set(63)
	ba.Set(64)
	ba.Set(125)
	ba.Set(127)
	for i := 0; i < b.N; i++ {
		ba.Count()
	}
}

func BenchmarkCountSetAll(b *testing.B) {
	ba := New(1<<10, false)
	ba.SetAll()
	for i := 0; i < b.N; i++ {
		ba.Count()
	}
}

func BenchmarkSetConcurrent(b *testing.B) {
	ba := New(1<<31, true)
	for i := 0; i < b.N; i++ {
		if i < ba.Len() {
			ba.Set(i)
		}
	}
}

func BenchmarkGetConcurrent(b *testing.B) {
	ba := New(1<<31, true)
	ba.SetAll()
	for i := 0; i < b.N; i++ {
		if i < ba.Len() {
			ba.Get(i)
		}
	}
}

func BenchmarkCountConcurrent(b *testing.B) {
	ba := New(1<<10, true)
	ba.Set(63)
	ba.Set(64)
	ba.Set(125)
	ba.Set(127)
	for i := 0; i < b.N; i++ {
		ba.Count()
	}
}

func BenchmarkCountSetAllConcurrent(b *testing.B) {
	ba := New(1<<10, true)
	ba.SetAll()
	for i := 0; i < b.N; i++ {
		ba.Count()
	}
}

func TestBitArrayUnify(t *testing.T) {
	ba1 := New(64, true)
	ba2 := New(128, true)

	ba1.Set(0)
	ba2.Set(0)
	ba2.Set(1)
	ba1.Set(63)
	ba2.Set(64)
	ba2.Set(127)

	t.Log(ba1.sprint())
	t.Log(ba2.sprint())
	ba3 := ba1.UnifyWith(ba2)
	t.Log(ba3.sprint())

	if ba3.Count() != 5 {
		t.Fatalf("failed on test case 1")
	}

}

func TestBitArrayIntersect(t *testing.T) {
	ba1 := New(128, false)
	ba2 := New(64, false)

	ba1.Set(0)
	ba2.Set(1)
	ba1.Set(63)
	ba2.Set(63)
	ba1.Set(125)
	ba2.Set(12)

	t.Log(ba1.sprint())
	t.Log(ba2.sprint())
	ba3 := ba1.IntersectWith(ba2)
	t.Log(ba3.sprint())

	if ba3.Count() != 1 {
		t.Fatalf("failed on test case 1")
	}

}
