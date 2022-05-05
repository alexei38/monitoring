package memory

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewStorage()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewStorage()

		firstItem := l.PushFront(10) // [10]
		require.NotNil(t, firstItem)
		l.PushBack(20) // [10, 20]

		lastItem := l.PushBack(30) // [10, 20, 30]
		require.NotNil(t, lastItem)
		require.Equal(t, 3, l.Len())
		require.Equal(t, 10, l.Front().Value)
		require.Equal(t, 30, l.Back().Value)

		middle := l.Front().Next // 20
		require.Equal(t, 20, middle.Value)

		l.Remove(middle) // [10, 30]
		require.Equal(t, 2, l.Len())
		require.Equal(t, 10, l.Front().Value)
		require.Equal(t, 30, l.Back().Value)

		require.Equal(t, firstItem, l.Front())
		require.Equal(t, lastItem, l.Back())

		require.Equal(t, firstItem.Next, lastItem)
		require.Equal(t, lastItem.Prev, firstItem)

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)
		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})
	t.Run("pushBackFirst", func(t *testing.T) {
		l := NewStorage()
		firstItem := l.PushBack(10) // [10]
		require.Nil(t, firstItem.Prev)
		require.Nil(t, firstItem.Next)
		require.Equal(t, 1, l.Len())

		secondItem := l.PushBack(20) // [10, 20]
		require.Nil(t, firstItem.Prev)
		require.Equal(t, firstItem.Next, secondItem)
		require.Equal(t, secondItem.Prev, firstItem)
		require.Nil(t, secondItem.Next)

		require.Equal(t, 2, l.Len())
		require.Equal(t, l.Front().Value, 10)
		require.Equal(t, l.Back().Value, 20)
	})

	t.Run("removeItem", func(t *testing.T) {
		l := NewStorage()
		l.PushFront(10)
		l.PushBack(20)
		l.PushBack(30)
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next

		// Удаляем средний элемент
		// Проверяем, что Next и Prev у первого и последнего элемента изменились
		l.Remove(middle)
		require.Equal(t, 2, l.Len())
		require.Equal(t, l.Front().Value, 10)
		require.Equal(t, l.Back().Value, 30)
		require.Equal(t, l.Front(), l.Back().Prev)
		require.Equal(t, l.Front().Next, l.Back())
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)

		// Еще раз удаляем middle элемент. Ничего не должно поменяться
		l.Remove(middle)
		require.Equal(t, 2, l.Len())
		require.Equal(t, l.Front().Value, 10)
		require.Equal(t, l.Back().Value, 30)

		// При удалении пустого элемента нет паники и количество остается прежним
		l.Remove(nil)
		require.Equal(t, 2, l.Len())

		// Удаляем первый элемент
		removeitem := l.Front()
		l.Remove(removeitem)
		require.Equal(t, 1, l.Len())
		require.Equal(t, l.Front().Value, 30)
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Front().Next)
		require.NotEqual(t, removeitem, l.Front())
		require.Equal(t, l.Front(), l.Back())
		require.Nil(t, removeitem.Prev)
		require.Nil(t, removeitem.Next)
	})

	t.Run("toFront", func(t *testing.T) {
		l := NewStorage()
		l.PushFront(10)
		l.PushBack(20)
		l.PushBack(30)
		require.Equal(t, 3, l.Len()) // 10, 20, 30

		last := l.Back()
		middle := last.Prev

		// Переменстим в начало
		l.MoveToFront(last)
		require.NotNil(t, l.Back())
		require.NotEqual(t, last, l.Back())
		require.Equal(t, last, l.Front())
		require.Equal(t, middle, l.Back())

		// Переместим второй раз, ничего не должно поменяться
		l.MoveToFront(last)
		require.Equal(t, last, l.Front())

		// Удалим объект и попробуем еще раз переместить
		l.Remove(last)
		require.NotEqual(t, last, l.Front())
		require.Nil(t, last.Next)
		require.Nil(t, last.Prev)

		// Переместить в начало, можно только существующие в списке
		l.MoveToFront(last)
		require.NotEqual(t, last, l.Front())
		require.Nil(t, last.Next)
		require.Nil(t, last.Prev)
	})
}
