"""Hot RSS collector worker.

This keeps the status file and PostgreSQL writes compatible with the retired
Node collector so the Go API and admin worker page can keep reading the same
runtime contract.
"""

from __future__ import annotations

import argparse
import concurrent.futures
import hashlib
import html
import json
import os
import re
import sys
import threading
import time
import urllib.error
import urllib.request
import uuid
import xml.etree.ElementTree as ET
from dataclasses import dataclass
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

import psycopg
from psycopg.rows import dict_row
from psycopg.types.json import Json

try:
    import feedparser
except Exception:  # pragma: no cover - optional runtime dependency fallback
    feedparser = None

HOT_RSS_ADVISORY_LOCK_KEY = 874_270_153

NON_AI_KEYWORDS = [
    "护发",
    "美妆",
    "护肤",
    "彩妆",
    "美容",
    "美甲",
    "美发",
    "汽车大灯",
    "车灯",
    "轮胎",
    "发动机",
    "变速箱",
    "美食",
    "菜谱",
    "烹饪",
    "餐厅",
    "外卖",
    "旅游",
    "酒店",
    "机票",
    "签证",
    "景点",
    "游戏攻略",
    "游戏评测",
    "手游",
    "网游",
    "电竞",
    "健康养生",
    "药品",
    "保健品",
    "中医",
    "房产",
    "装修",
    "家具",
    "家电",
    "股票",
    "基金",
    "理财",
    "保险",
    "体育",
    "足球",
    "篮球",
    "网球",
    "奥运",
    "娱乐",
    "明星",
    "八卦",
    "综艺",
    "电影",
    "时尚",
    "服装",
    "穿搭",
    "潮流",
    "感染",
    "肠胃",
    "胃肠",
    "蛋白质",
    "肿瘤",
    "癌症",
    "细胞治疗",
    "疫苗",
    "约会",
    "Tinder",
    "Airbnb",
    "诗歌相机",
    "Poetry Camera",
    "鞋公司",
    "Allbirds",
    "OkCupid",
    "A股",
    "B股",
    "港股",
    "美股",
    "股市",
    "涨跌",
    "涨幅",
    "跌幅",
    "沪指",
    "深成指",
    "创业板指",
    "三大指数",
    "集体收涨",
    "集体收跌",
    "中国移动",
    "中国联通",
    "中国电信",
    "工商银行",
    "建设银行",
    "兖矿能源",
    "岳阳兴长",
    "英维克",
    "寒武纪",
    "翠微股份",
    "理想汽车",
    "蔚来",
    "小鹏",
    "比亚迪",
    "特斯拉",
    "华为汽车",
    "问界",
    "极氪",
    "岚图",
    "零跑",
    "哪吒",
]

AI_KEYWORDS_ZH = [
    "AI",
    "人工智能",
    "大模型",
    "LLM",
    "GPT",
    "ChatGPT",
    "Claude",
    "Gemini",
    "Llama",
    "深度学习",
    "机器学习",
    "神经网络",
    "自然语言",
    "NLP",
    "计算机视觉",
    "CV",
    "生成式",
    "AIGC",
    "AGI",
    "智能体",
    "Agent",
    "Transformer",
    "扩散模型",
    "Stable Diffusion",
    "Midjourney",
    "DALL-E",
    "Sora",
    "可灵",
    "通义",
    "文心",
    "豆包",
    "Kimi",
    "DeepSeek",
    "智谱",
    "百川",
    "MiniMax",
    "零一万物",
    "机器人",
    "具身智能",
    "自动驾驶",
    "语音识别",
    "TTS",
    "ASR",
    "英伟达",
    "NVIDIA",
    "GPU",
    "算力",
    "芯片",
    "H100",
    "A100",
    "OpenAI",
    "Anthropic",
    "Google AI",
    "Meta AI",
    "微软 AI",
    "百度 AI",
    "阿里 AI",
    "腾讯 AI",
    "字节 AI",
    "华为 AI",
    "科大讯飞",
    "商汤",
    "旷视",
]

AI_KEYWORDS_EN = [
    "AI",
    "artificial intelligence",
    "machine learning",
    "deep learning",
    "LLM",
    "GPT",
    "ChatGPT",
    "Claude",
    "Gemini",
    "Llama",
    "transformer",
    "diffusion",
    "generative",
    "AIGC",
    "AGI",
    "agent",
    "neural network",
    "NLP",
    "OpenAI",
    "Anthropic",
    "Google AI",
    "Meta AI",
    "Microsoft AI",
    "NVIDIA",
    "GPU",
    "inference",
    "training",
    "fine-tuning",
    "RLHF",
    "Stable Diffusion",
    "Midjourney",
    "DALL-E",
    "Sora",
    "Whisper",
]

SOURCE_WEIGHTS = {
    "机器之心": 10,
    "36氪": 8,
    "TechCrunch AI": 10,
    "The Verge AI": 8,
    "量子位": 8,
    "InfoQ AI": 7,
    "MIT Tech Review": 8,
    "Ars Technica AI": 6,
    "VentureBeat AI": 7,
    "IT之家": 5,
}

NOISY_NEWS_SOURCES = {"Hacker News AI", "r/LocalLLaMA", "r/MachineLearning", "r/artificial", "V2EX"}
BAD_DESC_PATTERNS = (
    "点击查看原文",
    "文章网址：",
    "评论网址：",
    "reddit.com/",
    "v2ex.com/",
    "![图片",
    "```",
    "i hope they include it",
    "deepseek v4 人",
    "this is today's edition of the download",
    "introducing: the nature issue",
)
BREAKING_MODEL_KEYWORDS = {
    "gpt-5.5": 18,
    "deepseek v4": 18,
    "deepseek-v4": 18,
    "deepseek: deepseek v4": 18,
}


def _env(key: str, fallback: str) -> str:
    value = os.getenv(key)
    return fallback if value is None or value == "" else value


def _int_env(key: str, fallback: int) -> int:
    try:
        value = int(os.getenv(key, ""))
    except ValueError:
        return fallback
    return value if value > 0 else fallback


