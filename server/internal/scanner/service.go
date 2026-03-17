package scanner

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ScanResult represents a scan run against a project directory.
type ScanResult struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	ScannedBy     string    `json:"scanned_by"`
	ScanPath      string    `json:"scan_path"`
	FindingsCount int       `json:"findings_count"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

// ScanFinding represents a single secret found during a scan.
type ScanFinding struct {
	ID               string    `json:"id"`
	ScanID           string    `json:"scan_id"`
	FilePath         string    `json:"file_path"`
	LineNumber       int       `json:"line_number"`
	PatternName      string    `json:"pattern_name"`
	CredentialType   string    `json:"credential_type"`
	RedactedPreview  string    `json:"redacted_preview"`
	Status           string    `json:"status"`
	ImportedSecretID *string   `json:"imported_secret_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// Service handles scan database operations.
type Service struct {
	db *pgxpool.Pool
}

// NewService creates a new scanner service.
func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// CreateScan inserts a scan result and its findings, returning the scan.
func (s *Service) CreateScan(ctx context.Context, projectID, scannedBy, scanPath string, findings []ScanFinding) (*ScanResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var scan ScanResult
	err = tx.QueryRow(ctx,
		`INSERT INTO scan_results (project_id, scanned_by, scan_path, findings_count)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, project_id, scanned_by, scan_path, findings_count, status, created_at`,
		projectID, scannedBy, scanPath, len(findings),
	).Scan(&scan.ID, &scan.ProjectID, &scan.ScannedBy, &scan.ScanPath,
		&scan.FindingsCount, &scan.Status, &scan.CreatedAt)
	if err != nil {
		return nil, err
	}

	for i := range findings {
		findings[i].ScanID = scan.ID
		_, err = tx.Exec(ctx,
			`INSERT INTO scan_findings (scan_id, file_path, line_number, pattern_name, credential_type, redacted_preview)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			scan.ID, findings[i].FilePath, findings[i].LineNumber,
			findings[i].PatternName, findings[i].CredentialType, findings[i].RedactedPreview,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &scan, nil
}

// ListScans returns all scans for a project.
func (s *Service) ListScans(ctx context.Context, projectID string) ([]ScanResult, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, project_id, scanned_by, scan_path, findings_count, status, created_at
		 FROM scan_results WHERE project_id = $1 ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []ScanResult
	for rows.Next() {
		var sr ScanResult
		if err := rows.Scan(&sr.ID, &sr.ProjectID, &sr.ScannedBy, &sr.ScanPath,
			&sr.FindingsCount, &sr.Status, &sr.CreatedAt); err != nil {
			return nil, err
		}
		scans = append(scans, sr)
	}
	return scans, rows.Err()
}

// ListFindings returns all findings for a scan.
func (s *Service) ListFindings(ctx context.Context, scanID string) ([]ScanFinding, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, scan_id, file_path, line_number, pattern_name, credential_type,
		        redacted_preview, status, imported_secret_id, created_at
		 FROM scan_findings WHERE scan_id = $1 ORDER BY file_path, line_number`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []ScanFinding
	for rows.Next() {
		var f ScanFinding
		if err := rows.Scan(&f.ID, &f.ScanID, &f.FilePath, &f.LineNumber,
			&f.PatternName, &f.CredentialType, &f.RedactedPreview,
			&f.Status, &f.ImportedSecretID, &f.CreatedAt); err != nil {
			return nil, err
		}
		findings = append(findings, f)
	}
	return findings, rows.Err()
}

// ImportFinding marks a finding as imported and records the secret ID.
func (s *Service) ImportFinding(ctx context.Context, findingID, secretID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE scan_findings SET status = 'imported', imported_secret_id = $1 WHERE id = $2`,
		secretID, findingID,
	)
	return err
}

// DismissFinding marks a finding as dismissed.
func (s *Service) DismissFinding(ctx context.Context, findingID string) error {
	_, err := s.db.Exec(ctx,
		`UPDATE scan_findings SET status = 'dismissed' WHERE id = $1`,
		findingID,
	)
	return err
}
