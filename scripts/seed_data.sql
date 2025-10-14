-- Insert sample users
INSERT INTO users (id, username, email) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'john_doe', 'john@example.com'),
    ('550e8400-e29b-41d4-a716-446655440002', 'jane_smith', 'jane@example.com'),
    ('550e8400-e29b-41d4-a716-446655440003', 'bob_wilson', 'bob@example.com')
ON CONFLICT (email) DO NOTHING;

-- Insert sample board
INSERT INTO boards (id, name, description, created_by) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', 'Project Alpha', 'Main project board', '550e8400-e29b-41d4-a716-446655440001')
ON CONFLICT (id) DO NOTHING;

-- Insert sample tasks
INSERT INTO tasks (board_id, title, description, status, priority, assigned_to, created_by, position) VALUES
    ('660e8400-e29b-41d4-a716-446655440001', 'Setup development environment', 'Install all necessary tools and dependencies', 'done', 'high', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001', 1),
    ('660e8400-e29b-41d4-a716-446655440001', 'Design database schema', 'Create ERD and database tables', 'in_progress', 'high', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 2),
    ('660e8400-e29b-41d4-a716-446655440001', 'Implement WebSocket handler', 'Build real-time communication layer', 'todo', 'medium', '550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440001', 3),
    ('660e8400-e29b-41d4-a716-446655440001', 'Create frontend dashboard', 'Build React/Vue dashboard for task management', 'todo', 'medium', NULL, '550e8400-e29b-41d4-a716-446655440001', 4),
    ('660e8400-e29b-41d4-a716-446655440001', 'Write unit tests', 'Achieve 80%+ test coverage', 'todo', 'low', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001', 5);