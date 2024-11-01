package svg

const (
	characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	size       = uint64(len(characters))
)

func encodeIndex(value uint64) string {
	var res [16]byte
	var i = len(res) - 1
	for {
		res[i] = characters[value%size]
		i--

		value /= size
		if value == 0 {
			break
		}
	}
	return string(res[i+1:])
}
