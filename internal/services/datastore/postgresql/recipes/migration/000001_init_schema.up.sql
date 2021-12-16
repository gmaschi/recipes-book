CREATE TABLE "authors" (
                           "username" varchar PRIMARY KEY,
                           "hashed_password" varchar NOT NULL,
                           "email" varchar UNIQUE NOT NULL,
                           "created_at" timestamptz NOT NULL DEFAULT (now()),
                           "updated_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "recipes" (
                           "id" bigserial PRIMARY KEY,
                           "author" varchar NOT NULL,
                           "ingredients" varchar[] NOT NULL,
                           "steps" varchar[] NOT NULL,
                           "created_at" timestamptz NOT NULL DEFAULT (now()),
                           "updated_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "recipes" ADD FOREIGN KEY ("author") REFERENCES "authors" ("username");

CREATE INDEX ON "recipes" ("author");