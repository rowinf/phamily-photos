-- +goose Up

CREATE SEQUENCE IF NOT EXISTS public.families_id_seq
    INCREMENT 1
    START 1
    MINVALUE 1
    MAXVALUE 2147483647
    CACHE 1;

CREATE TABLE IF NOT EXISTS public.families
(
    id bigint NOT NULL DEFAULT nextval('families_id_seq'::regclass),
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    name text COLLATE pg_catalog."default" NOT NULL,
    description text COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT families_pkey PRIMARY KEY (id)
);

-- +goose Down
DROP TABLE families;