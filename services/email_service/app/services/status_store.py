import logging
from typing import Optional

import psycopg2
from psycopg2 import sql

from app.config.settings import settings

logger = logging.getLogger(__name__)


class StatusStore:
    def __init__(self) -> None:
        self.dsn = settings.status_database_url
        self.table = settings.status_table
        self._ensure_table()

    def _get_connection(self):
        return psycopg2.connect(self.dsn)

    def _ensure_table(self) -> None:
        query = sql.SQL(
            """
            CREATE TABLE IF NOT EXISTS {table} (
                request_id TEXT PRIMARY KEY,
                status TEXT NOT NULL,
                provider TEXT,
                detail TEXT,
                updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
            );
            """
        ).format(table=sql.Identifier(self.table))

        with self._get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute(query)
        logger.info("Status table %s ready", self.table)

    def update_status(self, request_id: str, status: str, provider: str, detail: Optional[str] = None) -> None:
        if not request_id:
            return

        query = sql.SQL(
            """
            INSERT INTO {table} (request_id, status, provider, detail, updated_at)
            VALUES (%s, %s, %s, %s, NOW())
            ON CONFLICT (request_id)
            DO UPDATE SET status = EXCLUDED.status,
                          provider = EXCLUDED.provider,
                          detail = EXCLUDED.detail,
                          updated_at = NOW();
            """
        ).format(table=sql.Identifier(self.table))

        with self._get_connection() as conn:
            with conn.cursor() as cur:
                cur.execute(query, (request_id, status, provider, detail))
        logger.debug("Updated status for %s -> %s", request_id, status)
