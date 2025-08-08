CREATE TABLE IF NOT EXISTS invoices (
 id UUID,
 invoice_data JSONB,
 invoice_schema JSONB,
 invoice_type TEXT,
 organization_id UUID references organizations(id),

);
