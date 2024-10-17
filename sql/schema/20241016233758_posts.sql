
-- +goose Up

CREATE TABLE public.posts
(
    id bigserial,
    created_at timestamp without time zone NOT NULL,
    updated_at timestamp without time zone NOT NULL,
    description text NOT NULL,
    featured_photo_id text,
    user_id text NOT NULL,
    family_id bigserial NOT NULL,
    PRIMARY KEY (id)
);

ALTER TABLE IF EXISTS public.posts
    ADD CONSTRAINT user_id_fkey FOREIGN KEY (user_id)
    REFERENCES public.users (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

ALTER TABLE IF EXISTS public.posts
    ADD CONSTRAINT family_id_fkey FOREIGN KEY (family_id)
    REFERENCES public.families (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

ALTER TABLE IF EXISTS public.posts
    ADD CONSTRAINT featured_photo_id_fkey FOREIGN KEY (featured_photo_id)
    REFERENCES public.photos (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

ALTER TABLE IF EXISTS public.photos
    ADD COLUMN post_id bigserial;

ALTER TABLE IF EXISTS public.photos
    ADD CONSTRAINT post_id_fkey FOREIGN KEY (post_id)
    REFERENCES public.posts (id) MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;

-- +goose Down
DROP TABLE public.posts;

ALTER TABLE IF EXISTS public.photos
    DROP COLUMN IF EXISTS post_id;
