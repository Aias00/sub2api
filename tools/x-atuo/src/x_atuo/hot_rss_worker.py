"""Hot RSS collector worker.

This keeps the status file and PostgreSQL writes compatible with the retired
Node collector so the Go API and admin worker page can keep reading the same
runtime contract.
"""

from __future__ import annotations

import argparse
import hashlib
import html
import json
import os
import re
import sys
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
    request_timeout_ms: int
    health_max_age_ms: int

    @classmethod
    def from_env(cls, *, once: bool = False) -> "Config":
        return cls(
            database_url=_env("DATABASE_URL", ""),
            status_path=Path(_env("HOT_RSS_WORKER_STATUS_PATH", _env("HOT_WORKER_STATUS_PATH", "/app/runtime/hot-worker-status.json"))),
            interval_ms=_int_env("HOT_RSS_COLLECT_INTERVAL_MS", 30 * 60 * 1000),
            max_backoff_ms=_int_env("HOT_RSS_COLLECT_MAX_BACKOFF_MS", 10 * 60 * 1000),
            collect_on_start=_bool_env("HOT_RSS_COLLECT_ON_START", True),
            max_runs=_int_env("HOT_RSS_COLLECT_MAX_RUNS", 0),
            once=once or _bool_env("HOT_RSS_COLLECT_ONCE", False),
            limit_per_source=_int_env("HOT_RSS_LIMIT_PER_SOURCE", 10),
            request_timeout_ms=_int_env("HOT_RSS_REQUEST_TIMEOUT_MS", 20 * 1000),
            health_max_age_ms=_int_env("HOT_RSS_WORKER_HEALTH_MAX_AGE_MS", 0),
        )

    def validate(self) -> None:
        if not self.database_url:
            raise RuntimeError("DATABASE_URL is required")

    def status_max_age_ms(self) -> int:
        if self.health_max_age_ms > 0:
            return self.health_max_age_ms
        return max(self.interval_ms * 2, self.interval_ms + self.max_backoff_ms)


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


def _score_for_item(published_at: str, tags: list[str], title: str, source_config: dict[str, Any]) -> int:
    category = source_config.get("category") or "blog"
    score = 70 if category == "official" else 60
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


def parse_feed(config: Config, xml: str, source: dict[str, Any], feed_url: str) -> list[dict[str, Any]]:
    xml = re.sub(r"<(/?)content:encoded", r"<\1content", xml)
    root = ET.fromstring(xml.encode("utf-8"))
    entries = [elem for elem in root.iter() if _local_name(elem.tag) in {"item", "entry"}]
    items: list[dict[str, Any]] = []
    for entry in entries[: config.limit_per_source]:
        title = _first_text(entry, "title")
        link = _link(entry)
        summary = _first_text(entry, "description", "summary", "content", "encoded")
        published_at = _normalize_date(_first_text(entry, "pubDate", "published", "updated", "date"))
        author = _first_text(entry, "author", "creator", "name")
        guid = _first_text(entry, "guid", "id") or link or title
        tags = _categories(entry)
        hash_input = "\n".join([str(source.get("source_id") or ""), guid, link, title, summary])
        content_hash = hashlib.sha256(hash_input.encode("utf-8")).hexdigest()
        external_id = f"{feed_url}::{source.get('source_id')}:{guid or content_hash}"
        source_config_value = _source_config(source)
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
            "source_name": _source_display_name(source),
            "source_handle": _source_handle(source, feed_url),
            "badge": _source_category_label(source_config_value),
            "score": str(hot_score),
            "content_type": "article",
            "tags_json": tags,
            "metrics_json": metrics,
            "raw_ref_json": {"feed_url": feed_url, "guid": guid},
            "content_hash": content_hash,
        }
        if item["title"] and item["external_id"]:
            items.append(item)
    return items


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
                  metrics_json, raw_ref_json, content_hash, has_media, created_at, updated_at
                ) VALUES (
                  %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, FALSE, %s, %s
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
                    now,
                    now,
                ),
            )


def collect_source(config: Config, conn: psycopg.Connection[Any], source: dict[str, Any]) -> dict[str, Any]:
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
        items = [item for group in groups for item in group]
        persist_items(conn, items)
        finished_at = _iso()
        summary = {
            "discovered_count": len(items),
            "hydrated_count": len(items),
            "normalized_count": len(items),
            "persisted_count": len(items),
            "error_count": len(errors),
            "errors": errors[:3],
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
        return {"source_id": source["source_id"], "status": "completed", "summary": summary}
    except Exception as exc:
        finished_at = _iso()
        message = str(exc)
        write_run(conn, run_id=run_id, source_id=source["source_id"], status="failed", summary={"discovered_count": 0, "persisted_count": 0}, error=message, started_at=started_at, finished_at=finished_at, limit=config.limit_per_source)
        return {"source_id": source["source_id"], "status": "failed", "error": message}


def run_collect(config: Config) -> None:
    config.validate()
    previous = read_status(config)
    started = _now()
    write_status(
        config,
        "running",
        last_started_at=_iso(started),
        run_count=int(previous.get("run_count") or 0),
        success_count=int(previous.get("success_count") or 0),
        failure_count=int(previous.get("failure_count") or 0),
    )
    with psycopg.connect(config.database_url) as conn:
        sources = load_sources(conn)
        print(f"[hot-rss-worker] collect started sources={len(sources)}", flush=True)
        results = [collect_source(config, conn, source) for source in sources]
        conn.commit()
    failed = [result for result in results if result.get("status") != "completed"]
    finished = _now()
    extra = {
        "last_started_at": _iso(started),
        "last_finished_at": _iso(finished),
        "last_run_duration_ms": int((finished - started).total_seconds() * 1000),
        "source_count": len(sources),
        "item_count": sum(int((result.get("summary") or {}).get("persisted_count") or 0) for result in results),
        "failed_source_count": len(failed),
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
