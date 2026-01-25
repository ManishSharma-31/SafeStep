package engine

import (
        "database/sql"
        "encoding/json"
        "sync"

        _ "modernc.org/sqlite"
)

type StepStatus string

const (
        StatusRunning   StepStatus = "RUNNING"
        StatusCompleted StepStatus = "COMPLETED"
        StatusFailed    StepStatus = "FAILED"
)

type StepRecord struct {
        WorkflowID string
        StepKey    string
        Status     StepStatus
        Output     string
}

type PersistenceLayer struct {
        db *sql.DB
        mu sync.Mutex
}

func NewPersistence(dbPath string) (*PersistenceLayer, error) {
        db, err := sql.Open("sqlite", dbPath)
        if err != nil {
                return nil, err
        }

        _, err = db.Exec(`
                CREATE TABLE IF NOT EXISTS workflows (
                        workflow_id TEXT PRIMARY KEY,
                        status TEXT NOT NULL,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );

                CREATE TABLE IF NOT EXISTS steps (
                        workflow_id TEXT NOT NULL,
                        step_key TEXT NOT NULL,
                        status TEXT NOT NULL,
                        output TEXT,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        PRIMARY KEY (workflow_id, step_key)
                );
        `)
        if err != nil {
                return nil, err
        }

        return &PersistenceLayer{db: db}, nil
}

func (p *PersistenceLayer) CreateWorkflow(workflowID string) error {
        p.mu.Lock()
        defer p.mu.Unlock()

        _, err := p.db.Exec(
                "INSERT OR IGNORE INTO workflows (workflow_id, status) VALUES (?, ?)",
                workflowID, StatusRunning,
        )
        return err
}

func (p *PersistenceLayer) GetStep(workflowID, stepKey string) (*StepRecord, error) {
        p.mu.Lock()
        defer p.mu.Unlock()

        var record StepRecord
        err := p.db.QueryRow(
                "SELECT workflow_id, step_key, status, COALESCE(output, '') FROM steps WHERE workflow_id = ? AND step_key = ?",
                workflowID, stepKey,
        ).Scan(&record.WorkflowID, &record.StepKey, &record.Status, &record.Output)

        if err == sql.ErrNoRows {
                return nil, nil
        }
        return &record, err
}

func (p *PersistenceLayer) SaveStepStart(workflowID, stepKey string) error {
        p.mu.Lock()
        defer p.mu.Unlock()

        _, err := p.db.Exec(
                "INSERT OR REPLACE INTO steps (workflow_id, step_key, status) VALUES (?, ?, ?)",
                workflowID, stepKey, StatusRunning,
        )
        return err
}

func (p *PersistenceLayer) SaveStepComplete(workflowID, stepKey string, output interface{}) error {
        p.mu.Lock()
        defer p.mu.Unlock()

        outputJSON, err := json.Marshal(output)
        if err != nil {
                return err
        }

        _, err = p.db.Exec(
                "UPDATE steps SET status = ?, output = ?, updated_at = CURRENT_TIMESTAMP WHERE workflow_id = ? AND step_key = ?",
                StatusCompleted, string(outputJSON), workflowID, stepKey,
        )
        return err
}

func (p *PersistenceLayer) SaveStepFailed(workflowID, stepKey string, errMsg string) error {
        p.mu.Lock()
        defer p.mu.Unlock()

        _, err := p.db.Exec(
                "UPDATE steps SET status = ?, output = ?, updated_at = CURRENT_TIMESTAMP WHERE workflow_id = ? AND step_key = ?",
                StatusFailed, errMsg, workflowID, stepKey,
        )
        return err
}

func (p *PersistenceLayer) Close() error {
        return p.db.Close()
}
