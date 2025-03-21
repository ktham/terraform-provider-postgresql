FROM postgres:17-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
  postgresql-$PG_MAJOR-pgaudit \
  postgresql-$PG_MAJOR-pgvector \
