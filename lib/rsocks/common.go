package rsocks

func htons(val uint16) uint16 {
	return (val >> 8) | (val << 8)
}