def _bool_env(key: str, fallback: bool) -> bool:
    value = os.getenv(key)
    if value is None or value == "":
        return fallback
    return value.strip().lower() in {"1", "true", "yes", "on"}


def _now() -> datetime:
    return datetime.now(UTC)


def _iso(value: datetime | None = None) -> str:
    return (value or _now()).isoformat().replace("+00:00", "Z")


@dataclass(frozen=True)
class Config:
    database_url: str
    status_path: Path
    interval_ms: int
    max_backoff_ms: int
    collect_on_start: bool
    max_runs: int
    once: bool
    limit_per_source: int
    source_concurrency: int
    request_timeout_ms: int
    health_max_age_ms: int
    ai_enrich_enabled: bool
    ai_api_key: str
    ai_base_url: str
    ai_model: str
    ai_timeout_ms: int
    ai_max_items_per_source: int
    ai_max_items_per_run: int
    clean_enabled: bool
    clean_recent_days: int
    clean_max_per_source: int

    @classmethod
    def from_env(cls, *, once: bool = False) -> "Config":
        ai_api_key = _env("HOT_AI_API_KEY", _env("OPENAI_API_KEY", ""))
        return cls(
            database_url=_env("DATABASE_URL", ""),
            status_path=Path(_env("HOT_RSS_WORKER_STATUS_PATH", _env("HOT_WORKER_STATUS_PATH", "/app/runtime/hot-worker-status.json"))),
            interval_ms=_int_env("HOT_RSS_COLLECT_INTERVAL_MS", 30 * 60 * 1000),
            max_backoff_ms=_int_env("HOT_RSS_COLLECT_MAX_BACKOFF_MS", 10 * 60 * 1000),
            collect_on_start=_bool_env("HOT_RSS_COLLECT_ON_START", True),
            max_runs=_int_env("HOT_RSS_COLLECT_MAX_RUNS", 0),
            once=once or _bool_env("HOT_RSS_COLLECT_ONCE", False),
            limit_per_source=_int_env("HOT_RSS_LIMIT_PER_SOURCE", 10),
            source_concurrency=_int_env("HOT_RSS_SOURCE_CONCURRENCY", 6),
            request_timeout_ms=_int_env("HOT_RSS_REQUEST_TIMEOUT_MS", 20 * 1000),
            health_max_age_ms=_int_env("HOT_RSS_WORKER_HEALTH_MAX_AGE_MS", 0),
            ai_enrich_enabled=_bool_env("HOT_AI_ENRICH_ENABLED", True),
            ai_api_key=ai_api_key,
            ai_base_url=_env("HOT_AI_BASE_URL", _env("OPENAI_BASE_URL", "https://api.openai.com/v1")),
            ai_model=_env("HOT_AI_MODEL", "gpt-4o-mini"),
            ai_timeout_ms=_int_env("HOT_AI_TIMEOUT_MS", 30 * 1000),
            ai_max_items_per_source=_int_env("HOT_AI_MAX_ITEMS_PER_SOURCE", 5),
            ai_max_items_per_run=_int_env("HOT_AI_MAX_ITEMS_PER_RUN", 20),
            clean_enabled=_bool_env("HOT_RANK_CLEAN_ENABLED", True),
            clean_recent_days=_int_env("HOT_RANK_CLEAN_RECENT_DAYS", 7),
            clean_max_per_source=_int_env("HOT_RANK_CLEAN_MAX_PER_SOURCE", 2),
        )

    def validate(self) -> None:
        if not self.database_url:
            raise RuntimeError("DATABASE_URL is required")

    def status_max_age_ms(self) -> int:
        if self.health_max_age_ms > 0:
            return self.health_max_age_ms
        return max(self.interval_ms * 2, self.interval_ms + self.max_backoff_ms)

    def ai_enrich_active(self) -> bool:
        return bool(self.ai_enrich_enabled and self.ai_api_key and self.ai_base_url and self.ai_model and self.ai_max_items_per_source > 0 and self.ai_max_items_per_run > 0)


class AIBudget:
    def __init__(self, limit: int):
        self._remaining = max(0, int(limit))
        self._lock = threading.Lock()

    def acquire(self) -> bool:
        with self._lock:
            if self._remaining <= 0:
                return False
            self._remaining -= 1
            return True

    def remaining(self) -> int:
        with self._lock:
            return self._remaining


def read_status(config: Config) -> dict[str, Any]:
    if not config.status_path.exists():
        return {}
    try:
        return json.loads(config.status_path.read_text(encoding="utf-8"))
    except Exception:
        return {}


def write_status(config: Config, status: str, **extra: Any) -> None:
    payload = {
        "status": status,
        "apply": True,
        "storage": "postgres",
        "mode": "once" if config.once else "loop",
        "collect_on_start": config.collect_on_start,
        "interval_ms": config.interval_ms,
        "max_backoff_ms": config.max_backoff_ms,
        "max_runs": config.max_runs,
        "limit_per_source": config.limit_per_source,
        "source_concurrency": config.source_concurrency,
        "ai_enrich_enabled": config.ai_enrich_active(),
        "ai_model": config.ai_model if config.ai_enrich_active() else "",
        "ai_max_items_per_run": config.ai_max_items_per_run,
        "rank_clean_enabled": config.clean_enabled,
        "updated_at": _iso(),
        **extra,
    }
    config.status_path.parent.mkdir(parents=True, exist_ok=True)
    config.status_path.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def run_healthcheck(config: Config) -> None:
    config.validate()
    if not config.status_path.exists():
        raise RuntimeError(f"HOT_RSS_WORKER_STATUS_PATH not found: {config.status_path}")
    status = json.loads(config.status_path.read_text(encoding="utf-8"))
    if status.get("skipped_reason") == "advisory_lock_held":
        raise RuntimeError("hot rss worker skipped because advisory lock is held")
    if status.get("status") not in {"ok", "running"}:
        raise RuntimeError(f"hot rss worker unhealthy status: {status.get('status') or 'unknown'}")
    updated_at = status.get("updated_at") or ""
    parsed = datetime.fromisoformat(str(updated_at).replace("Z", "+00:00"))
    age_ms = int((_now() - parsed).total_seconds() * 1000)
    if age_ms > config.status_max_age_ms():
        raise RuntimeError(f"hot rss worker status is stale: {age_ms}ms > {config.status_max_age_ms()}ms")
    print(f"[hot-rss-worker] healthcheck ok status={status.get('status')} age_ms={age_ms}")


