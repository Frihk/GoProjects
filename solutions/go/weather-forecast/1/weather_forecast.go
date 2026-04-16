// Package weather has all the logic need for the weather forcasting.
package weather

var (
	// CurrentCondition represents the current weather conditions.
	CurrentCondition string
	// CurrentLocation represents the position you are currently situated in.
	CurrentLocation  string
)
// Forecast take in city, which is the current position, and condtion, showing the weather condition. It returns a string of the weather condtion of the city you currently in.
func Forecast(city, condition string) string {
	CurrentLocation, CurrentCondition = city, condition
	return CurrentLocation + " - current weather condition: " + CurrentCondition
}
