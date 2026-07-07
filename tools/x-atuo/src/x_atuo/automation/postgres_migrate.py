from __future__ import annotations

import argparse
import sqlite3
from pathlib import Path
from typing import Any

from x_atuo.automation.author_alpha_schema import initialize_author_alpha_storage_schema
from x_atuo.automation.db import connect_postgres
from x_atuo.automation.storage_schema import initialize_automation_storage_schema

AUTOMATION_TABLES = [
    "jobs",
    "runs",
    "audit_events",
    "dedupe_keys",
    "engagements",
    "shared_engagements",
    "candidate_cache",
]

AUTHOR_ALPHA_TABLES = [
    "alpha_authors",
    "alpha_reply_daily_metrics",
    "alpha_author_daily_rollups",
    "alpha_sync_runs",
    "alpha_sync_checkpoints",
    "alpha_engagements",
    "alpha_runs",
    "alpha_run_audit_events",
]


def migrate_sqlite_to_postgres(
    *,
    sqlite_path: Path,
    database_url: str,
    tables: list[str],
    replace_existing: bool,
) -> dict[str, int]:
    if not sqlite_path.exists():
        raise FileNotFoundError(f"SQLite database not found: {sqlite_path}")
    source = sqlite3.connect(sqlite_path)
    source.row_factory = sqlite3.Row
    imported: dict[str, int] = {}
    try:
        with connect_postgres(database_url) as target:
            for table in tables:
                if not _sqlite_table_exists(source, table):
                    imported[table] = 0
                    continue
                if replace_existing:
                    target.execute(f"DELETE FROM {table}")
                rows = source.execute(f"SELECT * FROM {table}").fetchall()
                if not rows:
                    imported[table] = 0
                    continue
                columns = [str(column) for column in rows[0].keys()]
                placeholders = ", ".join("?" for _ in columns)
                column_list = ", ".join(columns)
                target.executemany(
                    f"INSERT INTO {table} ({column_list}) VALUES ({placeholders}) ON CONFLICT DO NOTHING",
                    [tuple(row[column] for column in columns) for row in rows],
                )
                imported[table] = len(rows)
            _reset_postgres_sequences(target)
    finally:
        source.close()
    return imported


def _sqlite_table_exists(connection: sqlite3.Connection, table: str) -> bool:
    row = connection.execute(
        "SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?",
        (table,),
    ).fetchone()
    return row is not None


def _reset_postgres_sequences(connection: Any) -> None:
    sequence_tables = {
        "audit_events": "id",
        "engagements": "id",
        "shared_engagements": "id",
        "alpha_engagements": "id",
        "alpha_run_audit_events": "id",
    }
    for table, column in sequence_tables.items():
        if not connection.has_table(table):
            continue
        connection.execute(
            """
            SELECT setval(
                pg_get_serial_sequence(?, ?),
                COALESCE((SELECT MAX(id) FROM %s), 1),
                (SELECT COUNT(*) > 0 FROM %s)
            )
            """
            % (table, table),
            (table, column),
        )


def main() -> None:
    parser = argparse.ArgumentParser(description="Migrate X Auto SQLite storage into PostgreSQL.")
    parser.add_argument("--database-url", required=True, help="Target PostgreSQL connection string.")
    parser.add_argument("--automation-sqlite", type=Path, help="Path to x_atuo.sqlite3.")
    parser.add_argument("--author-alpha-sqlite", type=Path, help="Path to author_alpha.sqlite3.")
    parser.add_argument("--replace-existing", action="store_true", help="Delete target rows before importing each table.")
    args = parser.parse_args()

    with connect_postgres(args.database_url) as target:
        initialize_automation_storage_schema(target)
        initialize_author_alpha_storage_schema(target)

    result: dict[str, dict[str, int]] = {}
    if args.automation_sqlite:
        result["automation"] = migrate_sqlite_to_postgres(
            sqlite_path=args.automation_sqlite,
            database_url=args.database_url,
            tables=AUTOMATION_TABLES,
            replace_existing=args.replace_existing,
        )
    if args.author_alpha_sqlite:
        result["author_alpha"] = migrate_sqlite_to_postgres(
            sqlite_path=args.author_alpha_sqlite,
            database_url=args.database_url,
            tables=AUTHOR_ALPHA_TABLES,
            replace_existing=args.replace_existing,
        )
    for scope, tables in result.items():
        for table, count in tables.items():
            print(f"{scope}.{table}={count}")


if __name__ == "__main__":
    main()