def _local_name(tag: str) -> str:
    return tag.rsplit("}", 1)[-1].split(":", 1)[-1].lower()


def _text(value: str | None) -> str:
    return re.sub(r"\s+", " ", html.unescape(value or "")).strip()


def _strip_tags(value: str | None) -> str:
    return _text(re.sub(r"<[^>]+>", " ", value or ""))


def _children(elem: ET.Element, *names: str) -> list[ET.Element]:
    wanted = {name.lower() for name in names}
    return [child for child in list(elem) if _local_name(child.tag) in wanted]


def _first_text(elem: ET.Element, *names: str) -> str:
    for child in _children(elem, *names):
        raw = "".join(child.itertext())
        value = _strip_tags(raw)
        if value:
            return value
    return ""


def _link(elem: ET.Element) -> str:
    for child in _children(elem, "link"):
        href = child.attrib.get("href", "")
        if href:
            return _text(href)
        raw = "".join(child.itertext())
        if _text(raw):
            return _text(raw)
    return ""


def _categories(elem: ET.Element) -> list[str]:
    values: list[str] = []
    for child in _children(elem, "category"):
        term = _text(child.attrib.get("term", ""))
        if term:
            values.append(term)
        raw = _strip_tags("".join(child.itertext()))
        if raw:
            values.append(raw)
    deduped: list[str] = []
    for value in values:
        if value not in deduped:
            deduped.append(value)
    return deduped[:8]


def _normalize_date(value: str) -> str:
    if not value:
        return ""
    try:
        from email.utils import parsedate_to_datetime

        parsed = parsedate_to_datetime(value)
        if parsed.tzinfo is None:
            parsed = parsed.replace(tzinfo=UTC)
        return parsed.astimezone(UTC).isoformat().replace("+00:00", "Z")
    except Exception:
        pass
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00")).astimezone(UTC).isoformat().replace("+00:00", "Z")
    except Exception:
        return ""


def _source_config(row: dict[str, Any]) -> dict[str, Any]:
    value = row.get("config_json")
    if isinstance(value, dict):
        return value
    try:
        return json.loads(value or "{}")
    except Exception:
        return {}


def _seed_urls(row: dict[str, Any]) -> list[str]:
    value = row.get("seed_urls_json")
    if isinstance(value, list):
        return [str(item) for item in value if item]
    try:
        parsed = json.loads(value or "[]")
        return [str(item) for item in parsed if item] if isinstance(parsed, list) else []
    except Exception:
        return []


def _source_category_label(config: dict[str, Any]) -> str:
    label = str(config.get("category_label") or "").strip()
    if label:
        return label
    if config.get("category") == "news":
        return "AI 资讯"
    return "官方信源" if config.get("category") == "official" else "博客文章"


def _source_display_name(source: dict[str, Any]) -> str:
    return str(source.get("title") or source.get("source_id") or "").strip()


def _source_handle(source: dict[str, Any], feed_url: str) -> str:
    config = _source_config(source)
    configured = str(config.get("handle") or config.get("source_handle") or "").strip()
    if configured:
        return configured
    try:
        from urllib.parse import urlparse

        return (urlparse(str(source.get("base_url") or feed_url)).hostname or "").removeprefix("www.")
    except Exception:
        return str(source.get("source_id") or "").strip()


def _contains_keyword(text: str, keyword: str) -> bool:
    keyword = str(keyword or "").strip()
    if not keyword:
        return False
    if any("\u4e00" <= char <= "\u9fff" for char in keyword):
        return keyword.upper() in text.upper()
    if re.fullmatch(r"[A-Za-z0-9.+-]{1,4}", keyword):
        return re.search(rf"(?<![A-Za-z0-9]){re.escape(keyword)}(?![A-Za-z0-9])", text, flags=re.I) is not None
    return keyword.upper() in text.upper()


def is_ai_related(title: str, summary: str, source_config: dict[str, Any]) -> bool:
    if source_config.get("skip_ai_filter") is True:
        return True
    text = f"{title or ''} {summary or ''}"
    if any(_contains_keyword(text, keyword) for keyword in NON_AI_KEYWORDS):
        return False
    keywords = AI_KEYWORDS_ZH + AI_KEYWORDS_EN
    if source_config.get("ai_only", False):
        return any(_contains_keyword(text, keyword) for keyword in keywords)
    return any(_contains_keyword(text, keyword) for keyword in keywords)


def _zh_ratio(text: str) -> float:
    text = str(text or "").strip()
    if not text:
        return 0.0
    zh_count = sum("\u4e00" <= char <= "\u9fff" for char in text)
    letter_count = sum(char.isalpha() or ("\u4e00" <= char <= "\u9fff") for char in text)
    return zh_count / max(letter_count, 1)


def _is_noisy_source(source_name: str, source_config: dict[str, Any]) -> bool:
    if source_config.get("noisy") is True:
        return True
    return source_name in NOISY_NEWS_SOURCES


def is_publishable_hot_item(title: str, summary: str, source_name: str, source_config: dict[str, Any]) -> bool:
    if source_config.get("skip_publishable_filter") is True:
        return True
    if _is_noisy_source(source_name, source_config) and source_config.get("allow_noisy") is not True:
        return False
    blob = f"{title or ''} {summary or ''}".lower()
    if any(pattern in blob for pattern in BAD_DESC_PATTERNS):
        return False
    if str(title or "").lower().startswith("the download:"):
        return False
    min_zh_ratio = source_config.get("min_zh_ratio")
    if isinstance(min_zh_ratio, (int, float)) and min_zh_ratio > 0:
        if _zh_ratio(title) < float(min_zh_ratio) and _zh_ratio(summary) < float(min_zh_ratio):
            return False
    return True


