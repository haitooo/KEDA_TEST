import express from "express";
import client from "prom-client";

const app = express();

// ===== Prometheus メトリクス設定 =====
const register = new client.Registry();

// HTTP リクエスト数カウンタ
const httpRequestsTotal = new client.Counter({
  name: "service1_http_requests_total",
  help: "Total HTTP requests received by service1",
  labelNames: ["method", "path", "status"],
});

register.registerMetric(httpRequestsTotal);

// Node.js の標準メトリクスもついでに
client.collectDefaultMetrics({ register });

// 各リクエストごとにカウンタをインクリメント
app.use((req, res, next) => {
  res.on("finish", () => {
    httpRequestsTotal
      .labels(req.method, req.path, String(res.statusCode))
      .inc();
  });
  next();
});

// Prometheus が scrape するエンドポイント
app.get("/metrics", async (_req, res) => {
  res.set("Content-Type", register.contentType);
  res.end(await register.metrics());
});

// ===== ここから元のサービス =====

// ヘルスチェック
app.get("/health", (_req, res) => {
  res.status(200).send("ok");
});

// トップページ
app.get("/", (_req, res) => {
  res.set("Content-Type", "text/html; charset=utf-8");
  res.send(`
    <h1>service1</h1>
    <p>これは最小のサンプルWebページです。</p>
  `);
});

// 3000番で全IF待受
const PORT = 3000;
app.listen(PORT, "0.0.0.0", () => {
  console.log(`[service1] listening on http://0.0.0.0:${PORT}`);
});
