CREATE TABLE policy_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    base_credential_type VARCHAR(50),
    parameters JSONB NOT NULL,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_policy_templates_project_name_ci
    ON policy_templates (project_id, lower(name));
CREATE INDEX idx_policy_templates_project_id ON policy_templates (project_id);

CREATE TABLE policy_template_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES policy_templates(id) ON DELETE CASCADE,
    version INT NOT NULL CHECK (version >= 1),
    parameters JSONB NOT NULL,
    change_note TEXT NOT NULL DEFAULT '',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (template_id, version)
);

CREATE INDEX idx_policy_template_versions_template_id
    ON policy_template_versions (template_id, version DESC);

CREATE TABLE secret_policy_bindings (
    secret_id UUID PRIMARY KEY REFERENCES secrets(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES policy_templates(id),
    template_version INT NOT NULL,
    override_parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    override_warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_by UUID REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_secret_policy_template_version
        FOREIGN KEY (template_id, template_version)
            REFERENCES policy_template_versions(template_id, version)
            ON DELETE RESTRICT
);

CREATE INDEX idx_secret_policy_bindings_template_id
    ON secret_policy_bindings (template_id, template_version);

CREATE TABLE secret_policy_agent_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agent_identities(id) ON DELETE CASCADE,
    can_read_effective_policy BOOLEAN NOT NULL DEFAULT TRUE,
    granted_by_user_id UUID NOT NULL REFERENCES users(id),
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_secret_policy_agent_permissions_active
    ON secret_policy_agent_permissions (secret_id, agent_id)
    WHERE revoked_at IS NULL;
CREATE INDEX idx_secret_policy_agent_permissions_secret_id
    ON secret_policy_agent_permissions (secret_id);

ALTER TABLE access_requests
    ADD COLUMN applied_policy JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN applied_template_id UUID REFERENCES policy_templates(id) ON DELETE SET NULL,
    ADD COLUMN applied_template_version INT,
    ADD COLUMN applied_policy_source VARCHAR(32) NOT NULL DEFAULT 'default',
    ADD COLUMN applied_policy_warnings JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD CONSTRAINT chk_access_requests_applied_policy_source
        CHECK (applied_policy_source IN ('default', 'template', 'template+override')),
    ADD CONSTRAINT fk_access_requests_applied_template_version
        FOREIGN KEY (applied_template_id, applied_template_version)
            REFERENCES policy_template_versions(template_id, version)
            ON DELETE SET NULL;