def _parse_json_object(content: str) -> dict[str, Any]:
    cleaned = str(content or "").strip()
    fenced = re.fullmatch(r"```(?:json)?\s*(.*?)\s*```", cleaned, flags=re.DOTALL | re.IGNORECASE)
    if fenced:
        cleaned = fenced.group(1).strip()
    if not cleaned.startswith("{"):
        match = re.search(r"(\{.*\})", cleaned, flags=re.DOTALL)
        if match:
            cleaned = match.group(1).strip()
    parsed = json.loads(cleaned)
    if not isinstance(parsed, dict):
        raise RuntimeError("AI response must be a JSON object")
    return parsed


def _ai_chat(config: Config, *, system: str, user: str) -> str:
    url = config.ai_base_url.rstrip("/") + "/chat/completions"
    payload = {
        "model": config.ai_model,
        "messages": [
            {"role": "system", "content": system},
            {"role": "user", "content": user},
        ],
        "temperature": 0.2,
    }
    request = urllib.request.Request(
        url,
        data=json.dumps(payload, ensure_ascii=False).encode("utf-8"),
        headers={
            "content-type": "application/json",
            "authorization": f"Bearer {config.ai_api_key}",
        },
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=config.ai_timeout_ms / 1000) as response:
            body = json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")[:300]
        raise RuntimeError(f"AI HTTP {exc.code}: {detail}") from exc
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as exc:
        raise RuntimeError(f"AI request failed: {exc}") from exc
    try:
        return str(body["choices"][0]["message"]["content"]).strip()
    except (KeyError, IndexError, TypeError) as exc:
        raise RuntimeError("AI response missing choices[0].message.content") from exc


def ai_clean_item(config: Config, item: dict[str, Any]) -> dict[str, Any]:
    source_name = str(item.get("source_name") or item.get("source_id") or "")
    payload = {
        "title": item.get("title") or "",
        "summary": item.get("summary") or "",
        "url": item.get("canonical_url") or "",
        "source": source_name,
        "published_at": item.get("published_at") or "",
        "tags": item.get("tags_json") or [],
    }
    content = _ai_chat(
        config,
        system=(
            "You clean AI news feed items for a Chinese AI热点 product. "
            "Return only JSON with keys: publish(boolean), title_zh(string), ai_summary(string), "
            "reason(string), tags(array of short strings). publish=false for non-AI, spam, finance-only, "
            "car-only, entertainment-only, health/medical, travel, beauty, low-quality comments, or unreadable items. "
            "When publish=true, title_zh must be concise Chinese and ai_summary must be one Chinese sentence under 120 chars."
        ),
        user=json.dumps(payload, ensure_ascii=False),
    )
    parsed = _parse_json_object(content)
    return {
        "publish": bool(parsed.get("publish", True)),
        "title_zh": _text(str(parsed.get("title_zh") or ""))[:120],
        "ai_summary": _text(str(parsed.get("ai_summary") or ""))[:180],
        "reason": _text(str(parsed.get("reason") or ""))[:180],
        "tags": [str(tag).strip()[:30] for tag in parsed.get("tags") or [] if str(tag).strip()][:8],
    }


def enrich_items_with_ai(config: Config, items: list[dict[str, Any]], *, max_items: int | None = None, ai_budget: AIBudget | None = None) -> tuple[list[dict[str, Any]], dict[str, Any]]:
    if not config.ai_enrich_active() or not items:
        return items, {"enabled": False, "attempted_count": 0, "enriched_count": 0, "rejected_count": 0, "error_count": 0}
    enriched: list[dict[str, Any]] = []
    attempted = 0
    enriched_count = 0
    rejected_count = 0
    error_count = 0
    errors: list[str] = []
    max_items = min(len(items), config.ai_max_items_per_source, max_items if max_items is not None else config.ai_max_items_per_source)
    for index, item in enumerate(items):
        if index >= max_items:
            enriched.append(item)
            continue
        if ai_budget is not None and not ai_budget.acquire():
            enriched.append(item)
            continue
        attempted += 1
        original_title = item.get("title") or ""
        original_summary = item.get("summary") or ""
        try:
            result = ai_clean_item(config, item)
        except Exception as exc:
            error_count += 1
            errors.append(str(exc)[:180])
            metrics = dict(item.get("metrics_json") or {})
            metrics["ai_enrich_error"] = str(exc)[:300]
            metrics["ai_enrich_model"] = config.ai_model
            item["metrics_json"] = metrics
            enriched.append(item)
            continue
        metrics = dict(item.get("metrics_json") or {})
        metrics.update(
            {
                "ai_enriched": True,
                "ai_enrich_model": config.ai_model,
                "ai_enriched_at": _iso(),
                "ai_publish": result["publish"],
                "ai_reason": result["reason"],
                "title_original": original_title,
                "summary_original": original_summary,
            }
        )
        if result["title_zh"]:
            metrics["title_zh"] = result["title_zh"]
            item["title"] = result["title_zh"]
        if result["ai_summary"]:
            metrics["ai_summary"] = result["ai_summary"]
            item["summary"] = result["ai_summary"]
            item["body"] = result["ai_summary"]
        if result["tags"]:
            existing_tags = [str(tag) for tag in item.get("tags_json") or []]
            item["tags_json"] = list(dict.fromkeys(existing_tags + result["tags"]))[:8]
        item["metrics_json"] = metrics
        if not result["publish"]:
            rejected_count += 1
            continue
        enriched_count += 1
        enriched.append(item)
    return enriched, {
        "enabled": True,
        "model": config.ai_model,
        "attempted_count": attempted,
        "enriched_count": enriched_count,
        "rejected_count": rejected_count,
        "error_count": error_count,
        "budget_remaining": ai_budget.remaining() if ai_budget is not None else None,
        "errors": errors[:3],
    }


