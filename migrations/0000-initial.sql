CREATE ROLE iam WITH NOSUPERUSER INHERIT NOCREATEROLE NOCREATEDB LOGIN NOREPLICATION NOBYPASSRLS ENCRYPTED PASSWORD 'development';

CREATE DATABASE iam;

\c iam

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

GRANT USAGE ON SCHEMA public TO iam;

CREATE OR REPLACE FUNCTION public.tf_set_updated_at()
RETURNS TRIGGER AS
$$
    BEGIN
        NEW.updated_at := now();
        RETURN NEW;
    END;
$$
LANGUAGE 'plpgsql';

ALTER FUNCTION public.tf_set_updated_at OWNER TO lab;

GRANT EXECUTE ON FUNCTION public.tf_set_updated_at TO iam;

REVOKE ALL ON FUNCTION public.tf_set_updated_at FROM public;

CREATE TABLE public.t_user (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    name TEXT NOT NULL CHECK(char_length(name) BETWEEN 4 AND 128),
    username TEXT NOT NULL UNIQUE CHECK(char_length(username) BETWEEN 4 AND 64),
    password TEXT NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    allowed_hours TIME[] CHECK (array_length(allowed_hours, 1) % 2 = 0),
    allowed_networks INET[],
    allowed_days SMALLINT[]
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON public.t_user
FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM  NEW.*)
EXECUTE PROCEDURE public.tf_set_updated_at();

ALTER TABLE public.t_user OWNER TO lab;

GRANT SELECT, UPDATE, DELETE ON TABLE public.t_user TO iam;

REVOKE ALL ON TABLE public.t_user FROM public;

CREATE TABLE public.t_login_attempt (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    donet_at TIMESTAMP,
    user_agent TEXT NOT NULL CHECK(char_length(user_agent) < 256),
    successful BOOLEAN NOT NULL DEFAULT FALSE,
    error TEXT CHECK (char_length(error) < 256)
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON public.t_login_attempt
FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM  NEW.*)
EXECUTE PROCEDURE public.tf_set_updated_at();

ALTER TABLE public.t_login_attempt OWNER TO lab;

GRANT SELECT, UPDATE, DELETE ON TABLE public.t_login_attempt TO iam;

REVOKE ALL ON TABLE public.t_login_attempt FROM public;

CREATE TABLE public.t_session (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    expires_at TIMESTAMP,
    login_attempt_id UUID NOT NULL REFERENCES t_login_attempt(id)
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON public.t_session
FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM  NEW.*)
EXECUTE PROCEDURE public.tf_set_updated_at();

ALTER TABLE public.t_session OWNER TO lab;

GRANT SELECT, UPDATE, DELETE ON TABLE public.t_session TO iam;

REVOKE ALL ON TABLE public.t_session FROM public;

-- TODO

-- CREATE TABLE public.t_role (
-- );
--
-- CREATE TABLE public.t_user_role (
-- );
--
-- CREATE TABLE public.t_action (
-- );
--
-- CREATE TABLE public.t_action_role (
-- );

