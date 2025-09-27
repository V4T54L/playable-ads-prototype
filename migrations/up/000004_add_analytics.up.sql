CREATE TYPE event_type AS ENUM ('play', 'click', 'impression');

CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type event_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON analytics_events (project_id);
CREATE INDEX ON analytics_events (user_id);
CREATE INDEX ON analytics_events (event_type);
