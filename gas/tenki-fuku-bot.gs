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
 *   2. スクリプトのプロパティに DISCORD_WEBHOOK_URL を設定
 *      （プロジェクトの設定 > スクリプトプロパティ）
 *   3. トリガーを設定: 毎日 20:00〜21:00 に実行
 *   気象データは気象庁API（api.jma.go.jp）を使用。APIキー不要。
 */

var CONFIG = {
  areaCode: "130000",
  weatherArea: "東京地方",
  tempArea: "東京",
  categories: {
    men: true,
    women: true,
    kids: true,
  },
};

// --- Main ---

function main() {
  var webhookUrl = PropertiesService.getScriptProperties().getProperty("DISCORD_WEBHOOK_URL");

  if (!webhookUrl) throw new Error("DISCORD_WEBHOOK_URL is not set in script properties");

  var wd = fetchTomorrowWeather(CONFIG.areaCode, CONFIG.weatherArea, CONFIG.tempArea);
  var advices = generateAdvice(wd, CONFIG.categories);

  if (advices.length === 0) {
    Logger.log("No categories enabled, skipping notification");
    return;
  }

  sendDiscordNotification(webhookUrl, advices, wd);
  Logger.log("Notification sent for " + wd.city + " tomorrow " + wd.date);
}

// --- Weather ---

var WEATHER_EMOJI = {
  "晴れ": "☀️",
  "晴時々曇": "⛅",
  "曇時々晴": "⛅",
  "曇り": "☁️",
  "薄い曇": "⛅",
  "雨": "🌧️",
  "小雨": "🌧️",
  "強い雨": "🌧️",
  "雷雨": "⛈️",
  "雪": "❄️",
  "小雪": "🌨️",
  "大雪": "❄️",
  "霧": "🌫️",
  "暴風雨": "🌪️",
  "暴風雪": "🌪️",
};

function formatWeatherDesc(desc) {
  var emoji = WEATHER_EMOJI[desc];
  return emoji ? emoji + " " + desc : desc;
}

function fetchTomorrowWeather(areaCode, weatherArea, tempArea) {
  var url = "https://www.jma.go.jp/bosai/forecast/data/forecast/" + areaCode + ".json";

  var response = UrlFetchApp.fetch(url, { muteHttpExceptions: true });
  var code = response.getResponseCode();

  if (code !== 200) {
    throw new Error("JMA forecast API returned status " + code + ": " + response.getContentText());
  }

  var data = JSON.parse(response.getContentText());

  var today = new Date();
  var tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  var todayStr = Utilities.formatDate(today, "JST", "yyyy-MM-dd");
  var tomorrowStr = Utilities.formatDate(tomorrow, "JST", "yyyy-MM-dd");

  // data[0]: short-term forecast (today + tomorrow)
  var shortTerm = data[0].timeSeries;

  // data[1]: weekly forecast
  var weekly = data[1].timeSeries;

  // --- Tomorrow's weather (timeSeries[0], area: weatherArea) ---
  var timeSlots = [];
  var ts0 = shortTerm[0];
  for (var a = 0; a < ts0.areas.length; a++) {
    if (ts0.areas[a].area.name !== weatherArea) continue;
    var weathers = ts0.areas[a].weathers;
    for (var w = 0; w < ts0.timeDefines.length; w++) {
      var wDate = ts0.timeDefines[w].substring(0, 10);
      if (wDate !== tomorrowStr) continue;
      var slotLabel = "";
      if (w === 0) slotLabel = "朝〜昼";
      else if (w === 1) slotLabel = "昼〜夕";
      else if (w === 2) slotLabel = "夕〜夜";
      else slotLabel = "時間帯" + (w + 1);
      if (weathers.length > w && weathers[w]) {
        timeSlots.push({ time: slotLabel, description: formatWeatherDesc(weathers[w]) });
      }
    }
    break;
  }

  // --- Tomorrow's temps (timeSeries[2], area: tempArea) ---
  // temps: [最低気温, 最高気温]
  var tempMax = -100;
  var tempMin = 100;
  var ts2 = shortTerm[2];
  for (var a2 = 0; a2 < ts2.areas.length; a2++) {
    if (ts2.areas[a2].area.name !== tempArea) continue;
    var temps = ts2.areas[a2].temps;
    if (temps.length >= 2) {
      tempMin = parseFloat(temps[0]);
      tempMax = parseFloat(temps[1]);
    }
    break;
  }

  // --- Today's temps for diff (weekly data timeSeries[1], area: tempArea) ---
  var todayMax = -100;
  var todayMin = 100;
  var wTs1 = weekly[1];
  for (var wa = 0; wa < wTs1.areas.length; wa++) {
    if (wTs1.areas[wa].area.name !== tempArea) continue;
    var mins = wTs1.areas[wa].tempsMin || [];
    var maxs = wTs1.areas[wa].tempsMax || [];
    for (var wt = 0; wt < wTs1.timeDefines.length; wt++) {
      if (wTs1.timeDefines[wt].substring(0, 10) === todayStr) {
        if (maxs.length > wt && maxs[wt] !== "") todayMax = parseFloat(maxs[wt]);
        if (mins.length > wt && mins[wt] !== "") todayMin = parseFloat(mins[wt]);
      }
    }
    break;
  }

  if (tempMax === -100) {
    throw new Error("No forecast data available for tomorrow (" + tomorrowStr + ")");
  }

  return {
    city: weatherArea,
    tempMax: tempMax,
    tempMin: tempMin,
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
      value: slot.description,
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
  var response = postWithRetry(webhookUrl, payload);

  var code = response.getResponseCode();
  if (code !== 204 && code !== 200) {
    Logger.log("Discord webhook failed with status " + code + ": " + response.getContentText());
    return;
  }
  Logger.log("Notification sent successfully");
}

function postWithRetry(url, payload, attempt) {
  var maxAttempts = 4;
  var delays = [5000, 15000, 30000];
  attempt = attempt || 1;

  var response = UrlFetchApp.fetch(url, {
    method: "post",
    contentType: "application/json",
    payload: payload,
    muteHttpExceptions: true,
  });

  var code = response.getResponseCode();

  if (code === 429 && attempt < maxAttempts) {
    var retryAfter = delays[attempt - 1] || 30000;
    var headers = response.getHeaders();
    if (headers["Retry-After"]) {
      retryAfter = Math.min(parseInt(headers["Retry-After"], 10) * 1000, retryAfter);
    }
    Logger.log("Rate limited (429), retrying after " + retryAfter + "ms (attempt " + attempt + "/" + maxAttempts + ")");
    Utilities.sleep(retryAfter);
    return postWithRetry(url, payload, attempt + 1);
  }

  return response;
}
