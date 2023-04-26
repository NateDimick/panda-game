package game

import "math/rand"

type Weather string

const (
	SunWeather    Weather = "SUN"    // Sun allows the player to take a third action
	RainWeather   Weather = "RAIN"   // Rain allows the player to grow one bamboo shoot by 1 unit
	WindWeather   Weather = "WIND"   // Wind allows the player to take the same action twice (not required, just allowed)
	BoltWeather   Weather = "BOLT"   // Bolt allows the player to move the panda anywhere
	CloudWeather  Weather = "CLOUD"  // Cloud allows the player to take one available improvement. If no improvements are available, then this becomes a choice
	ChoiceWeather Weather = "CHOICE" // Choice allows the player to choose which of the 5 above weather conditions for their turn.
)

// roll the weather die. The outcome depends on how many improvements are available.
func RollWeatherDie(improvements int) Weather {
	var w [6]Weather
	if improvements > 0 {
		w = [6]Weather{SunWeather, RainWeather, WindWeather, BoltWeather, CloudWeather, ChoiceWeather}
	} else {
		w = [6]Weather{SunWeather, RainWeather, WindWeather, BoltWeather, ChoiceWeather, ChoiceWeather}
	}
	r := rand.Intn(6)

	return w[r]
}
