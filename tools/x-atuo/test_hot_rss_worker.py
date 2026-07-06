import json
import sys
import tempfile
import types
import unittest
from dataclasses import replace
from pathlib import Path

if "psycopg" not in sys.modules:
    psycopg_stub = types.ModuleType("psycopg")
    psycopg_stub.Connection = object
    psycopg_stub.connect = lambda *_args, **_kwargs: None
    rows_stub = types.ModuleType("psycopg.rows")
    rows_stub.dict_row = object()
    types_stub = types.ModuleType("psycopg.types")
    json_stub = types.ModuleType("psycopg.types.json")
    json_stub.Json = lambda value: value
    sys.modules["psycopg"] = psycopg_stub
    sys.modules["psycopg.rows"] = rows_stub
    sys.modules["psycopg.types"] = types_stub
    sys.modules["psycopg.types.json"] = json_stub

import x_atuo.hot_rss_worker as hot_rss_worker
from x_atuo.hot_rss_worker import AIBudget, Config, _rank_clean_reason, collect_ai_trends_items, collect_sources, enrich_items_with_ai, parse_feed, run_collect, run_healthcheck, is_ai_related


def _config(limit: int = 10, ai_api_key: str = "") -> Config:
    return Config(
        database_url="postgres://example",
        status_path=Path("/tmp/hot-worker-status.json"),
        interval_ms=30_000,
        max_backoff_ms=60_000,
        collect_on_start=True,
        max_runs=0,
        once=True,
        limit_per_source=limit,
        source_concurrency=6,
        request_timeout_ms=20_000,
        health_max_age_ms=0,
        ai_enrich_enabled=True,
        ai_api_key=ai_api_key,
        ai_base_url="https://api.openai.com/v1",
        ai_model="gpt-test",
        ai_timeout_ms=30_000,
        ai_max_items_per_source=5,
        ai_max_items_per_run=20,
        clean_enabled=True,
        clean_recent_days=7,
        clean_max_per_source=2,
    )


def _source(**config):
    return {
        "source_id": "ai-hot-techcrunch",
        "adapter_kind": "rss-generic",
        "title": "TechCrunch AI",
        "base_url": "https://techcrunch.com",
        "seed_urls_json": ["https://techcrunch.com/category/artificial-intelligence/feed/"],
        "config_json": {"category": "news", "ai_only": True, "source_weight": 10, **config},
    }


def _ai_trends_source(**config):
    return {
        "source_id": "ai-hot-trends",
        "adapter_kind": "ai-trends",
        "title": "AI Trends",
        "base_url": "https://news.ycombinator.com",
        "seed_urls_json": [],
        "config_json": {
            "category": "news",
            "category_label": "AI 资讯",
            "ai_only": True,
            "source_weight": 7,
            "sources": ["hackernews", "devto"],
            **config,
        },
    }


