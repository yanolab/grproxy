CREATE TABLE IF NOT EXISTS tests (
  id         VARCHAR(256) NOT NULL
, created_at DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
, updated_at DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
, PRIMARY KEY(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;