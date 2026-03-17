CREATE TABLE usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    metric VARCHAR(100) NOT NULL,
    value INT NOT NULL DEFAULT 0,
    metric_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, metric, metric_date)
);
CREATE INDEX idx_usage_metrics_org_date ON usage_metrics(org_id, metric_date);
