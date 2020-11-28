CREATE ROLE iam WITH NOSUPERUSER INHERIT NOCREATEROLE NOCREATEDB LOGIN NOREPLICATION NOBYPASSRLS ENCRYPTED PASSWORD 'development';

CREATE DATABASE iam;

\c iam

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

CREATE TABLE public.t_audit (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    run_at TIMESTAMP NOT NULL DEFAULT clock_timestamp(),
    database_user TEXT NOT NULL,
    application_user TEXT NOT NULL,
    origin_ip INET NOT NULL,
    schema TEXT NOT NULL,
    "table" TEXT NOT NULL,
    operation TEXT NOT NULL,
    query TEXT NOT NULL,
    request_id UUID,
    old JSONB,
    new JSONB
);

ALTER TABLE public.t_audit OWNER TO lab;

REVOKE ALL ON TABLE public.t_audit FROM public;

GRANT INSERT ON TABLE public.t_audit TO public;

CREATE OR REPLACE FUNCTION public.tf_add_audit()
RETURNS TRIGGER AS
$$
    DECLARE
        _old JSONB := NULL;
        _new JSONB := NULL;

        _user_id    TEXT := NULL;
        _request_id TEXT := NULL;

        _super      BOOLEAN := FALSE;
    BEGIN
        IF TG_OP = 'INSERT' THEN
            _new := to_jsonb(NEW.*);
        END IF;

        IF TG_OP = 'UPDATE' THEN
            _old := to_jsonb(OLD.*);
            _new := to_jsonb(NEW.*);
        END IF;

        IF TG_OP = 'DELETE' THEN
            _old := to_jsonb(OLD.*);
        END IF;

        BEGIN
            SHOW application.user_id    INTO _user_id;
            SHOW application.request_id INTO _request_id;
        EXCEPTION WHEN OTHERS THEN
            SHOW IS_SUPERUSER INTO _super;
            IF _super THEN
                _user_id := 'SUPER_USER';
                _request_id := NULL;
            ELSE
                RAISE EXCEPTION assert_failure USING HINT = 'unable to perform operations without the associated user/request';
            END IF;

        END;

        INSERT INTO public.t_audit(database_user, application_user, origin_ip, schema, "table", operation, query, request_id, old, new)
        VALUES (CURRENT_USER, _user_id,  COALESCE(inet_client_addr(), '127.0.0.1'::INET),  TG_TABLE_SCHEMA, TG_TABLE_NAME, TG_OP, current_query(), _request_id::UUID, _old, _new);

        IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
            RETURN NEW;
        END IF;

        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        END IF;

        RETURN NULL;
    END;
$$
LANGUAGE 'plpgsql';

ALTER FUNCTION public.tf_add_audit() OWNER TO lab;

REVOKE ALL ON FUNCTION public.tf_add_audit() FROM public;

GRANT EXECUTE ON FUNCTION public.tf_add_audit() TO public;

CREATE TABLE public.t_migration (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    name TEXT NOT NULL CHECK(char_length(name) BETWEEN 4 AND 128),
    rolled_back BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON public.t_migration
FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM  NEW.*)
EXECUTE PROCEDURE public.tf_set_updated_at();

CREATE TRIGGER add_audit
BEFORE UPDATE OR DELETE OR INSERT ON public.t_migration
FOR EACH ROW
EXECUTE PROCEDURE public.tf_add_audit();

ALTER TABLE public.t_migration OWNER TO lab;

GRANT SELECT ON TABLE public.t_migration TO iam;

REVOKE ALL ON TABLE public.t_migration FROM public;

CREATE TABLE public.t_outgoing_message (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    sent_at TIMESTAMP,
    error TEXT,
    event TEXT NOT NULL,
    queue TEXT NOT NULL,
    payload JSONB NOT NULL
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON public.t_outgoing_message
FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM  NEW.*)
EXECUTE PROCEDURE public.tf_set_updated_at();

CREATE TRIGGER add_audit
BEFORE UPDATE OR DELETE OR INSERT ON public.t_outgoing_message
FOR EACH ROW
EXECUTE PROCEDURE public.tf_add_audit();

ALTER TABLE public.t_outgoing_message OWNER TO lab;

GRANT SELECT, INSERT, UPDATE ON TABLE public.t_outgoing_message TO iam;

REVOKE ALL ON TABLE public.t_outgoing_message FROM public;

CREATE TABLE public.t_user (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
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

CREATE TRIGGER add_audit
BEFORE UPDATE OR DELETE OR INSERT ON public.t_user
FOR EACH ROW
EXECUTE PROCEDURE public.tf_add_audit();

ALTER TABLE public.t_user OWNER TO lab;

GRANT SELECT, UPDATE, INSERT ON TABLE public.t_user TO iam;

REVOKE ALL ON TABLE public.t_user FROM public;

CREATE TABLE public.t_session (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL DEFAULT NOW() + '90 days'::INTERVAL,
    user_id UUID NOT NULL REFERENCES public.t_user(id),
    user_agent TEXT NOT NULL CHECK(char_length(user_agent) BETWEEN 5 AND 256),
    login_ip INET NOT NULL,
    login_location TEXT CHECK(char_length(login_location) BETWEEN 2 AND 128)
);

CREATE TRIGGER add_audit
BEFORE UPDATE OR DELETE OR INSERT ON public.t_session
FOR EACH ROW
EXECUTE PROCEDURE public.tf_add_audit();

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON public.t_session
FOR EACH ROW WHEN (OLD.* IS DISTINCT FROM  NEW.*)
EXECUTE PROCEDURE public.tf_set_updated_at();

ALTER TABLE public.t_session OWNER TO lab;

GRANT SELECT, UPDATE, INSERT ON TABLE public.t_session TO iam;

REVOKE ALL ON TABLE public.t_session FROM public;

SET application.user_id TO 'migration';

-- plain password: admin123
INSERT INTO public.t_user(name, username, password) VALUES ('admin', 'admin', '$2a$10$lYazTR27nzKt5Z5HIX.lz.RNKk5t1U6PciMJS/Ig72odgh0WsBk3m');
INSERT INTO public.t_migration (name) VALUES ('0000');
