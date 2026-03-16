-- Create guest user for widget bookings
INSERT INTO users (id, email, password_hash, name, phone, role, is_active, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'widget-guest@bani.local',
    '',
    'Widget Guest',
    '',
    'client',
    true,
    now(),
    now()
) ON CONFLICT (id) DO NOTHING;
