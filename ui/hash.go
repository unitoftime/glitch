package ui

import (
	"hash/crc64"
	"strings"
)

var hashTable = crc64.MakeTable(crc64.ISO)

func crc(label string) uint64 {
	return crc64.Checksum([]byte(label), hashTable)
}
func bumpCrc(crc uint64, bump []byte) uint64 {
	return crc64.Update(crc, hashTable, bump)
}

func getIdNoBump(label string) eid {
	crc := crc(label)

	id, ok := global.elements[crc]
	if !ok {
		id = global.idCounter
		global.idCounter++
		global.elements[crc] = id
		// g.elementsRev[id] = label
	}

	return id
}

func getId(label string) eid {
	crc := crc(label)

	bump, alreadyFetched := global.dedup[crc]
	if alreadyFetched {
		global.dedup[crc] = bump + 1
		crc = bumpCrc(crc, []byte{uint8(bump)})
		// label = fmt.Sprintf("%s##%d", label, bump)
		// fmt.Printf("duplicate label, using bump: %s\n", label)
		// panic(fmt.Sprintf("duplicate label found: %s", label))
	} else {
		global.dedup[crc] = 0
	}

	id, ok := global.elements[crc]
	if !ok {
		id = global.idCounter
		global.idCounter++
		global.elements[crc] = id
		// g.elementsRev[id] = label
	}

	return id
}

func removeDedup(label string) string {
	ret, _, _ := strings.Cut(label, "##")
	return ret
}
