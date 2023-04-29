package game

import (
	"fmt"
	_ "unsafe"
)

var rollResult int = 0

//go:linkname randIntn math/rand.Intn
func randIntn(n int) int {
	fmt.Println(rollResult)
	return rollResult
}

// this test works in debug but not under regular run conditions
// randIntn isn't linked?
// func TestRollDie(t *testing.T) {
// 	cases := []struct {
// 		Name                 string
// 		ImprovementFlag      bool
// 		ExpectedWeatherOrder [6]WeatherType
// 	}{
// 		{"with improvements", true, [6]WeatherType{SunWeather, RainWeather, WindWeather, BoltWeather, CloudWeather, ChoiceWeather}},
// 		{"no improvements", false, [6]WeatherType{SunWeather, RainWeather, WindWeather, BoltWeather, ChoiceWeather, ChoiceWeather}},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.Name, func(tt *testing.T) {
// 			for i := 0; i < 6; i++ {
// 				rollResult = i
// 				w := RollWeatherDie(tc.ImprovementFlag)
// 				fmt.Println(tc.Name, tc.ExpectedWeatherOrder[i], w)
// 				assert.Equal(tt, tc.ExpectedWeatherOrder[i], w)
// 			}
// 		})
// 	}
// }
