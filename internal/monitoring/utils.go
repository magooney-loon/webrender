package monitoring

// calculatePercentiles takes a slice of latencies and returns percentile statistics
func calculatePercentiles(latencies []float64) map[string]float64 {
	if len(latencies) == 0 {
		return nil
	}

	quickSort(latencies, 0, len(latencies)-1)

	return map[string]float64{
		"p50": percentile(latencies, 50),
		"p75": percentile(latencies, 75),
		"p90": percentile(latencies, 90),
		"p95": percentile(latencies, 95),
		"p99": percentile(latencies, 99),
		"min": latencies[0],
		"max": latencies[len(latencies)-1],
	}
}

// percentile calculates the p-th percentile of sorted values
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	rank := (p / 100) * float64(len(sorted)-1)
	index := int(rank)
	if index+1 == len(sorted) {
		return sorted[index]
	}
	return sorted[index] + (rank-float64(index))*(sorted[index+1]-sorted[index])
}

// quickSort sorts the given array in-place
func quickSort(arr []float64, low, high int) {
	if low < high {
		pivot := partition(arr, low, high)
		quickSort(arr, low, pivot-1)
		quickSort(arr, pivot+1, high)
	}
}

// partition is a helper function for quickSort
func partition(arr []float64, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
