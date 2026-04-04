package outfit

import (
	"fmt"

	"github.com/y-maeda1116/tenki-fuku-bot/internal/weather"
)

type OutfitAdvice struct {
	Category string
	Outfit   string
	AllTips  []string
	TempMax  float64
	TempMin  float64
	TempDiff float64
}

var categoryLabels = map[string]string{
	"men":   "成人男性",
	"women": "成人女性",
	"kids":  "子供",
}

func selectOutfit(tempMax float64) string {
	switch {
	case tempMax < 15:
		return "厚手のアウター（コート、ダウン）"
	case tempMax < 20:
		return "薄手のジャケット、カーディガン"
	case tempMax < 25:
		return "長袖シャツ"
	default:
		return "半袖"
	}
}

func Advise(wd *weather.WeatherData, categories map[string]bool) []OutfitAdvice {
	var results []OutfitAdvice
	tempDiff := wd.TempMax - wd.TempMin

	for _, cat := range []string{"men", "women", "kids"} {
		if !categories[cat] {
			continue
		}

		outfit := selectOutfit(wd.TempMax)
		var tips []string

		if tempDiff >= 10 {
			tips = append(tips, "寒暖差が大きいです。脱ぎ着しやすい服装をおすすめします")
		}

		if cat == "kids" {
			tips = append(tips, "活動量を考慮して+1枚多めに着せるのがおすすめ")
		}

		results = append(results, OutfitAdvice{
			Category: cat,
			Outfit:   outfit,
			AllTips:  tips,
			TempMax:  wd.TempMax,
			TempMin:  wd.TempMin,
			TempDiff: tempDiff,
		})
	}

	return results
}

func TempColor(tempMax float64) int {
	switch {
	case tempMax < 15:
		return 0x3498DB
	case tempMax < 20:
		return 0x2ECC71
	case tempMax < 25:
		return 0xE67E22
	default:
		return 0xE74C3C
	}
}

func FormatAdvice(advice OutfitAdvice) string {
	label := categoryLabels[advice.Category]
	msg := fmt.Sprintf("**%sの服装アドバイス**\n👕 %s", label, advice.Outfit)
	for _, tip := range advice.AllTips {
		msg += fmt.Sprintf("\n💡 %s", tip)
	}
	msg += fmt.Sprintf("\n🌡️ 最高 %.1f℃ / 最低 %.1f℃（寒暖差 %.1f℃）", advice.TempMax, advice.TempMin, advice.TempDiff)
	return msg
}
