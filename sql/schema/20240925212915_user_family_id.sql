-- +goose Up
ALTER TABLE IF EXISTS public.users
    ADD COLUMN family_id bigserial;
ALTER TABLE IF EXISTS public.users
    ADD CONSTRAINT family_id_fkey FOREIGN KEY (family_id)
    REFERENCES public.families (id) MATCH SIMPLE
    ON UPDATE CASCADE
    ON DELETE CASCADE
    NOT VALID;

-- +goose Down
ALTER TABLE IF EXISTS public.users
    DROP COLUMN IF EXISTS family_id;
