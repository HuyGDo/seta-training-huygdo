-- Enable UUID generation function if not already enabled
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =================================================================
-- Table: teams
-- Stores basic information about teams.
-- =================================================================
CREATE TABLE teams (
    team_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =================================================================
-- Mapping Table: team_managers
-- Links users with the 'manager' role to teams.
-- =================================================================
CREATE TABLE team_managers (
    team_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (team_id, user_id),
    FOREIGN KEY (team_id) REFERENCES teams(team_id) ON DELETE CASCADE
);

-- =================================================================
-- Mapping Table: team_members
-- Links users with the 'member' role to teams.
-- =================================================================
CREATE TABLE team_members (
    team_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (team_id, user_id),
    FOREIGN KEY (team_id) REFERENCES teams(team_id) ON DELETE CASCADE
);

-- =================================================================
-- Table: folders
-- Stores folder information, linked to an owner.
-- =================================================================
CREATE TABLE folders (
    folder_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    folder_name VARCHAR(255) NOT NULL,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add an index on owner_id for faster lookups of a user's folders
CREATE INDEX idx_folders_owner_id ON folders(owner_id);

-- =================================================================
-- Table: notes
-- Stores note content, linked to a folder and an owner.
-- =================================================================
CREATE TABLE notes (
    note_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    body TEXT,
    folder_id UUID NOT NULL,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (folder_id) REFERENCES folders(folder_id) ON DELETE CASCADE
);

-- Add indexes for faster lookups
CREATE INDEX idx_notes_folder_id ON notes(folder_id);
CREATE INDEX idx_notes_owner_id ON notes(owner_id);

-- =================================================================
-- Sharing Table: folder_shares
-- Manages access permissions for folders shared with other users.
-- =================================================================
CREATE TABLE folder_shares (
    folder_id UUID NOT NULL,
    user_id UUID NOT NULL, -- The user the folder is shared WITH
    access_level VARCHAR(10) NOT NULL CHECK (access_level IN ('read', 'write')),
    PRIMARY KEY (folder_id, user_id),
    FOREIGN KEY (folder_id) REFERENCES folders(folder_id) ON DELETE CASCADE
);

-- =================================================================
-- Sharing Table: note_shares
-- Manages access permissions for individual notes shared with others.
-- =================================================================
CREATE TABLE note_shares (
    note_id UUID NOT NULL,
    user_id UUID NOT NULL, -- The user the note is shared WITH
    access_level VARCHAR(10) NOT NULL CHECK (access_level IN ('read', 'write')),
    PRIMARY KEY (note_id, user_id),
    FOREIGN KEY (note_id) REFERENCES notes(note_id) ON DELETE CASCADE
);


-- =================================================================
-- =================================================================
-- MOCK DATA INSERTION
-- =================================================================
-- =================================================================

-- Declare variables for IDs to make linking easier
DO $$
DECLARE
    -- User IDs
    manager_alice_id UUID := 'a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1';
    manager_bob_id UUID   := 'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2';
    member_carol_id UUID  := 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3';
    member_dave_id UUID   := 'd4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4';
    member_eve_id UUID    := 'e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5';

    -- Team IDs
    team_eng_id UUID      := 'f1f1f1f1-f1f1-f1f1-f1f1-f1f1f1f1f1f1';
    team_mkt_id UUID      := 'a2a2a2a2-a2a2-a2a2-a2a2-a2a2a2a2a2a2'; -- Corrected from 'g'

    -- Asset IDs
    folder_alice_id UUID  := 'b3b3b3b3-b3b3-b3b3-b3b3-b3b3b3b3b3b3'; -- Corrected from 'h'
    folder_carol_id UUID  := 'c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4'; -- Corrected from 'i'
    note_alice_id UUID    := 'd5d5d5d5-d5d5-d5d5-d5d5-d5d5d5d5d5d5'; -- Corrected from 'j'
    note_carol_id UUID    := 'e6e6e6e6-e6e6-e6e6-e6e6-e6e6e6e6e6e6'; -- Corrected from 'k'
BEGIN

-- Insert Teams
INSERT INTO teams (team_id, team_name) VALUES
(team_eng_id, 'Engineering'),
(team_mkt_id, 'Marketing');

-- Assign Managers to Teams
INSERT INTO team_managers (team_id, user_id) VALUES
(team_eng_id, manager_alice_id), -- Alice manages Engineering
(team_mkt_id, manager_bob_id);   -- Bob manages Marketing

-- Assign Members to Teams
INSERT INTO team_members (team_id, user_id) VALUES
(team_eng_id, member_carol_id),  -- Carol is in Engineering
(team_eng_id, member_dave_id),   -- Dave is in Engineering
(team_mkt_id, member_eve_id);    -- Eve is in Marketing

-- Insert Folders
INSERT INTO folders (folder_id, folder_name, owner_id) VALUES
(folder_alice_id, 'Project Phoenix Docs', manager_alice_id),
(folder_carol_id, 'Personal Notes', member_carol_id);

-- Insert Notes
INSERT INTO notes (note_id, title, body, folder_id, owner_id) VALUES
(note_alice_id, 'Q3 Architecture Plan', 'The plan is to use microservices...', folder_alice_id, manager_alice_id),
(note_carol_id, 'Meeting Summary', 'Discussed project timelines.', folder_carol_id, member_carol_id);

-- Insert Folder and Note Shares
-- Alice shares her "Project Phoenix Docs" folder with Bob with read-only access
INSERT INTO folder_shares (folder_id, user_id, access_level) VALUES
(folder_alice_id, manager_bob_id, 'read');

-- Carol shares her "Meeting Summary" note with Dave with write access
INSERT INTO note_shares (note_id, user_id, access_level) VALUES
(note_carol_id, member_dave_id, 'write');

END $$;