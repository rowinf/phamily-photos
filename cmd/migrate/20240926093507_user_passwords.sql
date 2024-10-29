-- +goose Up
-- 1. Add the column without NOT NULL constraint
ALTER TABLE public.users
    ADD COLUMN password text;

-- 2. Update existing rows to have a default value for the password column
UPDATE public.users
SET password = 'default_password';  -- Change 'default_password' as appropriate

-- 3. Alter the column to add the NOT NULL constraint now that it has valid values
ALTER TABLE public.users
    ALTER COLUMN password SET NOT NULL;

-- +goose Down
ALTER TABLE IF EXISTS public.users
   DROP COLUMN IF EXISTS password;