class HotRSSWorkerAIHotRulesTest(unittest.TestCase):
    def test_ai_related_rejects_non_ai_blacklist(self):
        self.assertFalse(is_ai_related("护肤新品发布", "美容品牌推出新品", {"ai_only": False}))
        self.assertTrue(is_ai_related("OpenAI releases new agent model", "", {"ai_only": True}))

    def test_parse_feed_filters_non_ai_items(self):
        xml = """
        <rss><channel>
          <item>
            <title>OpenAI releases a new coding agent</title>
            <link>https://example.com/ai</link>
            <description>Agent workflow for software teams</description>
            <pubDate>Mon, 06 Jul 2026 01:00:00 GMT</pubDate>
            <guid>ai-1</guid>
          </item>
          <item>
            <title>Best summer travel hotels</title>
            <link>https://example.com/travel</link>
            <description>旅游酒店攻略</description>
            <pubDate>Mon, 06 Jul 2026 01:00:00 GMT</pubDate>
            <guid>travel-1</guid>
          </item>
        </channel></rss>
        """
        items = parse_feed(_config(), xml, _source(), "https://example.com/feed.xml")
        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]["title"], "OpenAI releases a new coding agent")

    def test_parse_feed_marks_source_weight_in_metrics(self):
        xml = """
        <rss><channel>
          <item>
            <title>DeepSeek V4 model update improves reasoning</title>
            <link>https://example.com/deepseek</link>
            <description>LLM reasoning and inference update</description>
            <pubDate>Mon, 06 Jul 2026 01:00:00 GMT</pubDate>
            <guid>model-1</guid>
          </item>
        </channel></rss>
        """
        items = parse_feed(_config(), xml, _source(), "https://example.com/feed.xml")
        self.assertEqual(len(items), 1)
        metrics = items[0]["metrics_json"]
        self.assertEqual(metrics["source_weight"], 10)
        self.assertEqual(metrics["ai_hot_rules_version"], 1)
        self.assertGreaterEqual(int(items[0]["score"]), 80)

    def test_noisy_sources_are_not_published_by_default(self):
        xml = """
        <rss><channel>
          <item>
            <title>OpenAI agent discussion on Hacker News</title>
            <link>https://news.ycombinator.com/item?id=1</link>
            <description>LLM and agent discussion</description>
            <pubDate>Mon, 06 Jul 2026 01:00:00 GMT</pubDate>
            <guid>hn-1</guid>
          </item>
        </channel></rss>
        """
        noisy_source = _source(noisy=True)
        noisy_source["title"] = "Hacker News AI"
        self.assertEqual(parse_feed(_config(), xml, noisy_source, "https://example.com/feed.xml"), [])

    def test_parse_feed_raw_ref_json_remains_serializable(self):
        xml = """
        <rss><channel>
          <item>
            <title>Claude agent improves AI coding</title>
            <link>https://example.com/claude</link>
            <description>AI coding assistant update</description>
            <guid>claude-1</guid>
          </item>
        </channel></rss>
        """
        item = parse_feed(_config(), xml, _source(), "https://example.com/feed.xml")[0]
        json.dumps(item["raw_ref_json"])
        json.dumps(item["metrics_json"])

    def test_collect_ai_trends_items_uses_hackernews_and_devto(self):
        original_fetch_json = hot_rss_worker.fetch_json

        def fake_fetch_json(_config, url):
            if url.endswith("/topstories.json"):
                return [101, 102]
            if url.endswith("/item/101.json"):
                return {
                    "id": 101,
                    "type": "story",
                    "title": "OpenAI releases a new coding agent",
                    "url": "https://example.com/openai-agent?utm_source=hn",
                    "score": 120,
                    "time": 1_783_320_000,
                    "by": "alice",
                }
            if url.endswith("/item/102.json"):
                return {
                    "id": 102,
                    "type": "story",
                    "title": "Best summer travel hotels",
                    "url": "https://example.com/travel",
                    "score": 99,
                    "time": 1_783_320_000,
                    "by": "bob",
                }
            if url.startswith("https://dev.to/api/articles"):
                return [
                    {
                        "id": 201,
                        "title": "Claude improves AI coding workflows",
                        "url": "https://dev.to/example/claude-ai-coding",
                        "description": "LLM agent update for software teams",
                        "public_reactions_count": 17,
                        "published_at": "2026-07-06T01:00:00Z",
                        "user": {"username": "carol"},
                    }
                ]
            raise AssertionError(url)

        hot_rss_worker.fetch_json = fake_fetch_json
        try:
            items, errors = collect_ai_trends_items(_config(limit=10), _ai_trends_source(limit_per_source=10))
        finally:
            hot_rss_worker.fetch_json = original_fetch_json

        self.assertEqual(errors, [])
        self.assertEqual(len(items), 2)
        self.assertEqual({item["source_name"] for item in items}, {"Hacker News", "Dev.to"})
        self.assertTrue(all(item["source_id"] == "ai-hot-trends" for item in items))
        self.assertTrue(all(item["metrics_json"]["adapter_kind"] == "ai-trends" for item in items))
        self.assertEqual(items[0]["canonical_url"], "https://example.com/openai-agent")

    def test_collect_ai_trends_items_dedupes_by_canonical_url(self):
        original_fetch_json = hot_rss_worker.fetch_json

        def fake_fetch_json(_config, url):
            if url.endswith("/topstories.json"):
                return [101]
            if url.endswith("/item/101.json"):
                return {
                    "id": 101,
                    "type": "story",
                    "title": "OpenAI releases a new coding agent",
                    "url": "https://example.com/openai-agent?utm_source=hn",
                    "score": 120,
                    "time": 1_783_320_000,
                }
            if url.startswith("https://dev.to/api/articles"):
                return [
                    {
                        "id": 201,
                        "title": "OpenAI releases a new coding agent",
                        "url": "https://example.com/openai-agent?utm_medium=social",
                        "description": "AI agent update",
                        "public_reactions_count": 10,
                        "published_at": "2026-07-06T01:00:00Z",
                    }
                ]
            raise AssertionError(url)

        hot_rss_worker.fetch_json = fake_fetch_json
        try:
            items, _errors = collect_ai_trends_items(_config(limit=10), _ai_trends_source(limit_per_source=10))
        finally:
            hot_rss_worker.fetch_json = original_fetch_json

        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]["canonical_url"], "https://example.com/openai-agent")

    def test_ai_enrichment_skips_without_api_key(self):
        item = {"title": "OpenAI launches agent", "summary": "AI agent update", "metrics_json": {}}
        items, summary = enrich_items_with_ai(_config(ai_api_key=""), [item])
        self.assertEqual(items, [item])
        self.assertFalse(summary["enabled"])
        self.assertEqual(summary["attempted_count"], 0)

    def test_ai_enrichment_uses_api_key_and_updates_display_fields(self):
        item = {
            "title": "OpenAI launches agent",
            "summary": "AI agent update",
            "source_name": "TechCrunch AI",
            "metrics_json": {},
            "tags_json": [],
        }
        calls = []
        original = hot_rss_worker.ai_clean_item

        def fake_clean(config, candidate):
            calls.append((config.ai_api_key, candidate["title"]))
            return {
                "publish": True,
                "title_zh": "OpenAI 发布智能体",
                "ai_summary": "OpenAI 发布面向开发者的智能体能力。",
                "reason": "AI news",
                "tags": ["OpenAI", "Agent"],
            }

        hot_rss_worker.ai_clean_item = fake_clean
        try:
            items, summary = enrich_items_with_ai(_config(ai_api_key="sk-test"), [item])
        finally:
            hot_rss_worker.ai_clean_item = original

        self.assertEqual(calls, [("sk-test", "OpenAI launches agent")])
        self.assertTrue(summary["enabled"])
        self.assertEqual(summary["attempted_count"], 1)
        self.assertEqual(summary["enriched_count"], 1)
        self.assertEqual(items[0]["title"], "OpenAI 发布智能体")
        self.assertEqual(items[0]["summary"], "OpenAI 发布面向开发者的智能体能力。")
        self.assertIn("ai_summary", items[0]["metrics_json"])

    def test_ai_enrichment_filters_rejected_items(self):
        item = {"title": "Beauty product launch", "summary": "makeup", "metrics_json": {}, "tags_json": []}
        original = hot_rss_worker.ai_clean_item

        def fake_clean(_config, _candidate):
            return {"publish": False, "title_zh": "", "ai_summary": "", "reason": "not AI", "tags": []}

        hot_rss_worker.ai_clean_item = fake_clean
        try:
            items, summary = enrich_items_with_ai(_config(ai_api_key="sk-test"), [item])
        finally:
            hot_rss_worker.ai_clean_item = original

        self.assertEqual(items, [])
        self.assertEqual(summary["rejected_count"], 1)

    def test_ai_enrichment_respects_explicit_max_items(self):
        items = [
            {"title": "OpenAI launches agent", "summary": "AI agent update", "metrics_json": {}, "tags_json": []},
            {"title": "Claude improves coding", "summary": "AI coding update", "metrics_json": {}, "tags_json": []},
        ]
        calls = []
        original = hot_rss_worker.ai_clean_item

        def fake_clean(_config, candidate):
            calls.append(candidate["title"])
            return {"publish": True, "title_zh": "", "ai_summary": "", "reason": "ok", "tags": []}

        hot_rss_worker.ai_clean_item = fake_clean
        try:
            output, summary = enrich_items_with_ai(_config(ai_api_key="sk-test"), items, max_items=1)
        finally:
            hot_rss_worker.ai_clean_item = original

        self.assertEqual(calls, ["OpenAI launches agent"])
        self.assertEqual(len(output), 2)
        self.assertEqual(summary["attempted_count"], 1)

    def test_ai_enrichment_respects_shared_run_budget(self):
        calls = []
        original = hot_rss_worker.ai_clean_item

        def fake_clean(_config, candidate):
            calls.append(candidate["title"])
            return {"publish": True, "title_zh": "", "ai_summary": "", "reason": "ok", "tags": []}

        hot_rss_worker.ai_clean_item = fake_clean
        try:
            budget = AIBudget(1)
            first = [{"title": "OpenAI launches agent", "summary": "AI agent update", "metrics_json": {}, "tags_json": []}]
            second = [{"title": "Claude improves coding", "summary": "AI coding update", "metrics_json": {}, "tags_json": []}]
            _, first_summary = enrich_items_with_ai(_config(ai_api_key="sk-test"), first, ai_budget=budget)
            _, second_summary = enrich_items_with_ai(_config(ai_api_key="sk-test"), second, ai_budget=budget)
        finally:
            hot_rss_worker.ai_clean_item = original

        self.assertEqual(calls, ["OpenAI launches agent"])
        self.assertEqual(first_summary["attempted_count"], 1)
        self.assertEqual(second_summary["attempted_count"], 0)

    def test_parse_feed_falls_back_to_xml_when_feedparser_fails(self):
        original = hot_rss_worker.feedparser

        class BrokenFeedparser:
            @staticmethod
            def parse(_xml):
                raise RuntimeError("broken parser")

        hot_rss_worker.feedparser = BrokenFeedparser()
        try:
            xml = """
            <rss><channel>
              <item>
                <title>OpenAI releases a new coding agent</title>
                <link>https://example.com/ai</link>
                <description>Agent workflow for software teams</description>
                <guid>ai-1</guid>
              </item>
            </channel></rss>
            """
            items = parse_feed(_config(), xml, _source(), "https://example.com/feed.xml")
        finally:
            hot_rss_worker.feedparser = original

        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]["metrics_json"]["parser"], "xml")

    def test_feedparser_bad_content_shape_only_skips_bad_fields(self):
        original = hot_rss_worker.feedparser

        class FakeFeedparser:
            @staticmethod
            def parse(_xml):
                return types.SimpleNamespace(
                    entries=[
                        {
                            "title": "OpenAI updates coding agent",
                            "link": "https://example.com/openai-agent",
                            "content": [],
                            "id": "agent-1",
                        }
                    ]
                )

        hot_rss_worker.feedparser = FakeFeedparser()
        try:
            items = parse_feed(_config(), "<rss>bad xml", _source(), "https://example.com/feed.xml")
        finally:
            hot_rss_worker.feedparser = original

        self.assertEqual(len(items), 1)
        self.assertEqual(items[0]["metrics_json"]["parser"], "feedparser")

    def test_rank_clean_reason_detects_duplicates_and_source_quota(self):
        seen = {"openaireleasesagent"}
        duplicate = {
            "title": "OpenAI releases agent",
            "summary": "AI agent update",
            "source_id": "src-1",
            "source_name": "TechCrunch AI",
        }
        self.assertEqual(_rank_clean_reason(duplicate, {}, seen, 2), "duplicate_title")
        quota = {
            "title": "Claude improves AI coding",
            "summary": "AI coding update",
            "source_id": "src-1",
            "source_name": "TechCrunch AI",
        }
        self.assertEqual(_rank_clean_reason(quota, {"src-1": 2}, set(), 2), "source_quota_exceeded")

    def test_rank_clean_reason_uses_existing_ai_related_metric(self):
        row = {
            "title": "月之暗面发布新版",
            "summary": "上下文窗口升级",
            "source_id": "src-1",
            "source_name": "TechCrunch AI",
            "metrics_json": {"ai_related": True},
        }
        self.assertEqual(_rank_clean_reason(row, {}, set(), 2), "")

        row["metrics_json"] = {"ai_related": False}
        self.assertEqual(_rank_clean_reason(row, {}, set(), 2), "low_quality_not_ai_related")

    def test_healthcheck_rejects_advisory_lock_skip_status(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            config = replace(_config(), status_path=Path(temp_dir) / "status.json", health_max_age_ms=60_000)
            config.status_path.write_text(
                json.dumps({"status": "ok", "updated_at": hot_rss_worker._iso(), "skipped_reason": "advisory_lock_held"}),
                encoding="utf-8",
            )

            with self.assertRaisesRegex(RuntimeError, "advisory lock"):
                run_healthcheck(config)

    def test_run_collect_does_not_refresh_status_when_advisory_lock_is_held(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            config = replace(_config(), status_path=Path(temp_dir) / "status.json")
            original_status = {"status": "ok", "updated_at": "2026-07-06T00:00:00Z", "run_count": 7}
            config.status_path.write_text(json.dumps(original_status), encoding="utf-8")
            original_connect = hot_rss_worker.psycopg.connect
            original_try_lock = hot_rss_worker.try_acquire_run_lock

            class FakeConn:
                def __enter__(self):
                    return self

                def __exit__(self, *_args):
                    return False

            hot_rss_worker.psycopg.connect = lambda _url: FakeConn()
            hot_rss_worker.try_acquire_run_lock = lambda _conn: False
            try:
                run_collect(config)
            finally:
                hot_rss_worker.psycopg.connect = original_connect
                hot_rss_worker.try_acquire_run_lock = original_try_lock

            self.assertEqual(json.loads(config.status_path.read_text(encoding="utf-8")), original_status)

    def test_collect_sources_records_parallel_connection_failure(self):
        config = replace(_config(), source_concurrency=2)
        source = _source()
        source["source_id"] = "src-fails-before-run"
        original_collect_isolated = hot_rss_worker.collect_source_isolated

        class FakeCursor:
            def __init__(self):
                self.statements = []

            def __enter__(self):
                return self

            def __exit__(self, *_args):
                return False

            def execute(self, sql, params=None):
                self.statements.append((sql, params))

        class FakeConn:
            def __init__(self):
                self.cursor_obj = FakeCursor()

            def cursor(self, *_args, **_kwargs):
                return self.cursor_obj

        def fake_collect_isolated(_config, _source, _ai_budget):
            return {"source_id": "src-fails-before-run", "status": "failed", "error": "connection refused", "_run_recorded": False}

        hot_rss_worker.collect_source_isolated = fake_collect_isolated
        try:
            conn = FakeConn()
            results = collect_sources(config, conn, [source, source], None)
        finally:
            hot_rss_worker.collect_source_isolated = original_collect_isolated

        self.assertEqual([result["status"] for result in results], ["failed", "failed"])
        self.assertNotIn("_run_recorded", results[0])
        inserts = [sql for sql, _params in conn.cursor_obj.statements if "INSERT INTO hot_runs" in sql]
        self.assertEqual(len(inserts), 2)


if __name__ == "__main__":
    unittest.main()
