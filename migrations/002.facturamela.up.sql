-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- Organizations table (assuming it exists, adding for reference)
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Providers table with enhanced structure
CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    provider_code TEXT, -- Optional code for internal reference
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT providers_name_org_unique UNIQUE (organization_id, name)
);

-- Projects table with enhanced structure
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT projects_name_org_unique UNIQUE (organization_id, name)
);

-- Junction table for project-provider relationships
CREATE TABLE project_providers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role TEXT, -- Optional role description
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT project_providers_unique UNIQUE (project_id, provider_id),
    -- Ensure organization consistency
    CONSTRAINT project_providers_org_consistency 
        CHECK (organization_id IS NOT NULL)
);

-- Invoice types with comprehensive schema validation
CREATE TABLE invoice_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_type TEXT NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    
    -- JSONB schema with validation
    invoice_schema JSONB NOT NULL,
    
    -- Schema metadata
    schema_version TEXT NOT NULL DEFAULT '1.0',
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    -- Audit fields
    created_by UUID, -- Could reference users table
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Unique constraints
    CONSTRAINT invoice_types_type_org_unique UNIQUE (organization_id, invoice_type),
    
    -- JSONB validation constraints
    CONSTRAINT invoice_schema_not_empty CHECK (jsonb_typeof(invoice_schema) = 'object'),
    CONSTRAINT invoice_schema_has_required_fields CHECK (
        invoice_schema ? 'fields' AND 
        jsonb_typeof(invoice_schema->'fields') = 'array'
    )
);

-- Main invoices table with comprehensive optimization
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Core invoice data
    invoice_data JSONB NOT NULL,
    invoice_type_id UUID NOT NULL REFERENCES invoice_types(id) ON DELETE RESTRICT,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    provider_id UUID REFERENCES providers(id) ON DELETE SET NULL,
    
    -- Business fields extracted from JSONB for querying performance
    invoice_number TEXT, -- Extracted from JSONB for indexing
    invoice_date DATE,   -- Extracted from JSONB for indexing
    due_date DATE,
    total_amount DECIMAL(15,2), -- Extracted from JSONB for indexing
    currency_code CHAR(3), -- ISO currency code
    status TEXT, -- Extracted from JSONB, defined by user schema
    
    -- Audit and versioning
    version INTEGER NOT NULL DEFAULT 1,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    created_by UUID, -- Could reference users table
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    
    -- Constraints
    CONSTRAINT invoices_invoice_data_not_empty CHECK (jsonb_typeof(invoice_data) = 'object'),
    CONSTRAINT invoices_amount_positive CHECK (total_amount IS NULL OR total_amount >= 0),
    CONSTRAINT invoices_due_date_logical CHECK (due_date IS NULL OR invoice_date IS NULL OR due_date >= invoice_date),
    CONSTRAINT invoices_currency_valid CHECK (currency_code IS NULL OR length(currency_code) = 3),
    CONSTRAINT invoices_org_consistency CHECK (organization_id IS NOT NULL),
    
    -- Unique invoice number per organization
    CONSTRAINT invoices_number_org_unique UNIQUE (organization_id, invoice_number) 
        DEFERRABLE INITIALLY DEFERRED
);

-- Performance indexes
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_providers_organization_active 
    ON providers(organization_id, is_active) WHERE is_active = true;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_projects_organization_active 
    ON projects(organization_id, is_active) WHERE is_active = true;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoice_types_organization_active 
    ON invoice_types(organization_id, is_active) WHERE is_active = true;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_organization_status 
    ON invoices(organization_id, status) WHERE is_deleted = false;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_dates 
    ON invoices(invoice_date, due_date) WHERE is_deleted = false;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_amount 
    ON invoices(total_amount) WHERE is_deleted = false AND total_amount IS NOT NULL;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_number 
    ON invoices(invoice_number) WHERE is_deleted = false AND invoice_number IS NOT NULL;

-- Composite indexes for common query patterns
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_org_status_date 
    ON invoices(organization_id, status, invoice_date DESC) 
    WHERE is_deleted = false;

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_invoices_org_type_status 
    ON invoices(organization_id, invoice_type_id, status) 
    WHERE is_deleted = false;

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for automatic timestamp updates
CREATE TRIGGER trigger_organizations_updated_at 
    BEFORE UPDATE ON organizations 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_providers_updated_at 
    BEFORE UPDATE ON providers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_projects_updated_at 
    BEFORE UPDATE ON projects 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_project_providers_updated_at 
    BEFORE UPDATE ON project_providers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_invoice_types_updated_at 
    BEFORE UPDATE ON invoice_types 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_invoices_updated_at 
    BEFORE UPDATE ON invoices 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Function to extract and sync key fields from invoice JSONB
