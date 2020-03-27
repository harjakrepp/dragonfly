package item

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"reflect"
	"strings"
)

// Stack represents a stack of items. The stack shares the same item type and has a count which specifies the
// size of the stack.
type Stack struct {
	item  world.Item
	count int

	customName string
	lore       []string
}

// NewStack returns a new stack using the item type and the count passed. NewStack panics if the count passed
// is negative or if the item type passed is nil.
func NewStack(t world.Item, count int) Stack {
	if count < 0 {
		panic("cannot use negative count for item stack")
	}
	if t == nil {
		panic("cannot have a stack with item type nil")
	}
	return Stack{item: t, count: count}
}

// Count returns the amount of items that is present on the stack. The count is guaranteed never to be
// negative.
func (s Stack) Count() int {
	return s.count
}

// MaxCount returns the maximum count that the stack is able to hold when added to an inventory or when added
// to an item entity.
func (s Stack) MaxCount() int {
	if counter, ok := s.Item().(MaxCounter); ok {
		return counter.MaxCount()
	}
	return 64
}

// Empty checks if the stack is empty (has a count of 0). If this is the case
func (s Stack) Empty() bool {
	return s.Count() == 0
}

// Item returns the item that the stack holds. If the stack is considered empty (Stack.Empty()), Item will
// always return nil.
func (s Stack) Item() world.Item {
	if s.Empty() || s.item == nil {
		return nil
	}
	return s.item
}

// AttackDamage returns the attack damage of the stack. By default, the value returned is 2.0. If the item
// held implements the item.Weapon interface, this damage may be different.
func (s Stack) AttackDamage() float32 {
	if weapon, ok := s.Item().(Weapon); ok {
		return weapon.AttackDamage()
	}
	return 2.0
}

// WithCustomName returns a copy of the Stack with the custom name passed. The custom name is formatted
// according to the rules of fmt.Sprintln.
func (s Stack) WithCustomName(a ...interface{}) Stack {
	s.customName = format(a)
	if nameable, ok := s.Item().(nameable); ok {
		s.item = nameable.WithName(a...)
	}
	return s
}

// CustomName returns the custom name set for the Stack. An empty string is returned if the Stack has no
// custom name set.
func (s Stack) CustomName() string {
	return s.customName
}

// WithLore returns a copy of the Stack with the lore passed. Each string passed is put on a different line,
// where the first string is at the top and the last at the bottom.
// The lore may be cleared by passing no lines into the Stack.
func (s Stack) WithLore(lines ...string) Stack {
	s.lore = lines
	return s
}

// Lore returns the lore set for the Stack. If no lore is present, the slice returned has a len of 0.
func (s Stack) Lore() []string {
	return s.lore
}

// AddStack adds another stack to the stack and returns both stacks. The first stack returned will have as
// many items in it as possible to fit in the stack, according to a max count of either 64 or otherwise as
// returned by Item.MaxCount(). The second stack will have the leftover items: It may be empty if the count of
// both stacks together don't exceed the max count.
// If the two stacks are not comparable, AddStack will return both the original stack and the stack passed.
func (s Stack) AddStack(s2 Stack) (a, b Stack) {
	if !s.Comparable(s2) {
		// The items are not comparable and thus cannot be stacked together.
		return s, s2
	}
	if s.Count() >= s.MaxCount() {
		// No more items could be added to the original stack.
		return s, s2
	}
	diff := s.MaxCount() - s.Count()
	if s2.Count() < diff {
		diff = s2.Count()
	}

	s.count, s2.count = s.count+diff, s2.count-diff
	return s, s2
}

// Grow grows the Stack's count by n, returning the resulting Stack. If a positive number is passed, the stack
// is grown, whereas if a negative size is passed, the resulting Stack will have a lower count. The count of
// the returned Stack will never be negative.
func (s Stack) Grow(n int) Stack {
	s.count += n
	if s.count < 0 {
		s.count = 0
	}
	return s
}

// Comparable checks if two stacks can be considered comparable. True is returned if the two stacks have an
// equal item type and have equal enchantments, lore and custom names, or if one of the stacks is empty.
func (s Stack) Comparable(s2 Stack) bool {
	if s.Empty() || s2.Empty() {
		return true
	}
	id, meta := s.Item().EncodeItem()
	id2, meta2 := s2.Item().EncodeItem()
	if id != id2 || meta != meta2 {
		return false
	}
	if s.customName != s2.customName || len(s.lore) != len(s2.lore) {
		return false
	}
	for i := range s.lore {
		if s.lore[i] != s2.lore[i] {
			return false
		}
	}
	if nbt, ok := s.Item().(world.NBTer); ok {
		nbt2, ok := s2.Item().(world.NBTer)
		if !ok {
			return false
		}
		return reflect.DeepEqual(nbt.EncodeNBT(), nbt2.EncodeNBT())
	}
	return true
}

// String implements the fmt.Stringer interface.
func (s Stack) String() string {
	if s.item == nil {
		return fmt.Sprintf("Stack<nil> x%v", s.count)
	}
	return fmt.Sprintf("Stack<%T%+v>(custom name='%v', lore='%v') x%v", s.item, s.item, s.customName, s.lore, s.count)
}

// format is a utility function to format a list of values to have spaces between them, but no newline at the
// end, which is typically used for sending messages, popups and tips.
func format(a []interface{}) string {
	return strings.TrimSuffix(fmt.Sprintln(a...), "\n")
}