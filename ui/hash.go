package ui

import (
	"encoding/binary"
	"hash/crc64"
	"strings"

	"golang.org/x/exp/constraints"
)

// TODO: You could potententially do an ID stack system: https://github.com/ocornut/imgui/blob/master/docs/FAQ.md#q-about-the-id-stack-system
func PushId[T constraints.Integer](id T) {
	currentIndex := global.stackIndex
	global.stackIndex++

	// If not enough room in slice, add a new one
	if currentIndex >= len(global.idStack) {
		global.idStack = append(global.idStack, make([]byte, 0, 8))
	}

	// Convert the data to a byte slice
	buf := global.idStack[currentIndex]
	buf = buf[:0]
	buf = binary.AppendUvarint(buf, uint64(id))

	global.idStack[currentIndex] = buf
}

func PopId() {
	// Note: Ignore unmatched pops so that the next PushId doesn't index out of range
	global.stackIndex = max(global.stackIndex-1, 0)
}

func updateCrcWithIdStack(current uint64) uint64 {
	for i := 0; i < global.stackIndex; i++ {
		current = bumpCrc(current, global.idStack[i])
	}
	return current
}

type eid uint64 // Element Id
const invalidId eid = 0

var hashTable = crc64.MakeTable(crc64.ISO)

func crc(label string) uint64 {
	return crc64.Checksum([]byte(label), hashTable)
}
func bumpCrc(crc uint64, bump []byte) uint64 {
	return crc64.Update(crc, hashTable, bump)
}

// Element ids are the 64 bit hash itself, so they don't need to be stored anywhere
// Note: Two labels could collide, but with 64 bits that is extremely unlikely
func hashToId(crc uint64) eid {
	if eid(crc) == invalidId {
		return invalidId + 1
	}
	return eid(crc)
}

// Note: This must match the id that getId returns for the first use of the label in a frame (outside of any PushId)
func getIdNoBump(label string) eid {
	return hashToId(crc(label))
}

func getId(label string) eid {
	crc := crc(label)
	crc = updateCrcWithIdStack(crc)

	bump, alreadyFetched := global.dedup[crc]
	if alreadyFetched {
		global.dedup[crc] = bump + 1

		// Note: The full bump is hashed so that ids don't repeat after 256 duplicates
		var bumpBuf [binary.MaxVarintLen32]byte
		n := binary.PutUvarint(bumpBuf[:], uint64(bump))
		crc = bumpCrc(crc, bumpBuf[:n])
		// label = fmt.Sprintf("%s##%d", label, bump)
		// fmt.Printf("duplicate label, using bump: %s\n", label)
		// panic(fmt.Sprintf("duplicate label found: %s", label))
	} else {
		global.dedup[crc] = 0
	}

	// global.cachedId = id
	return hashToId(crc)
}

func removeDedup(label string) string {
	ret, _, _ := strings.Cut(label, "##")
	return ret
}

// Idea 1: To minimize the fmt.Sprintf calls with Ids embedded. you could construct Ids on the fly
// func GetId[T constraints.Ordered](label string, val T) eid {
// 	crc := crc(label)

// 	buf := make([]byte, 0, 8) // Note: causes an alloc
// 	buf, err := binary.Append(buf, binary.LittleEndian, val)
// 	if err != nil {
// 		panic(err)
// 	}
// 	crc = bumpCrc(crc, buf)

// 	id, ok := global.elements[crc]
// 	if !ok {
// 		id = global.idCounter
// 		global.idCounter++
// 		global.elements[crc] = id
// 		// g.elementsRev[id] = label
// 	}

// 	return id
// }

// func (id eid) TextExt(label string, rect glitch.Rect, textStyle TextStyle) glitch.Rect {
// 	style := Style{
// 		Text: textStyle,
// 	}

// 	mask := wmDrawText
// 	text := removeDedup(label)
// 	resp := doWidget(id, text, mask, style, rect)
// 	return resp.textRect
// }

// func (id eid) ButtonExt(label string, rect glitch.Rect, style Style) bool {
// 	mask := wmHoverable | wmClickable | wmDrawPanel | wmDrawText
// q	text := removeDedup(label)
// 	resp := doWidget(id, text, mask, style, rect)

// 	return resp.Released
// }
