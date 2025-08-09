-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create update trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ============================================================================
-- USERS PACKAGE
-- ============================================================================

-- Users table
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- User profiles table
CREATE TABLE user_profiles (
    user_id VARCHAR(36) PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    profile_picture TEXT,
    bio TEXT DEFAULT '',
    phone VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- ============================================================================
-- ORGANIZATIONS PACKAGE
-- ============================================================================

-- Organizations table
CREATE TABLE organizations (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    owner_user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Organization memberships table
CREATE TABLE organization_memberships (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id VARCHAR(36) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    role_name VARCHAR(50) NOT NULL,
    role_permissions TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Ensure one membership per user per organization
    UNIQUE(user_id, organization_id)
);

-- Organization invitations table
CREATE TABLE organization_invitations (
    id VARCHAR(36) PRIMARY KEY,
    organization_id VARCHAR(36) NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    inviter_user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role_name VARCHAR(50) NOT NULL,
    role_permissions TEXT[] NOT NULL DEFAULT '{}',
    token VARCHAR(255) UNIQUE NOT NULL,
    is_used BOOLEAN DEFAULT FALSE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- ============================================================================
-- AUTH PACKAGE
-- ============================================================================

-- OAuth accounts table
CREATE TABLE oauth_accounts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255) NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    -- Ensure unique provider account per user
    UNIQUE(provider, provider_id)
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Users indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_users_created_at ON users(created_at);

-- User profiles indexes
CREATE INDEX idx_user_profiles_name ON user_profiles(first_name, last_name);

-- Organizations indexes
CREATE INDEX idx_organizations_owner_user_id ON organizations(owner_user_id);
CREATE INDEX idx_organizations_is_active ON organizations(is_active);
CREATE INDEX idx_organizations_name ON organizations(name);
CREATE INDEX idx_organizations_created_at ON organizations(created_at);

-- Organization memberships indexes
CREATE INDEX idx_memberships_user_id ON organization_memberships(user_id);
CREATE INDEX idx_memberships_organization_id ON organization_memberships(organization_id);
CREATE INDEX idx_memberships_role_name ON organization_memberships(role_name);
CREATE INDEX idx_memberships_is_active ON organization_memberships(is_active);
CREATE INDEX idx_memberships_user_org_active ON organization_memberships(user_id, organization_id, is_active);
CREATE INDEX idx_memberships_org_active_role ON organization_memberships(organization_id, is_active, role_name);

-- Organization invitations indexes
CREATE INDEX idx_invitations_organization_id ON organization_invitations(organization_id);
CREATE INDEX idx_invitations_email ON organization_invitations(email);
CREATE INDEX idx_invitations_token ON organization_invitations(token);
CREATE INDEX idx_invitations_expires_at ON organization_invitations(expires_at);
CREATE INDEX idx_invitations_is_used ON organization_invitations(is_used);
CREATE INDEX idx_invitations_active ON organization_invitations(is_used, expires_at) WHERE is_used = FALSE;

-- OAuth accounts indexes
CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);
CREATE INDEX idx_oauth_accounts_provider ON oauth_accounts(provider);
CREATE INDEX idx_oauth_accounts_provider_id ON oauth_accounts(provider_id);
CREATE INDEX idx_oauth_accounts_provider_email ON oauth_accounts(provider_email);
CREATE INDEX idx_oauth_accounts_expires_at ON oauth_accounts(expires_at);
CREATE INDEX idx_oauth_accounts_user_provider ON oauth_accounts(user_id, provider);

-- GIN index for metadata JSONB queries
CREATE INDEX idx_oauth_accounts_metadata ON oauth_accounts USING GIN(metadata);

-- ============================================================================
-- CONSTRAINTS
-- ============================================================================

-- Email format validation for invitations
ALTER TABLE organization_invitations 
ADD CONSTRAINT chk_invitation_email_format 
CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$');

-- OAuth provider validation
ALTER TABLE oauth_accounts 
ADD CONSTRAINT chk_oauth_provider 
CHECK (provider IN ('google', 'github', 'facebook', 'linkedin', 'microsoft'));

-- Name length validations
ALTER TABLE organizations 
ADD CONSTRAINT chk_org_name_length 
CHECK (length(trim(name)) >= 2);

ALTER TABLE users 
ADD CONSTRAINT chk_user_name_length 
CHECK (length(trim(name)) >= 1);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Update triggers for all tables
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_profiles_updated_at 
    BEFORE UPDATE ON user_profiles 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organizations_updated_at 
    BEFORE UPDATE ON organizations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_memberships_updated_at 
    BEFORE UPDATE ON organization_memberships 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_oauth_accounts_updated_at 
    BEFORE UPDATE ON oauth_accounts 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View to get user with organization memberships
CREATE VIEW user_organization_memberships AS
SELECT 
    u.id as user_id,
    u.email,
    u.name as user_name,
    u.is_active as user_active,
    o.id as organization_id,
    o.name as organization_name,
    o.description as organization_description,
    om.id as membership_id,
    om.role_name,
    om.role_permissions,
    om.is_active as membership_active,
    om.joined_at
FROM users u
    LEFT JOIN organization_memberships om ON u.id = om.user_id
    LEFT JOIN organizations o ON om.organization_id = o.id
WHERE u.is_active = true;

-- View for active invitations
CREATE VIEW active_invitations AS
SELECT 
    i.*,
    o.name as organization_name,
    u.name as inviter_name
FROM organization_invitations i
    JOIN organizations o ON i.organization_id = o.id
    JOIN users u ON i.inviter_user_id = u.id
WHERE 
    i.is_used = false 
    AND i.expires_at > NOW()
    AND o.is_active = true;

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Function to get user's active organizations
CREATE OR REPLACE FUNCTION get_user_organizations(user_uuid VARCHAR(36))
RETURNS TABLE (
    organization_id VARCHAR(36),
    organization_name VARCHAR(255),
    role_name VARCHAR(50),
    membership_id VARCHAR(36)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        o.id,
        o.name,
        om.role_name,
        om.id
    FROM organizations o
        JOIN organization_memberships om ON o.id = om.organization_id
    WHERE 
        om.user_id = user_uuid
        AND om.is_active = true
        AND o.is_active = true
    ORDER BY om.joined_at;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

-- Table comments
COMMENT ON TABLE users IS 'User accounts in the system';
COMMENT ON TABLE user_profiles IS 'Extended user profile information';
COMMENT ON TABLE organizations IS 'Organizations that users can belong to';
COMMENT ON TABLE organization_memberships IS 'User memberships in organizations with roles';
COMMENT ON TABLE organization_invitations IS 'Pending invitations to join organizations';
COMMENT ON TABLE oauth_accounts IS 'OAuth provider account linkages';

-- Column comments
COMMENT ON COLUMN organization_memberships.role_name IS 'Role names: org_admin, org_member, product_provider';
COMMENT ON COLUMN organization_memberships.role_permissions IS 'Array of permission strings';
COMMENT ON COLUMN organization_invitations.role_name IS 'Role names: org_admin, org_member, product_provider';
COMMENT ON COLUMN organization_invitations.role_permissions IS 'Array of permission strings';
COMMENT ON COLUMN oauth_accounts.metadata IS 'Additional provider-specific user data';
COMMENT ON COLUMN oauth_accounts.provider IS 'OAuth provider name (google, github, etc.)';

-- ============================================================================
-- SAMPLE DATA (Optional - for development)
-- ============================================================================

-- Uncomment below for development/testing data
/*
-- Sample users
INSERT INTO users (id, email, name) VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'admin@example.com', 'Admin User'),
    ('550e8400-e29b-41d4-a716-446655440002', 'member@example.com', 'Member User'),
    ('550e8400-e29b-41d4-a716-446655440003', 'provider@example.com', 'Provider User');

-- Sample user profiles
INSERT INTO user_profiles (user_id, first_name, last_name, bio) VALUES 
    ('550e8400-e29b-41d4-a716-446655440001', 'Admin', 'User', 'System administrator'),
    ('550e8400-e29b-41d4-a716-446655440002', 'Member', 'User', 'Regular member'),
    ('550e8400-e29b-41d4-a716-446655440003', 'Provider', 'User', 'Product provider');

-- Sample organization
INSERT INTO organizations (id, name, description, owner_user_id) VALUES 
    ('550e8400-e29b-41d4-a716-446655440101', 'Example Corp', 'A sample organization for testing', '550e8400-e29b-41d4-a716-446655440001');

-- Sample memberships
INSERT INTO organization_memberships (id, user_id, organization_id, role_name, role_permissions) VALUES 
    ('550e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440101', 'org_admin', 
     ARRAY['org.view', 'org.manage', 'org.invite_members', 'org.remove_members', 'products.view', 'members.view', 'members.manage']),
    ('550e8400-e29b-41d4-a716-446655440202', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440101', 'org_member', 
     ARRAY['org.view', 'products.view', 'members.view']),
    ('550e8400-e29b-41d4-a716-446655440203', '550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440101', 'product_provider', 
     ARRAY['org.view', 'products.view', 'products.create', 'products.manage', 'products.delete']);
*/
