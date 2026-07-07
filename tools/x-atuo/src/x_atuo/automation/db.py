from __future__ import annotations

import re
import sqlite3
from collections.abc import Iterable, Iterator, Sequence
from contextlib import contextmanager
from typing import Any


class PostgresCursor:
    def __init__(self, cursor: Any) -> None:
        self._cursor = cursor
        self.lastrowid: int | None = None

    @property
    def rowcount(self) -> int:
        return int(self._cursor.rowcount or 0)

    def fetchone(self) -> dict[str, Any] | None:
        row = self._cursor.fetchone()
        return dict(row) if row is not None else None

    def fetchall(self) -> list[dict[str, Any]]:
        return [dict(row) for row in self._cursor.fetchall()]


class PostgresConnection:
    dialect = "postgres"

    def __init__(self, connection: Any) -> None:
        self._connection = connection

    def execute(self, sql: str, parameters: Sequence[Any] | None = None) -> PostgresCursor:
        translated = _translate_postgres_sql(sql)
        cursor = self._connection.execute(translated, tuple(parameters or ()))
        wrapped = PostgresCursor(cursor)
        upper = translated.strip().upper()
        if upper.startswith("INSERT INTO AUDIT_EVENTS") or upper.startswith("INSERT INTO ALPHA_RUN_AUDIT_EVENTS"):
            row = cursor.fetchone()
            wrapped.lastrowid = int(row["id"]) if row is not None and row.get("id") is not None else None
        return wrapped

    def executemany(self, sql: str, parameters: Iterable[Sequence[Any]]) -> None:
        self._connection.executemany(_translate_postgres_sql(sql), parameters)

    def executescript(self, script: str) -> None:
        for statement in _split_sql_script(script):
            self.execute(statement)

    def table_columns(self, table_name: str) -> set[str]:
        cursor = self._connection.execute(
            """
            SELECT column_name
            FROM information_schema.columns
            WHERE table_schema = current_schema() AND table_name = %s
            """,
            (table_name,),
        )
        return {str(row["column_name"]) for row in cursor.fetchall()}

    def has_table(self, table_name: str) -> bool:
        cursor = self._connection.execute(
            """
            SELECT 1
            FROM information_schema.tables
            WHERE table_schema = current_schema() AND table_name = %s
            LIMIT 1
            """,
            (table_name,),
        )
        return cursor.fetchone() is not None

    def commit(self) -> None:
        self._connection.commit()

    def close(self) -> None:
        self._connection.close()


@contextmanager
def connect_postgres(database_url: str) -> Iterator[PostgresConnection]:
    try:
        import psycopg
        from psycopg.rows import dict_row
    except ImportError as exc:  # pragma: no cover - exercised only without dependency.
        raise RuntimeError("psycopg is required for X Auto PostgreSQL storage") from exc

    connection = psycopg.connect(database_url, row_factory=dict_row)
    wrapped = PostgresConnection(connection)
    try:
        yield wrapped
        connection.commit()
    finally:
        connection.close()


def connection_dialect(connection: Any) -> str:
    return str(getattr(connection, "dialect", "sqlite"))


def table_columns(connection: Any, table_name: str) -> set[str]:
    if hasattr(connection, "table_columns"):
        return set(connection.table_columns(table_name))
    return {str(row["name"]) for row in connection.execute(f"PRAGMA table_info({table_name})").fetchall()}


def table_exists(connection: Any, table_name: str) -> bool:
    if hasattr(connection, "has_table"):
        return bool(connection.has_table(table_name))
    row = connection.execute(
        """
        SELECT name
        FROM sqlite_master
        WHERE type = 'table' AND name = ?
        """,
        (table_name,),
    ).fetchone()
    return row is not None


def bool_to_db(value: bool) -> int:
    return int(value)


def _split_sql_script(script: str) -> list[str]:
    return [statement.strip() for statement in script.split(";") if statement.strip()]


def _translate_postgres_sql(sql: str) -> str:
    translated = sql.strip()
    if translated.upper() == "BEGIN IMMEDIATE":
        return "BEGIN"
    translated = re.sub(r"\bifnull\s*\(", "COALESCE(", translated, flags=re.IGNORECASE)
    translated = re.sub(r"\bdate\s*\(\s*created_at\s*\)", "LEFT(created_at, 10)", translated, flags=re.IGNORECASE)
    translated = re.sub(r"\bdate\s*\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\)", r"LEFT(\1, 10)", translated, flags=re.IGNORECASE)
    if translated.upper().startswith("INSERT INTO AUDIT_EVENTS") and "RETURNING" not in translated.upper():
        translated += " RETURNING id"
    if translated.upper().startswith("INSERT INTO ALPHA_RUN_AUDIT_EVENTS") and "RETURNING" not in translated.upper():
        translated += " RETURNING id"
    return _replace_qmark_placeholders(translated)


def _replace_qmark_placeholders(sql: str) -> str:
    output: list[str] = []
    in_single_quote = False
    in_double_quote = False
    index = 0
    while index < len(sql):
        char = sql[index]
        if char == "'" and not in_double_quote:
            in_single_quote = not in_single_quote
            output.append(char)
        elif char == '"' and not in_single_quote:
            in_double_quote = not in_double_quote
            output.append(char)
        elif char == "?" and not in_single_quote and not in_double_quote:
            output.append("%s")
        else:
            output.append(char)
        index += 1
    return "".join(output)
