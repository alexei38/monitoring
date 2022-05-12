package monitor

func AvgFloat(args ...float32) float32 {
	var sum float32
	for _, i := range args {
		sum += i
	}
	return sum / float32(len(args))
}

func AvgInt64(args ...int64) int64 {
	var sum int64
	for _, i := range args {
		sum += i
	}
	return sum / int64(len(args))
}
