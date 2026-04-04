/**
 * Tenki Fuku Bot - Google Apps Script
 *
 * 前日20時に翌日の天気予報を取得し、カテゴリー別の服装アドバイスを
 * Discord Webhookで通知するツール。
 *
 * 設定:
 *   1. 下方の CONFIG オブジェクトを編集
 *   2. スクリプトのプロパティに WEATHER_API_KEY, DISCORD_WEBHOOK_URL を設定
 *      （プロジェクトの設定 > スクリプトプロパティ）
 *   3. トリガーを設定: 毎日 20:00〜21:00 に実行
 */

var CONFIG = {
  city: "Tokyo",
  categories: {
    men: true,
    women: true,
    kids: true,
  },
};

// --- Main ---

function main() {
  var apiKey = PropertiesService.getScriptProperties().getProperty("WEATHER_API_KEY");
  var webhookUrl = PropertiesService.getScriptProperties().getProperty("DISCORD_WEBHOOK_URL");

  if (!apiKey) throw new Error("WEATHER_API_KEY is not set in script properties");
  if (!webhookUrl) throw new Error("DISCORD_WEBHOOK_URL is not set in script properties");

  var weather = fetchTomorrowWeather(CONFIG.city, apiKey);
  var advices = generateAdvice(weather, CONFIG.categories);

  if (advices.length === 0) {
    Logger.log("No categories enabled, skipping notification");
    return;
  }

  sendDiscordNotification(webhookUrl, advices, weather);
  Logger.log("Notification sent for " + weather.city + " tomorrow " + weather.date + " (" + weather.tempMax + "/" + weather.tempMin + ")");
}

// --- Weather ---

function fetchTomorrowWeather(city, apiKey) {
  var url = "https://api.openweathermap.org/data/2.5/forecast?q=" + encodeURIComponent(city) +
    "&appid=" + encodeURIComponent(apiKey) + "&units=metric";

  var response = UrlFetchApp.fetch(url, { muteHttpExceptions: true });
  var code = response.getResponseCode();

  if (code !== 200) {
    throw new Error("Forecast API returned status " + code + ": " + response.getContentText());
  }

  var data = JSON.parse(response.getContentText());
  var cityName = data.city.name;

  var tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  var tomorrowStr = Utilities.formatDate(tomorrow, "JST", "yyyy-MM-dd");

  var maxTemp = -100;
  var minTemp = 100;
  var descCount = {};
  var topDesc = "";
  var topCount = 0;

  for (var i = 0; i < data.list.length; i++) {
    var item = data.list[i];
    var dtTxt = item.dt_txt;
    if (dtTxt.substring(0, 10) !== tomorrowStr) continue;

    if (item.main.temp_max > maxTemp) maxTemp = item.main.temp_max;
    if (item.main.temp_min < minTemp) minTemp = item.main.temp_min;

    if (item.weather.length > 0) {
      var desc = item.weather[0].description;
      descCount[desc] = (descCount[desc] || 0) + 1;
      if (descCount[desc] > topCount) {
        topCount = descCount[desc];
        topDesc = desc;
      }
    }
  }

  if (maxTemp === -100) {
    throw new Error("No forecast data available for tomorrow (" + tomorrowStr + ")");
  }

  return {
    city: cityName,
    tempMax: maxTemp,
    tempMin: minTemp,
    description: topDesc,
    date: tomorrowStr,
  };
}

// --- Outfit ---

function selectOutfit(tempMax) {
  if (tempMax < 15) return "厚手のアウター（コート、ダウン）";
  if (tempMax < 20) return "薄手のジャケット、カーディガン";
  if (tempMax < 25) return "長袖シャツ";
  return "半袖";
}

function generateAdvice(weather, categories) {
  var results = [];
  var tempDiff = weather.tempMax - weather.tempMin;
  var order = ["men", "women", "kids"];

  for (var i = 0; i < order.length; i++) {
    var cat = order[i];
    if (!categories[cat]) continue;

    var outfit = selectOutfit(weather.tempMax);
    var tips = [];

    if (tempDiff >= 10) {
      tips.push("寒暖差が大きいです。脱ぎ着しやすい服装をおすすめします");
    }

    if (cat === "kids") {
      tips.push("活動量を考慮して+1枚多めに着せるのがおすすめ");
    }

    results.push({
      category: cat,
      outfit: outfit,
      tips: tips,
      tempMax: weather.tempMax,
      tempMin: weather.tempMin,
      tempDiff: tempDiff,
    });
  }

  return results;
}

// --- Discord ---

var CATEGORY_EMOJI = {
  men: "\u{1F454}",
  women: "\u{1F457}",
  kids: "\u{1F9F8}",
};

var CATEGORY_LABEL = {
  men: "成人男性",
  women: "成人女性",
  kids: "子供",
};

function tempColor(tempMax) {
  if (tempMax < 15) return 0x3498DB;
  if (tempMax < 20) return 0x2ECC71;
  if (tempMax < 25) return 0xE67E22;
  return 0xE74C3C;
}

function buildEmbed(advice) {
  var fields = [
    { name: "服装", value: advice.outfit, inline: false },
    { name: "最高気温", value: advice.tempMax.toFixed(1) + "\u2103", inline: true },
    { name: "最低気温", value: advice.tempMin.toFixed(1) + "\u2103", inline: true },
    { name: "寒暖差", value: advice.tempDiff.toFixed(1) + "\u2103", inline: true },
  ];

  for (var i = 0; i < advice.tips.length; i++) {
    fields.push({ name: "アドバイス", value: advice.tips[i], inline: false });
  }

  return {
    title: CATEGORY_EMOJI[advice.category] + " 明日の" + CATEGORY_LABEL[advice.category] + "の服装アドバイス",
    color: tempColor(advice.tempMax),
    fields: fields,
  };
}

function sendDiscordNotification(webhookUrl, advices, weather) {
  var embeds = [];
  for (var i = 0; i < advices.length; i++) {
    embeds.push(buildEmbed(advices[i]));
  }

  var payload = JSON.stringify({ embeds: embeds });

  var response = UrlFetchApp.fetch(webhookUrl, {
    method: "post",
    contentType: "application/json",
    payload: payload,
    muteHttpExceptions: true,
  });

  var code = response.getResponseCode();
  if (code !== 204 && code !== 200) {
    throw new Error("Discord webhook returned status " + code + ": " + response.getContentText());
  }
}
