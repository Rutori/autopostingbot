package algo

func Algo(c []string) int64 {
	k := float64(len(c)) / float64(24)
	if k == 0 {
		return 0
	} else if k > 0 && k <= 1 {
		return 1
	} else if k > 1 && k <= 2 {
		return 2
	}

	return int64(k)

}
