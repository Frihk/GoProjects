package cars

import "math"

// CalculateWorkingCarsPerHour calculates how many working cars are
// produced by the assembly line every hour.
func CalculateWorkingCarsPerHour(productionRate int, successRate float64) float64 {
	// panic("CalculateWorkingCarsPerHour not implemented")
    result := float64(productionRate) * (successRate/100)
    return result
}

// CalculateWorkingCarsPerMinute calculates how many working cars are
// produced by the assembly line every minute.
func CalculateWorkingCarsPerMinute(productionRate int, successRate float64) int {
	// panic("CalculateWorkingCarsPerMinute not implemented")
    result := (float64(productionRate) * (successRate/100))/60
    return int(math.Floor(result))
}

// CalculateCost works out the cost of producing the given number of cars.
func CalculateCost(carsCount int) uint {
	// panic("CalculateCost not implemented")
    rem := carsCount%10
    suc := carsCount - rem
    cost := (suc*9500) + (rem*10000)
    return uint(cost)
}
