BEGIN;

-- 読み取り専用ユーザー
CREATE USER rouser IDENTIFIED BY 'pwrouser';
GRANT SELECT, SHOW VIEW ON devdb.* TO 'rouser';

-- アプリケーション接続用ユーザー
CREATE USER appuser IDENTIFIED BY 'pwappuser';
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE TEMPORARY TABLES, SHOW VIEW ON devdb.* TO 'appuser';

COMMIT;
