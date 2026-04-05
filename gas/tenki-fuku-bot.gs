/**
 * Tenki Fuku Bot - Google Apps Script
 *
 * 前日20時に翌日の天気予報を取得し、カテゴリー別の服装アドバイスを
 * Discord Webhookで通知するツール。
 *
 * 通知フォーマット:
 *   Embed 1: 天気（朝/昼/夕の天気＋気温）
 *   Embed 2〜: 服装（カテゴリー別、気温なし）
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

  var wd = fetchTomorrowWeather(CONFIG.city, apiKey);
  var advices = generateAdvice(wd, CONFIG.categories);

  if (advices.length === 0) {
    Logger.log("No categories enabled, skipping notification");
    return;
  }

  sendDiscordNotification(webhookUrl, advices, wd);
  Logger.log("Notification sent for " + wd.city + " tomorrow " + wd.date);
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

  var today = new Date();
  var tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  var todayStr = Utilities.formatDate(today, "JST", "yyyy-MM-dd");
  var tomorrowStr = Utilities.formatDate(tomorrow, "JST", "yyyy-MM-dd");

  var maxTemp = -100;
  var minTemp = 100;
  var todayMax = -100;
  var todayMin = 100;
  var descCount = {};
  var topDesc = "";
  var topCount = 0;

  var targetTimes = ["06:00:00", "12:00:00", "15:00:00"];
  var timeLabels = { "06:00:00": "朝 (7時)", "12:00:00": "昼 (12時)", "15:00:00": "夕 (17時)" };
  var timeSlots = [];

  for (var i = 0; i < data.list.length; i++) {
    var item = data.list[i];
    var dtTxt = item.dt_txt;
    var datePart = dtTxt.substring(0, 10);

    if (datePart === todayStr) {
      if (item.main.temp_max > todayMax) todayMax = item.main.temp_max;
      if (item.main.temp_min < todayMin) todayMin = item.main.temp_min;
    }

    if (datePart !== tomorrowStr) continue;

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

    var timePart = dtTxt.substring(11);
    for (var j = 0; j < targetTimes.length; j++) {
      if (timePart === targetTimes[j]) {
        var slotDesc = item.weather.length > 0 ? item.weather[0].description : "";
        timeSlots.push({
          time: timeLabels[timePart],
          description: slotDesc,
          temp: item.main.temp,
        });
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
    timeSlots: timeSlots,
    todayMax: todayMax,
    todayMin: todayMin,
  };
}

// --- Outfit ---

function selectOutfit(tempMax) {
  if (tempMax < 15) return "厚手のアウター（コート、ダウン）";
  if (tempMax < 20) return "薄手ジャケット、カーディガン";
  if (tempMax < 25) return "長袖シャツ";
  return "半袖";
}

function generateAdvice(wd, categories) {
  var results = [];
  var tempDiff = wd.tempMax - wd.tempMin;
  var order = ["men", "women", "kids"];

  for (var i = 0; i < order.length; i++) {
    var cat = order[i];
    if (!categories[cat]) continue;

    var outfit = selectOutfit(wd.tempMax);
    var tips = [];

    if (tempDiff >= 10) {
      tips.push("寒暖差が大きいです。脱ぎ着しやすい服装を");
    }

    if (cat === "kids") {
      tips.push("活動量を考慮して+1枚多めに");
    }

    results.push({
      category: cat,
      outfit: outfit,
      tips: tips,
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
  men: "男性",
  women: "女性",
  kids: "子ども",
};

function tempColor(tempMax) {
  if (tempMax < 15) return 0x3498DB;
  if (tempMax < 20) return 0x2ECC71;
  if (tempMax < 25) return 0xE67E22;
  return 0xE74C3C;
}

function buildWeatherEmbed(wd) {
  var fields = [];

  for (var i = 0; i < wd.timeSlots.length; i++) {
    var slot = wd.timeSlots[i];
    fields.push({
      name: slot.time,
      value: slot.description + "  " + slot.temp.toFixed(1) + "\u2103",
      inline: true,
    });
  }

  fields.push({ name: "最高", value: wd.tempMax.toFixed(1) + "\u2103", inline: true });
  fields.push({ name: "最低", value: wd.tempMin.toFixed(1) + "\u2103", inline: true });
  fields.push({ name: "寒暖差", value: (wd.tempMax - wd.tempMin).toFixed(1) + "\u2103", inline: true });

  if (wd.todayMax > -100) {
    var diffMax = wd.tempMax - wd.todayMax;
    var diffMin = wd.tempMin - wd.todayMin;
    fields.push({ name: "前日比", value: "最高 " + (diffMax >= 0 ? "+" : "") + diffMax.toFixed(1) + "\u2103 / 最低 " + (diffMin >= 0 ? "+" : "") + diffMin.toFixed(1) + "\u2103", inline: false });
  }

  return {
    title: "\u{1F324} 明日の天気（" + wd.city + "）",
    color: tempColor(wd.tempMax),
    fields: fields,
  };
}

function buildOutfitEmbed(advice, tempMax) {
  var fields = [{ name: "服装", value: advice.outfit, inline: false }];

  for (var i = 0; i < advice.tips.length; i++) {
    fields.push({ name: "アドバイス", value: advice.tips[i], inline: false });
  }

  return {
    title: CATEGORY_EMOJI[advice.category] + " " + CATEGORY_LABEL[advice.category],
    color: tempColor(tempMax),
    fields: fields,
  };
}

function sendDiscordNotification(webhookUrl, advices, wd) {
  var embeds = [buildWeatherEmbed(wd)];

  for (var i = 0; i < advices.length; i++) {
    embeds.push(buildOutfitEmbed(advices[i], wd.tempMax));
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
