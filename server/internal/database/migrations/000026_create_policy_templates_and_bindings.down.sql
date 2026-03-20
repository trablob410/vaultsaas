ALTER TABLE access_requests
    DROP CONSTRAINT IF EXISTS fk_access_requests_applied_template_version,
    DROP CONSTRAINT IF EXISTS chk_access_requests_applied_policy_source,
    DROP COLUMN IF EXISTS applied_policy_warnings,
    DROP COLUMN IF EXISTS applied_policy_source,
    DROP COLUMN IF EXISTS applied_template_version,
    DROP COLUMN IF EXISTS applied_template_id,
    DROP COLUMN IF EXISTS applied_policy;

DROP INDEX IF EXISTS idx_secret_policy_agent_permissions_secret_id;
DROP INDEX IF EXISTS idx_secret_policy_agent_permissions_active;
DROP TABLE IF EXISTS secret_policy_agent_permissions;

DROP INDEX IF EXISTS idx_secret_policy_bindings_template_id;
DROP TABLE IF EXISTS secret_policy_bindings;

DROP INDEX IF EXISTS idx_policy_template_versions_template_id;
DROP TABLE IF EXISTS policy_template_versions;

DROP INDEX IF EXISTS idx_policy_templates_project_id;
DROP INDEX IF EXISTS idx_policy_templates_project_name_ci;
DROP TABLE IF EXISTS policy_templates;