CREATE OR REPLACE FUNCTION sync_invoice_fields()
RETURNS TRIGGER AS $$
BEGIN
    -- Extract key fields from JSONB for indexing and querying
    NEW.invoice_number := NEW.invoice_data->>'invoice_number';
    NEW.invoice_date := (NEW.invoice_data->>'invoice_date')::DATE;
    NEW.due_date := (NEW.invoice_data->>'due_date')::DATE;
    NEW.total_amount := (NEW.invoice_data->>'total_amount')::DECIMAL(15,2);
    NEW.currency_code := NEW.invoice_data->>'currency_code';
    NEW.status := NEW.invoice_data->>'status';
    
    -- Increment version on updates
    IF TG_OP = 'UPDATE' AND OLD.invoice_data IS DISTINCT FROM NEW.invoice_data THEN
        NEW.version := OLD.version + 1;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically sync extracted fields
CREATE TRIGGER trigger_invoices_sync_fields 
    BEFORE INSERT OR UPDATE ON invoices 
    FOR EACH ROW EXECUTE FUNCTION sync_invoice_fields();

-- Function for soft delete
CREATE OR REPLACE FUNCTION soft_delete_invoice(invoice_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE invoices 
    SET is_deleted = true, deleted_at = NOW() 
    WHERE id = invoice_id AND is_deleted = false;
    
    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Views for common queries
CREATE VIEW active_invoices AS
SELECT 
    i.*,
    it.invoice_type,
    it.invoice_schema,
    o.name as organization_name,
    p.name as project_name,
    pr.name as provider_name
FROM invoices i
JOIN invoice_types it ON i.invoice_type_id = it.id
JOIN organizations o ON i.organization_id = o.id
LEFT JOIN projects p ON i.project_id = p.id
LEFT JOIN providers pr ON i.provider_id = pr.id
WHERE i.is_deleted = false;

-- View for invoice analytics
CREATE VIEW invoice_analytics AS
SELECT 
    organization_id,
    status,
    currency_code,
    COUNT(*) as invoice_count,
    SUM(total_amount) as total_amount,
    AVG(total_amount) as avg_amount,
    MIN(invoice_date) as earliest_invoice,
    MAX(invoice_date) as latest_invoice
FROM invoices 
WHERE is_deleted = false 
  AND status IS NOT NULL
GROUP BY organization_id, status, currency_code;

-- Row Level Security (RLS) setup
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoice_types ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE providers ENABLE ROW LEVEL SECURITY;
ALTER TABLE project_providers ENABLE ROW LEVEL SECURITY;

-- Example RLS policy (adjust based on your auth system)
-- CREATE POLICY invoices_organization_isolation ON invoices
--     FOR ALL TO authenticated_users
--     USING (organization_id = current_setting('app.current_organization_id')::UUID);

-- Useful utility functions for JSONB operations
CREATE OR REPLACE FUNCTION validate_invoice_schema(
    invoice_data JSONB, 
    schema_data JSONB
) RETURNS BOOLEAN AS $$
-- This is a placeholder - implement your specific schema validation logic
-- You might want to use a proper JSON Schema validator
BEGIN
    -- Basic validation example
    RETURN invoice_data ? 'invoice_number' 
       AND invoice_data ? 'total_amount'
       AND invoice_data ? 'currency_code';
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE invoices IS 'Core invoice storage with JSONB flexibility and extracted key fields for performance';
COMMENT ON COLUMN invoices.invoice_data IS 'Complete invoice data in JSONB format, schema defined by invoice_type';
COMMENT ON COLUMN invoices.invoice_number IS 'Extracted invoice number for indexing and uniqueness';
COMMENT ON COLUMN invoices.total_amount IS 'Extracted total amount for aggregations and filtering';
COMMENT ON COLUMN invoices.status IS 'Extracted status from JSONB, values defined by user schema';
COMMENT ON COLUMN invoices.version IS 'Optimistic locking version, incremented on each update';

COMMENT ON TABLE invoice_types IS 'Defines invoice schemas and types per organization/project';
COMMENT ON COLUMN invoice_types.invoice_schema IS 'JSON Schema defining the structure and validation rules for invoices';

-- Performance monitoring query
/*
-- Use this to monitor query performance:
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE tablename IN ('invoices', 'invoice_types', 'projects', 'providers')
ORDER BY tablename, attname;
*/