def _score_for_item(published_at: str, tags: list[str], title: str, source_config: dict[str, Any]) -> int:
    category = source_config.get("category") or "blog"
    source_name = str(source_config.get("source_name") or source_config.get("title") or "").strip()
    configured_weight = source_config.get("source_weight")
    if isinstance(configured_weight, (int, float)):
        source_weight = int(configured_weight)
    else:
        source_weight = SOURCE_WEIGHTS.get(source_name, 0)
    score = 70 if category == "official" else 60
    score += source_weight
    try:
        age_days = (_now() - datetime.fromisoformat(published_at.replace("Z", "+00:00"))).total_seconds() / 86400
    except Exception:
        age_days = 999
    if age_days <= 1:
        score += 25
    elif age_days <= 3:
        score += 15
    score += min(len(tags) * 2, 10)
    if re.search(r"\b(ai|agent|model|llm|copilot|image|video|search|reasoning)\b", title or "", flags=re.I):
        score += 8
    title_low = (title or "").lower()
    for keyword, bonus in BREAKING_MODEL_KEYWORDS.items():
        if keyword in title_low:
            score += bonus
    if source_config.get("noisy") is True:
        score -= 20
    return max(0, min(100, round(score)))


def fetch_text(config: Config, url: str) -> str:
    request = urllib.request.Request(
        url,
        headers={
            "accept": "application/rss+xml, application/atom+xml, application/xml, text/xml, */*",
            "user-agent": "cloudbase-hot-rss-collector/1.0",
        },
    )
    try:
        with urllib.request.urlopen(request, timeout=config.request_timeout_ms / 1000) as response:
            return response.read().decode(response.headers.get_content_charset() or "utf-8", errors="replace")
    except urllib.error.HTTPError as exc:
        raise RuntimeError(f"HTTP {exc.code}") from exc


def build_feed_item(
    source: dict[str, Any],
    feed_url: str,
    *,
    title: str,
    link: str,
    summary: str,
    published_at: str,
    author: str,
    guid: str,
    tags: list[str],
    parser: str,
) -> dict[str, Any] | None:
    source_config_value = _source_config(source)
    source_name = _source_display_name(source)
    source_config_value.setdefault("source_name", source_name)
    if not is_ai_related(title, summary, source_config_value):
        return None
    if not is_publishable_hot_item(title, summary, source_name, source_config_value):
        return None
    hash_input = "\n".join([str(source.get("source_id") or ""), guid, link, title, summary])
    content_hash = hashlib.sha256(hash_input.encode("utf-8")).hexdigest()
    external_id = f"{feed_url}::{source.get('source_id')}:{guid or content_hash}"
    hot_score = _score_for_item(published_at, tags, title, source_config_value)
    reason = (
        f"{_source_category_label(source_config_value)} · 当日热点"
        if published_at and (_now() - datetime.fromisoformat(published_at.replace("Z", "+00:00"))).total_seconds() <= 86400
        else f"{_source_category_label(source_config_value)} · RSS 采集"
    )
    metrics = {
        "reason": reason,
        "hot_score": hot_score,
        "source_category": source_config_value.get("category") or "blog",
        "source_category_label": _source_category_label(source_config_value),
        "ai_related": True,
        "source_weight": source_config_value.get("source_weight", SOURCE_WEIGHTS.get(source_name, 0)),
        "ai_hot_rules_version": 1,
        "parser": parser,
    }
    item = {
        "source_id": source.get("source_id"),
        "external_id": external_id,
        "canonical_url": link,
        "title": title,
        "summary": summary,
        "body": summary,
        "reason": reason,
        "published_at": published_at or None,
        "author": author,
        "source_name": source_name,
        "source_handle": _source_handle(source, feed_url),
        "badge": _source_category_label(source_config_value),
        "score": str(hot_score),
        "content_type": "article",
        "tags_json": tags,
        "metrics_json": metrics,
        "raw_ref_json": {"feed_url": feed_url, "guid": guid, "parser": parser},
        "content_hash": content_hash,
    }
    if item["title"] and item["external_id"]:
        return item
    return None


def _feedparser_tags(entry: Any) -> list[str]:
    values: list[str] = []
    for tag in getattr(entry, "tags", []) or []:
        if isinstance(tag, dict):
            value = tag.get("term") or tag.get("label")
        else:
            value = getattr(tag, "term", "") or getattr(tag, "label", "")
        value = _text(str(value or ""))
        if value and value not in values:
            values.append(value)
    return values[:8]


def _entry_get(entry: Any, key: str, default: Any = "") -> Any:
    if hasattr(entry, "get"):
        return entry.get(key, default)
    return getattr(entry, key, default)


def _entry_content_value(entry: Any) -> str:
    content = _entry_get(entry, "content", [])
    if isinstance(content, list) and content:
        first = content[0]
        if isinstance(first, dict):
            return str(first.get("value") or "")
        return str(getattr(first, "value", "") or "")
    if isinstance(content, dict):
        return str(content.get("value") or "")
    return ""


def parse_feed_with_feedparser(config: Config, xml: str, source: dict[str, Any], feed_url: str) -> list[dict[str, Any]]:
    if feedparser is None:
        return []
    parsed = feedparser.parse(xml)
    entries = list(getattr(parsed, "entries", []) or [])
    items: list[dict[str, Any]] = []
    for entry in entries[: config.limit_per_source]:
        try:
            title = _strip_tags(str(_entry_get(entry, "title", "")))
            link = _text(str(_entry_get(entry, "link", "")))
            summary = _strip_tags(str(_entry_get(entry, "summary", "") or _entry_get(entry, "description", "") or _entry_content_value(entry)))
            published_at = _normalize_date(str(_entry_get(entry, "published", "") or _entry_get(entry, "updated", "") or _entry_get(entry, "created", "")))
            author = _strip_tags(str(_entry_get(entry, "author", "")))
            guid = _text(str(_entry_get(entry, "id", "") or _entry_get(entry, "guid", "") or link or title))
        except Exception:
            continue
        item = build_feed_item(
            source,
            feed_url,
            title=title,
            link=link,
            summary=summary,
            published_at=published_at,
            author=author,
            guid=guid,
            tags=_feedparser_tags(entry),
            parser="feedparser",
        )
        if item:
            items.append(item)
    return items


