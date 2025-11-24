package memory

import "testing"

func TestLRUList_AddToFront(t *testing.T) {
	list := newLRUList()

	node1 := list.addToFront("key1")
	if list.head != node1 || list.tail != node1 {
		t.Error("First node should be both head and tail")
	}
	if list.len() != 1 {
		t.Errorf("Expected size 1, got %d", list.len())
	}

	node2 := list.addToFront("key2")
	if list.head != node2 {
		t.Error("Second node should be head")
	}
	if list.tail != node1 {
		t.Error("First node should still be tail")
	}
	if list.len() != 2 {
		t.Errorf("Expected size 2, got %d", list.len())
	}
}

func TestLRUList_MoveToFront(t *testing.T) {
	list := newLRUList()

	node1 := list.addToFront("key1")
	node2 := list.addToFront("key2")
	node3 := list.addToFront("key3")

	// Order: key3 -> key2 -> key1

	// Move middle node to front
	list.moveToFront(node2)

	// Order should now be: key2 -> key3 -> key1
	if list.head != node2 {
		t.Error("key2 should be head")
	}
	if list.tail != node1 {
		t.Error("key1 should still be tail")
	}
	if node2.next != node3 {
		t.Error("key2 should point to key3")
	}
}

func TestLRUList_MoveToFront_AlreadyAtFront(t *testing.T) {
	list := newLRUList()

	list.addToFront("key1")
	node2 := list.addToFront("key2")

	// Move head to front (should be no-op)
	list.moveToFront(node2)

	if list.head != node2 {
		t.Error("key2 should still be head")
	}
	if list.len() != 2 {
		t.Errorf("Expected size 2, got %d", list.len())
	}
}

func TestLRUList_Remove(t *testing.T) {
	list := newLRUList()

	node1 := list.addToFront("key1")
	node2 := list.addToFront("key2")
	node3 := list.addToFront("key3")

	// Remove middle node
	list.remove(node2)

	if list.len() != 2 {
		t.Errorf("Expected size 2, got %d", list.len())
	}
	if list.head != node3 {
		t.Error("key3 should still be head")
	}
	if list.tail != node1 {
		t.Error("key1 should still be tail")
	}
	if node3.next != node1 {
		t.Error("key3 should point to key1")
	}
}

func TestLRUList_RemoveHead(t *testing.T) {
	list := newLRUList()

	node1 := list.addToFront("key1")
	node2 := list.addToFront("key2")

	list.remove(node2)

	if list.head != node1 {
		t.Error("key1 should be new head")
	}
	if list.tail != node1 {
		t.Error("key1 should also be tail")
	}
	if list.len() != 1 {
		t.Errorf("Expected size 1, got %d", list.len())
	}
}

func TestLRUList_RemoveTail(t *testing.T) {
	list := newLRUList()

	node1 := list.addToFront("key1")
	node2 := list.addToFront("key2")

	list.remove(node1)

	if list.head != node2 {
		t.Error("key2 should still be head")
	}
	if list.tail != node2 {
		t.Error("key2 should be new tail")
	}
	if list.len() != 1 {
		t.Errorf("Expected size 1, got %d", list.len())
	}
}

func TestLRUList_RemoveLast(t *testing.T) {
	list := newLRUList()

	list.addToFront("key1")
	list.addToFront("key2")
	list.addToFront("key3")

	// Order: key3 -> key2 -> key1

	key := list.removeLast()
	if key != "key1" {
		t.Errorf("Expected to remove key1, got %s", key)
	}
	if list.len() != 2 {
		t.Errorf("Expected size 2, got %d", list.len())
	}

	key = list.removeLast()
	if key != "key2" {
		t.Errorf("Expected to remove key2, got %s", key)
	}
	if list.len() != 1 {
		t.Errorf("Expected size 1, got %d", list.len())
	}

	key = list.removeLast()
	if key != "key3" {
		t.Errorf("Expected to remove key3, got %s", key)
	}
	if list.len() != 0 {
		t.Errorf("Expected size 0, got %d", list.len())
	}
}

func TestLRUList_RemoveLast_EmptyList(t *testing.T) {
	list := newLRUList()

	key := list.removeLast()
	if key != "" {
		t.Errorf("Expected empty string, got %s", key)
	}
}

func TestLRUList_Clear(t *testing.T) {
	list := newLRUList()

	list.addToFront("key1")
	list.addToFront("key2")
	list.addToFront("key3")

	list.clear()

	if list.head != nil || list.tail != nil {
		t.Error("Head and tail should be nil after clear")
	}
	if list.len() != 0 {
		t.Errorf("Expected size 0, got %d", list.len())
	}
}

func TestLRUList_ComplexScenario(t *testing.T) {
	list := newLRUList()

	// Simulate cache access pattern
	nodes := make(map[string]*lruNode)

	// Add items
	nodes["a"] = list.addToFront("a")
	nodes["b"] = list.addToFront("b")
	nodes["c"] = list.addToFront("c")

	// Order: c -> b -> a

	// Access "a" (should move to front)
	list.moveToFront(nodes["a"])
	// Order: a -> c -> b

	if list.head.key != "a" {
		t.Error("a should be at front")
	}

	// Add new item
	nodes["d"] = list.addToFront("d")
	// Order: d -> a -> c -> b

	// Remove least recently used
	removed := list.removeLast()
	if removed != "b" {
		t.Errorf("Expected to remove b, got %s", removed)
	}

	// Order: d -> a -> c

	// Access "c"
	list.moveToFront(nodes["c"])
	// Order: c -> d -> a

	if list.head.key != "c" {
		t.Error("c should be at front")
	}
	if list.tail.key != "a" {
		t.Error("a should be at back")
	}
}
