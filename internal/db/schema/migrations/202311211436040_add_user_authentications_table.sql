-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_authentications
(
    connection_id                     uuid NOT NULL,
    session_id                        uuid NOT NULL,
    created_at                        timestamptz NOT NULL,
    CONSTRAINT user_authentications_session_connection_key UNIQUE (connection_id, session_id),
    CONSTRAINT fk_user_authentications_connection_id FOREIGN KEY (connection_id) REFERENCES public.connections(id)
);

CREATE INDEX user_authentications_session_id_idx ON user_authentications(session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_authentications;
DROP INDEX IF EXISTS user_authentications_session_id_idx;
-- +goose StatementEnd
