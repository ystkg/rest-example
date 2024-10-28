REVOKE ALL ON DATABASE postgres FROM PUBLIC;

CREATE DATABASE devdb;

REVOKE ALL ON DATABASE devdb FROM PUBLIC;

BEGIN;

-- データベースオーナー
CREATE ROLE devdb LOGIN PASSWORD 'pwdevdb';
ALTER DATABASE devdb OWNER TO devdb;

-- 読み取り専用ユーザー
CREATE ROLE rouser LOGIN PASSWORD 'pwrouser';
GRANT CONNECT ON DATABASE devdb TO rouser;
GRANT pg_read_all_data TO rouser;

-- アプリケーション接続用ユーザー
CREATE ROLE appuser LOGIN PASSWORD 'pwappuser';
GRANT CONNECT, TEMPORARY ON DATABASE devdb TO appuser;
GRANT pg_read_all_data, pg_write_all_data TO appuser;

COMMIT;
