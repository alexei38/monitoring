package memory

type Storage interface {
	Len() int
	Front() *Item
	List() []*Item
	Back() *Item
	PushFront(v interface{}) *Item
	Remove(i *Item)
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
	if l.Len() == 0 {
		return items
	}
	first := l.firstNode
	items = append(items, first)
	for first.Next != nil {
		items = append(items, first.Next)
		first = first.Next
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
	if l.firstNode == nil {
		l.firstNode = item
		l.lastNode = item
	} else {
		l.firstNode.Prev = item
	}
	l.firstNode = item
	l.len++
	return item
}

func (l *storage) Remove(i *Item) {
	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.lastNode = i
	}
	if i.Prev != nil {
		i.Prev.Prev = i.Next
	} else {
		l.firstNode = i
	}
	l.len--
}

func NewStorage() Storage {
	var s Storage = &storage{}
	return s
}
