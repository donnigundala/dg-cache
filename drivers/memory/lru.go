package memory

// lruNode represents a node in the LRU doubly-linked list.
type lruNode struct {
	key  string
	prev *lruNode
	next *lruNode
}

// lruList manages the LRU ordering using a doubly-linked list.
// The most recently used item is at the front, least recently used at the back.
type lruList struct {
	head *lruNode
	tail *lruNode
	size int
}

// newLRUList creates a new LRU list.
func newLRUList() *lruList {
	return &lruList{}
}

// moveToFront moves an existing node to the front of the list.
func (l *lruList) moveToFront(node *lruNode) {
	if node == l.head {
		// Already at front
		return
	}

	// Remove from current position
	l.remove(node)

	// Add to front
	l.addToFrontNode(node)
}

// addToFront creates a new node and adds it to the front of the list.
// Returns the created node.
func (l *lruList) addToFront(key string) *lruNode {
	node := &lruNode{key: key}
	l.addToFrontNode(node)
	return node
}

// addToFrontNode adds an existing node to the front of the list.
func (l *lruList) addToFrontNode(node *lruNode) {
	if l.head == nil {
		// Empty list
		l.head = node
		l.tail = node
		node.prev = nil
		node.next = nil
	} else {
		// Add to front
		node.next = l.head
		node.prev = nil
		l.head.prev = node
		l.head = node
	}
	l.size++
}

// remove removes a node from the list.
func (l *lruList) remove(node *lruNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		// Removing head
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		// Removing tail
		l.tail = node.prev
	}

	l.size--
}

// removeLast removes and returns the key of the least recently used item.
// Returns empty string if the list is empty.
func (l *lruList) removeLast() string {
	if l.tail == nil {
		return ""
	}

	key := l.tail.key
	l.remove(l.tail)
	return key
}

// clear removes all nodes from the list.
func (l *lruList) clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
}

// len returns the number of nodes in the list.
func (l *lruList) len() int {
	return l.size
}
