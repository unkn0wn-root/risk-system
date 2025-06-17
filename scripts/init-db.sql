CREATE DATABASE users;
CREATE DATABASE risk_analytics;

CREATE USER risk_admin WITH PASSWORD 'risky_password';
CREATE USER app_admin WITH PASSWORD 'app_password';

GRANT ALL PRIVILEGES ON DATABASE risk_analytics TO risk_admin;
GRANT ALL PRIVILEGES ON DATABASE users TO app_admin;

\c risk_analytics;
GRANT ALL ON SCHEMA public TO risk_admin;

\c users;
GRANT ALL ON SCHEMA public TO app_admin;
