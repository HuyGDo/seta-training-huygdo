-- Insert Users
INSERT INTO "Users" ("userId", username, email, password, role, "createdAt", "updatedAt") VALUES
('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1', 'Alice', 'alice@example.com', '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MANAGER', NOW(), NOW()),
('b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2', 'Bob', 'bob@example.com', '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MANAGER', NOW(), NOW()),
('c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 'Carol', 'carol@example.com', '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER', NOW(), NOW()),
('d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4', 'Dave', 'dave@example.com', '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER', NOW(), NOW()),
('e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5', 'Eve', 'eve@example.com', '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER', NOW(), NOW());