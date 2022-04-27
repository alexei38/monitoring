package monitor

const MaxErrors = 3

func AvgFloat(args ...float32) float32 {
	var sum float32
	for _, i := range args {
		sum += i
	}
	return sum / float32(len(args))
}
