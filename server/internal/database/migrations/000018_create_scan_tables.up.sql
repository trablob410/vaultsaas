CREATE TABLE scan_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scanned_by UUID NOT NULL REFERENCES users(id),
    scan_path TEXT NOT NULL,
    findings_count INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_scan_results_project ON scan_results(project_id);

CREATE TABLE scan_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id UUID NOT NULL REFERENCES scan_results(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    line_number INT NOT NULL,
    pattern_name VARCHAR(100) NOT NULL,
    credential_type VARCHAR(100) NOT NULL,
    redacted_preview TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    imported_secret_id UUID REFERENCES secrets(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_scan_findings_scan ON scan_findings(scan_id);
