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
  areaCode: "130000",
  areaName: "東京地方",
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

  var wd = fetchTomorrowWeather(CONFIG.areaCode, CONFIG.areaName);
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

function fetchTomorrowWeather(areaCode, areaName) {
  var url = "https://www.jma.go.jp/bosai/forecast/data/forecast/" + areaCode + ".json";

  var response = UrlFetchApp.fetch(url, { muteHttpExceptions: true });
  var code = response.getResponseCode();

  if (code !== 200) {
    throw new Error("JMA forecast API returned status " + code + ": " + response.getContentText());
  }

  var data = JSON.parse(response.getContentText());
  var report = data[0];
  var timeSeries = report.timeSeries;

  var today = new Date();
  var tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  var tomorrowStr = Utilities.formatDate(tomorrow, "JST", "yyyy-MM-dd");

  var tempMax = -100;
  var tempMin = 100;
  var todayMax = -100;
  var todayMin = 100;
  var timeSlots = [];

  for (var s = 0; s < timeSeries.length; s++) {
    var series = timeSeries[s];
    var timeDefines = series.timeDefines;
    var areas = series.areas;

    for (var a = 0; a < areas.length; a++) {
      if (areas[a].area.name !== areaName) continue;

      if (series.areas[a].temps) {
        for (var t = 0; t < timeDefines.length; t++) {
          var dateStr = timeDefines[t].substring(0, 10);
          var temps = series.areas[a].temps;

          if (dateStr === tomorrowStr && temps.length > t) {
            var hi = parseFloat(temps[t]);
            var lo = parseFloat(temps[t]);
            if (temps.length > t + 1) {
              hi = Math.max(parseFloat(temps[t]), parseFloat(temps[t + 1]));
              lo = Math.min(parseFloat(temps[t]), parseFloat(temps[t + 1]));
            }
            if (hi > tempMax) tempMax = hi;
            if (lo < tempMin) tempMin = lo;
          }
          if (dateStr === Utilities.formatDate(today, "JST", "yyyy-MM-dd") && temps.length > t) {
            var thi = parseFloat(temps[t]);
            var tlo = parseFloat(temps[t]);
            if (temps.length > t + 1) {
              thi = Math.max(parseFloat(temps[t]), parseFloat(temps[t + 1]));
              tlo = Math.min(parseFloat(temps[t]), parseFloat(temps[t + 1]));
            }
            if (thi > todayMax) todayMax = thi;
            if (tlo < todayMin) todayMin = tlo;
          }
        }
      }

      if (series.areas[a].weathers) {
        var weathers = series.areas[a].weathers;
        var weatherTimes = timeDefines;
        for (var w = 0; w < weatherTimes.length; w++) {
          var wDate = weatherTimes[w].substring(0, 10);
          if (wDate !== tomorrowStr) continue;
          var slotLabel = "";
          if (w === 0) slotLabel = "朝〜昼";
          else if (w === 1) slotLabel = "昼〜夕";
          else if (w === 2) slotLabel = "夕〜夜";
          else slotLabel = "時間帯" + (w + 1);
          if (weathers.length > w && weathers[w]) {
            timeSlots.push({
              time: slotLabel,
              description: formatWeatherDesc(weathers[w]),
            });
          }
        }
      }
    }
  }

  if (tempMax === -100) {
    throw new Error("No forecast data available for tomorrow (" + tomorrowStr + ")");
  }

  return {
    city: areaName,
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
  var maxAttempts = 2;
  attempt = attempt || 1;

  var response = UrlFetchApp.fetch(url, {
    method: "post",
    contentType: "application/json",
    payload: payload,
    muteHttpExceptions: true,
  });

  var code = response.getResponseCode();

  if (code === 429 && attempt < maxAttempts) {
    var retryAfter = 5000;
    var headers = response.getHeaders();
    if (headers["Retry-After"]) {
      retryAfter = Math.min(parseInt(headers["Retry-After"], 10) * 1000, 10000);
    }
    Logger.log("Rate limited (429), retrying after " + retryAfter + "ms (attempt " + attempt + ")");
    Utilities.sleep(retryAfter);
    return postWithRetry(url, payload, attempt + 1);
  }

  return response;
}
