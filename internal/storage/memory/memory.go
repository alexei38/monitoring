package memory

type Storage interface {
	Len() int
	Front() *Item
	List() []*Item
	Back() *Item
	PushFront(v interface{}) *Item
	Remove(i *Item)
	PushBack(v interface{}) *Item
	MoveToFront(i *Item)
}

type Item struct {
	Value interface{}
	Next  *Item
	Prev  *Item
}

type storage struct {
	lastNode  *Item
	firstNode *Item
	len       int
}

func (l *storage) Len() int {
	return l.len
}

func (l *storage) List() []*Item {
	var items []*Item
	for i := l.Front(); i != nil; i = i.Next {
		items = append(items, i)
	}
	return items
}

func (l *storage) Front() *Item {
	return l.firstNode
}

func (l *storage) Back() *Item {
	return l.lastNode
}

func (l *storage) PushFront(v interface{}) *Item {
	var item *Item
	switch res := v.(type) {
	case *Item:
		item = res
	default:
		item = &Item{Value: v}
	}
	if l.firstNode != nil {
		item.Next = l.firstNode
		item.Next.Prev = item
	}
	l.firstNode = item
	if l.lastNode == nil {
		l.lastNode = item
	}
	l.len++
	return item
}

func (l *storage) PushBack(v interface{}) *Item {
	var item *Item
	switch res := v.(type) {
	case *Item:
		item = res
	default:
		item = &Item{Value: v}
	}

	if l.firstNode == nil { // new list
		l.firstNode = item
	}
	if l.lastNode != nil {
		l.lastNode.Next = item
		item.Prev = l.lastNode
	}
	l.lastNode = item
	l.len++
	return item
}

func (l *storage) Remove(i *Item) {
	if i == nil || (i.Next == nil && i.Prev == nil) {
		return
	}
	if i == l.lastNode && i.Prev == nil {
		l.lastNode = nil
	}
	if i == l.firstNode && i.Next == nil {
		l.firstNode = nil
	}
	if i.Prev != nil {
		i.Prev.Next = i.Next
		if i == l.lastNode {
			l.lastNode = i.Prev
		}
	}
	if i.Next != nil {
		i.Next.Prev = i.Prev
		if i == l.firstNode {
			l.firstNode = i.Next
		}
	}
	i.Prev = nil
	i.Next = nil
	l.len--
}

func (l *storage) MoveToFront(i *Item) {
	if i.Prev == nil && i.Next == nil { // already deleted
		return
	}
	if l.firstNode == i {
		return
	}
	i.Prev.Next = i.Next
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.lastNode = i.Prev
	}
	i.Next = l.firstNode
	l.firstNode.Prev = i
	l.firstNode = i
}

func NewStorage() Storage {
	var s Storage = &storage{}
	return s
}
