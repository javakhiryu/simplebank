CREATE TABLE "sessions" (
    "id" uuid PRIMARY KEY,
    "refresh_token" varchar NOT NULL,
    "username" varchar NOT NULL,
    "user_agent" varchar NOT NULL,
    "ip_address" varchar NOT NULL,
    "expires_at" timestamptz NOT NULL NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "is_blocked" boolean NOT NULL DEFAULT false
);

ALTER TABLE "sessions" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");