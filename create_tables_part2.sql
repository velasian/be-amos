CREATE TABLE IF NOT EXISTS announcements (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    file_url VARCHAR(255),
    file_type VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_announcements_deleted_at ON announcements(deleted_at);

CREATE TABLE IF NOT EXISTS payslip_stagings (
    id SERIAL PRIMARY KEY,
    import_batch_id VARCHAR(50),
    status VARCHAR(20) DEFAULT 'PENDING',
    error_message TEXT,
    filename VARCHAR(255),
    nrp_raw VARCHAR(100),
    employee_id INTEGER,
    employee_nrp VARCHAR(100),
    employee_name VARCHAR(255),
    month INTEGER,
    year INTEGER,
    file_path VARCHAR(500),
    content_type VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_payslip_stagings_import_batch_id ON payslip_stagings(import_batch_id);
CREATE INDEX IF NOT EXISTS idx_payslip_stagings_status ON payslip_stagings(status);