def parse_feed_with_xml(config: Config, xml: str, source: dict[str, Any], feed_url: str) -> list[dict[str, Any]]:
    xml = re.sub(r"<(/?)content:encoded", r"<\1content", xml)
    root = ET.fromstring(xml.encode("utf-8"))
    entries = [elem for elem in root.iter() if _local_name(elem.tag) in {"item", "entry"}]
    items: list[dict[str, Any]] = []
    for entry in entries[: config.limit_per_source]:
        item = build_feed_item(
            source,
            feed_url,
            title=_first_text(entry, "title"),
            link=_link(entry),
            summary=_first_text(entry, "description", "summary", "content", "encoded"),
            published_at=_normalize_date(_first_text(entry, "pubDate", "published", "updated", "date")),
            author=_first_text(entry, "author", "creator", "name"),
            guid=_first_text(entry, "guid", "id") or _link(entry) or _first_text(entry, "title"),
            tags=_categories(entry),
            parser="xml",
        )
        if item:
            items.append(item)
    return items


def parse_feed(config: Config, xml: str, source: dict[str, Any], feed_url: str) -> list[dict[str, Any]]:
    try:
        items = parse_feed_with_feedparser(config, xml, source, feed_url)
        if items:
            return items
    except Exception:
        pass
    return parse_feed_with_xml(config, xml, source, feed_url)


def load_sources(conn: psycopg.Connection[Any]) -> list[dict[str, Any]]:
    with conn.cursor(row_factory=dict_row) as cur:
        cur.execute(
            """
            SELECT source_id, adapter_kind, title, description, enabled, base_url, seed_urls_json, config_json, created_at, updated_at
            FROM hot_sources
            WHERE enabled = TRUE AND adapter_kind = 'rss-generic'
            ORDER BY sort_order ASC, source_id
            """
        )
        return [dict(row) for row in cur.fetchall()]


def try_acquire_run_lock(conn: psycopg.Connection[Any]) -> bool:
    with conn.cursor() as cur:
        cur.execute("SELECT pg_try_advisory_lock(%s)", (HOT_RSS_ADVISORY_LOCK_KEY,))
        row = cur.fetchone()
    return bool(row[0] if row else False)


def release_run_lock(conn: psycopg.Connection[Any]) -> None:
    try:
        with conn.cursor() as cur:
            cur.execute("SELECT pg_advisory_unlock(%s)", (HOT_RSS_ADVISORY_LOCK_KEY,))
    except Exception:
        pass


def write_run(
    conn: psycopg.Connection[Any],
    *,
    run_id: str,
    source_id: str,
    status: str,
    summary: dict[str, Any],
    error: str = "",
    started_at: str,
    finished_at: str | None,
    limit: int,
) -> None:
    now = _iso()
    with conn.cursor() as cur:
        cur.execute(
            """
            INSERT INTO hot_runs (run_id, source_id, status, request_json, summary_json, error_message, created_at, updated_at, started_at, finished_at)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT(run_id) DO UPDATE SET
              status=excluded.status,
              summary_json=excluded.summary_json,
              error_message=excluded.error_message,
              updated_at=excluded.updated_at,
              finished_at=excluded.finished_at
            """,
            (run_id, source_id, status, Json({"source_id": source_id, "dry_run": False, "limit": limit}), Json(summary), error, started_at, now, started_at, finished_at),
        )
        cur.execute(
            """
            INSERT INTO hot_run_events (run_id, legacy_id, node, message, payload_json, created_at)
            VALUES (%s, 1, 'rss-worker', %s, %s, %s)
            ON CONFLICT(run_id, legacy_id) DO UPDATE SET
              node=excluded.node,
              message=excluded.message,
              payload_json=excluded.payload_json,
              created_at=excluded.created_at
            """,
            (run_id, "RSS source collected" if status == "completed" else "RSS source failed", Json(summary), now),
        )


def persist_items(conn: psycopg.Connection[Any], items: list[dict[str, Any]]) -> None:
    if not items:
        return
    now = _iso()
    with conn.cursor() as cur:
        for item in items:
            cur.execute(
                """
                INSERT INTO hot_items (
                  source_id, external_id, canonical_url, title, summary, body, reason, published_at,
                  author, source_name, source_handle, badge, score, content_type, tags_json,
                  metrics_json, raw_ref_json, content_hash, has_media, status, created_at, updated_at
                ) VALUES (
                  %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, FALSE, %s, %s, %s
                )
                ON CONFLICT(source_id, external_id) DO UPDATE SET
                  canonical_url=excluded.canonical_url,
                  title=excluded.title,
                  summary=excluded.summary,
                  body=excluded.body,
                  reason=excluded.reason,
                  published_at=excluded.published_at,
                  author=excluded.author,
                  source_name=excluded.source_name,
                  source_handle=excluded.source_handle,
                  badge=excluded.badge,
                  score=excluded.score,
                  content_type=excluded.content_type,
                  tags_json=excluded.tags_json,
                  metrics_json=excluded.metrics_json,
                  raw_ref_json=excluded.raw_ref_json,
                  content_hash=excluded.content_hash,
                  has_media=excluded.has_media,
                  status=CASE
                    WHEN hot_items.status = 'hidden'
                     AND COALESCE(hot_items.metrics_json->>'rank_clean_hidden', 'false') <> 'true'
                    THEN hot_items.status
                    ELSE excluded.status
                  END,
                  updated_at=excluded.updated_at
                """,
                (
                    item["source_id"],
                    item["external_id"],
                    item.get("canonical_url") or "",
                    item["title"],
                    item.get("summary") or "",
                    item.get("body") or "",
                    item.get("reason") or "",
                    item.get("published_at"),
                    item.get("author") or "",
                    item.get("source_name") or item["source_id"],
                    item.get("source_handle") or "",
                    item.get("badge") or "",
                    item.get("score") or "",
                    item.get("content_type") or "article",
                    Json(item.get("tags_json") or []),
                    Json(item.get("metrics_json") or {}),
                    Json(item.get("raw_ref_json") or {}),
                    item.get("content_hash") or "",
                    item.get("status") or "published",
                    now,
                    now,
                ),
            )


def _norm_rank_title(title: str) -> str:
    text = str(title or "").lower().strip()
    text = re.sub(r"^[^\w\u4e00-\u9fff]+", "", text)
    text = re.sub(r"[\W_]+", "", text)
    return text[:80]


