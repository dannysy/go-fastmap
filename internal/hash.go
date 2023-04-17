package internal

func Hash(in string) (out uint64) {
	for i := range in {
		out += uint64(in[i])
	}
	return out
}
