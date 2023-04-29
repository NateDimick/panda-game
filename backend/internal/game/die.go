package game

import "math/rand"

type WeatherType string

const (
	SunWeather    WeatherType = "SUN"    // Sun allows the player to take a third action
	RainWeather   WeatherType = "RAIN"   // Rain allows the player to grow one bamboo shoot by 1 unit
	WindWeather   WeatherType = "WIND"   // Wind allows the player to take the same action twice (not required, just allowed)
	BoltWeather   WeatherType = "BOLT"   // Bolt allows the player to move the panda anywhere
	CloudWeather  WeatherType = "CLOUD"  // Cloud allows the player to take one available improvement. If no improvements are available, then this becomes a choice
	ChoiceWeather WeatherType = "CHOICE" // Choice allows the player to choose which of the 5 above weather conditions for their turn.
	NoWeather     WeatherType = "NONE"
)

// roll the weather die. The outcome depends on how many improvements are available.
func RollWeatherDie(improvements bool) WeatherType {
	var w [6]WeatherType
	if improvements {
		w = [6]WeatherType{SunWeather, RainWeather, WindWeather, BoltWeather, CloudWeather, ChoiceWeather}
	} else {
		w = [6]WeatherType{SunWeather, RainWeather, WindWeather, BoltWeather, ChoiceWeather, ChoiceWeather}
	}
	r := rand.Intn(6)

	return w[r]
}