def _score_value(value: Any) -> int:
    try:
        return int(float(str(value or "0")))
    except Exception:
        return 0


def _rank_clean_reason(row: dict[str, Any], source_counts: dict[str, int], seen_titles: set[str], max_per_source: int) -> str:
    title = str(row.get("title") or "")
    summary = str(row.get("summary") or "")
    source_id = str(row.get("source_id") or "")
    metrics = row.get("metrics_json") if isinstance(row.get("metrics_json"), dict) else {}
    if not title.strip() or not summary.strip():
        return "low_quality_empty_title_or_summary"
    if metrics and metrics.get("ai_related") is not True:
        return "low_quality_not_ai_related"
    if not is_publishable_hot_item(title, summary, str(row.get("source_name") or source_id), {}):
        return "low_quality_publishable_filter"
    normalized = _norm_rank_title(title)
    if normalized and normalized in seen_titles:
        return "duplicate_title"
    if max_per_source > 0 and source_counts.get(source_id, 0) >= max_per_source:
        return "source_quota_exceeded"
    return ""


def clean_ranked_items(conn: psycopg.Connection[Any], config: Config) -> dict[str, Any]:
    if not config.clean_enabled:
        return {"enabled": False, "hidden_count": 0}
    with conn.cursor(row_factory=dict_row) as cur:
        cur.execute(
            """
            SELECT id, source_id, title, summary, source_name, score, metrics_json
            FROM hot_items
            WHERE status = 'published'
              AND (published_at IS NULL OR published_at >= NOW() - (%s::int * INTERVAL '1 day'))
            ORDER BY
              CASE WHEN score ~ '^[0-9]+(\\.[0-9]+)?$' THEN score::numeric ELSE 0 END DESC,
              published_at DESC NULLS LAST,
              id DESC
            """,
            (config.clean_recent_days,),
        )
        rows = [dict(row) for row in cur.fetchall()]
    source_counts: dict[str, int] = {}
    seen_titles: set[str] = set()
    hidden: list[tuple[int, str]] = []
    for row in rows:
        reason = _rank_clean_reason(row, source_counts, seen_titles, config.clean_max_per_source)
        if reason:
            hidden.append((int(row["id"]), reason))
            continue
        source_id = str(row.get("source_id") or "")
        source_counts[source_id] = source_counts.get(source_id, 0) + 1
        normalized = _norm_rank_title(str(row.get("title") or ""))
        if normalized:
            seen_titles.add(normalized)
    if hidden:
        now = _iso()
        with conn.cursor() as cur:
            for item_id, reason in hidden:
                cur.execute(
                    """
                    UPDATE hot_items
                    SET status = 'hidden',
                        metrics_json = metrics_json || %s::jsonb,
                        updated_at = %s
                    WHERE id = %s
                    """,
                    (Json({"rank_clean_hidden": True, "rank_clean_reason": reason, "rank_cleaned_at": now}), now, item_id),
                )
    reasons: dict[str, int] = {}
    for _, reason in hidden:
        reasons[reason] = reasons.get(reason, 0) + 1
    return {
        "enabled": True,
        "candidate_count": len(rows),
        "hidden_count": len(hidden),
        "kept_count": len(rows) - len(hidden),
        "max_per_source": config.clean_max_per_source,
        "recent_days": config.clean_recent_days,
        "reasons": reasons,
    }


def collect_source(config: Config, conn: psycopg.Connection[Any], source: dict[str, Any], *, ai_budget: AIBudget | None = None) -> dict[str, Any]:
    run_id = f"rss-{source['source_id']}-{uuid.uuid4()}"
    started_at = _iso()
    try:
        groups: list[list[dict[str, Any]]] = []
        errors: list[str] = []
        for url in _seed_urls(source):
            try:
                groups.append(parse_feed(config, fetch_text(config, url), source, url))
            except Exception as exc:
                errors.append(f"{url}: {exc}")
        discovered_count = sum(len(group) for group in groups)
        items = [item for group in groups for item in group]
        items, ai_summary = enrich_items_with_ai(config, items, ai_budget=ai_budget)
        persist_items(conn, items)
        finished_at = _iso()
        summary = {
            "discovered_count": discovered_count,
            "hydrated_count": len(items),
            "normalized_count": len(items),
            "persisted_count": len(items),
            "error_count": len(errors),
            "errors": errors[:3],
            "ai_enrichment": ai_summary,
        }
        write_run(conn, run_id=run_id, source_id=source["source_id"], status="completed", summary=summary, started_at=started_at, finished_at=finished_at, limit=config.limit_per_source)
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO hot_checkpoints (source_id, checkpoint_json, updated_at)
                VALUES (%s, %s, %s)
                ON CONFLICT(source_id) DO UPDATE SET checkpoint_json=excluded.checkpoint_json, updated_at=excluded.updated_at
                """,
                (source["source_id"], Json({"last_run_id": run_id, "last_collected_at": finished_at, "errors": errors[:3]}), finished_at),
            )
        return {"source_id": source["source_id"], "status": "completed", "summary": summary, "_run_recorded": True}
    except Exception as exc:
        finished_at = _iso()
        message = str(exc)
        write_run(conn, run_id=run_id, source_id=source["source_id"], status="failed", summary={"discovered_count": 0, "persisted_count": 0}, error=message, started_at=started_at, finished_at=finished_at, limit=config.limit_per_source)
        return {"source_id": source["source_id"], "status": "failed", "error": message, "_run_recorded": True}


def write_source_failure_run(conn: psycopg.Connection[Any], config: Config, source: dict[str, Any], error: str, started_at: str | None = None, finished_at: str | None = None) -> None:
    now = _iso()
    source_id = str(source.get("source_id") or "")
    write_run(
        conn,
        run_id=f"rss-{source_id}-{uuid.uuid4()}",
        source_id=source_id,
        status="failed",
        summary={"discovered_count": 0, "persisted_count": 0, "error_count": 1, "errors": [error][:1]},
        error=error,
        started_at=started_at or now,
        finished_at=finished_at or now,
        limit=config.limit_per_source,
    )


def collect_source_isolated(config: Config, source: dict[str, Any], ai_budget: AIBudget | None) -> dict[str, Any]:
    started_at = _iso()
    try:
        with psycopg.connect(config.database_url) as conn:
            result = collect_source(config, conn, source, ai_budget=ai_budget)
            conn.commit()
            return result
    except Exception as exc:
        return {"source_id": source.get("source_id") or "", "status": "failed", "error": str(exc), "_run_recorded": False, "_started_at": started_at, "_finished_at": _iso()}


def collect_sources(config: Config, conn: psycopg.Connection[Any], sources: list[dict[str, Any]], ai_budget: AIBudget | None) -> list[dict[str, Any]]:
    if not sources:
        return []
    concurrency = max(1, min(config.source_concurrency, len(sources)))
    if concurrency <= 1:
        serial_results = [collect_source(config, conn, source, ai_budget=ai_budget) for source in sources]
        for result in serial_results:
            result.pop("_run_recorded", None)
        return serial_results
    results: list[tuple[int, dict[str, Any]]] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency, thread_name_prefix="hot-rss-source") as executor:
        future_map = {
            executor.submit(collect_source_isolated, config, source, ai_budget): (index, source)
            for index, source in enumerate(sources)
        }
        for future in concurrent.futures.as_completed(future_map):
            index, source = future_map[future]
            try:
                result = future.result()
            except Exception as exc:
                result = {"source_id": source.get("source_id") or "", "status": "failed", "error": str(exc), "_run_recorded": False}
            if result.get("status") != "completed" and result.get("_run_recorded") is not True:
                write_source_failure_run(
                    conn,
                    config,
                    source,
                    str(result.get("error") or "source collection failed"),
                    started_at=result.get("_started_at"),
                    finished_at=result.get("_finished_at"),
                )
                result["_run_recorded"] = True
            result.pop("_run_recorded", None)
            result.pop("_started_at", None)
            result.pop("_finished_at", None)
            results.append((index, result))
    return [result for _, result in sorted(results, key=lambda item: item[0])]


def run_collect(config: Config) -> None:
    config.validate()
    previous = read_status(config)
    started = _now()
    clean_summary: dict[str, Any] = {"enabled": config.clean_enabled, "hidden_count": 0}
    with psycopg.connect(config.database_url) as conn:
        if not try_acquire_run_lock(conn):
            print("[hot-rss-worker] collect skipped advisory_lock_held", flush=True)
            return
        try:
            write_status(
                config,
                "running",
                last_started_at=_iso(started),
                run_count=int(previous.get("run_count") or 0),
                success_count=int(previous.get("success_count") or 0),
                failure_count=int(previous.get("failure_count") or 0),
            )
            sources = load_sources(conn)
            print(f"[hot-rss-worker] collect started sources={len(sources)}", flush=True)
            ai_budget = AIBudget(config.ai_max_items_per_run) if config.ai_enrich_active() else None
            results = collect_sources(config, conn, sources, ai_budget)
            clean_summary = clean_ranked_items(conn, config)
            conn.commit()
        finally:
            release_run_lock(conn)
    failed = [result for result in results if result.get("status") != "completed"]
    finished = _now()
    ai_results = [(result.get("summary") or {}).get("ai_enrichment") or {} for result in results]
    extra = {
        "last_started_at": _iso(started),
        "last_finished_at": _iso(finished),
        "last_run_duration_ms": int((finished - started).total_seconds() * 1000),
        "source_count": len(sources),
        "item_count": sum(int((result.get("summary") or {}).get("persisted_count") or 0) for result in results),
        "failed_source_count": len(failed),
        "ai_enrichment": {
            "enabled": config.ai_enrich_active(),
            "model": config.ai_model if config.ai_enrich_active() else "",
            "max_items_per_run": config.ai_max_items_per_run,
            "attempted_count": sum(int(result.get("attempted_count") or 0) for result in ai_results),
            "enriched_count": sum(int(result.get("enriched_count") or 0) for result in ai_results),
            "rejected_count": sum(int(result.get("rejected_count") or 0) for result in ai_results),
            "error_count": sum(int(result.get("error_count") or 0) for result in ai_results),
        },
        "rank_cleaning": clean_summary,
        "recent_results": results[-10:],
        "run_count": int(previous.get("run_count") or 0) + 1,
        "success_count": int(previous.get("success_count") or 0) + (1 if not failed else 0),
        "failure_count": int(previous.get("failure_count") or 0) + (1 if failed else 0),
    }
    if failed:
        write_status(config, "error", **extra, error_message=f"{len(failed)} sources failed")
        raise RuntimeError(f"{len(failed)} sources failed")
    write_status(config, "ok", **extra)
    print(f"[hot-rss-worker] collect finished items={extra['item_count']}", flush=True)


def run_loop(config: Config) -> None:
    backoff = config.interval_ms
    completed = 0
    if config.collect_on_start or config.once:
        run_collect(config)
        completed += 1
    if config.once or (config.max_runs > 0 and completed >= config.max_runs):
        return
    while True:
        time.sleep(backoff / 1000)
        try:
            run_collect(config)
            completed += 1
            backoff = config.interval_ms
            if config.max_runs > 0 and completed >= config.max_runs:
                return
        except Exception as exc:
            print(f"[hot-rss-worker] collect failed: {exc}", file=sys.stderr, flush=True)
            backoff = min(backoff * 2, config.max_backoff_ms)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--once", action="store_true")
    parser.add_argument("--healthcheck", action="store_true")
    args = parser.parse_args(argv)
    config = Config.from_env(once=args.once)
    try:
        if args.healthcheck:
            run_healthcheck(config)
        else:
            run_loop(config)
        return 0
    except Exception as exc:
        if not args.healthcheck:
            try:
                write_status(config, "fatal", last_failed_at=_iso(), error_message=str(exc))
            except Exception:
                pass
        print(f"[hot-rss-worker] {'healthcheck failed' if args.healthcheck else 'fatal'}: {exc}", file=sys.stderr, flush=True)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
