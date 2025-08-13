# User Service

A standalone GraphQL service to handle user operations.

## Installation

```node
cd user-service
npm install
```

```
-- install postgresql + dbeaver
-- set the password for the built-in "postgres" user
ALTER USER postgres
  WITH ENCRYPTED PASSWORD '191102';

\c postgres

-- create the "personnel" database
CREATE DATABASE personnel
  WITH 
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8'
    TEMPLATE = template0;

-- seed sample users
-- password is "JohnDoe01@" (for all sample users)
INSERT INTO "Users"
  ("userId",                               "username",        "email",                          "password",                                                     "role",      "createdAt",   "updatedAt")
VALUES
  ('a3f47e1b-4d2c-4e5f-9a11-1a2b3c4d5e6f', 'manager_olivia',  'olivia.manager@example.com',     '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MANAGER',   NOW(),         NOW()),
  ('b2c3d4e5-6f7a-8b9c-0d1e-2f3a4b5c6d7e', 'manager_lucas',   'lucas.manager@example.com',      '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MANAGER',   NOW(),         NOW()),
  ('c3d4e5f6-7a8b-9c0d-1e2f-3a4b5c6d7e8f', 'manager_emily',   'emily.manager@example.com',      '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MANAGER',   NOW(),         NOW()),
  ('d4e5f6a7-8b9c-0d1e-2f3a-4b5c6d7e8f9a', 'manager_james',   'james.manager@example.com',      '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MANAGER',   NOW(),         NOW()),
  ('e5f6a7b8-9c0d-1e2f-3a4b-5c6d7e8f9a0b', 'member_emma',     'emma.member@example.com',        '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER',    NOW(),         NOW()),
  ('f6a7b8c9-0d1e-2f3a-4b5c-6d7e8f9a0b1c', 'member_noah',     'noah.member@example.com',        '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER',    NOW(),         NOW()),
  ('a7b8c9d0-1e2f-3a4b-5c6d-7e8f9a0b1c2d', 'member_ava',      'ava.member@example.com',         '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER',    NOW(),         NOW()),
  ('b8c9d0e1-2f3a-4b5c-6d7e-8f9a0b1c2d3e', 'member_liam',     'liam.member@example.com',        '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER',    NOW(),         NOW()),
  ('c9d0e1f2-3a4b-5c6d-7e8f-9a0b1c2d3e4f', 'member_sophia',   'sophia.member@example.com',      '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER',    NOW(),         NOW()),
  ('d0e1f2a3-4b5c-6d7e-8f9a-0b1c2d3e4f5a', 'member_mason',    'mason.member@example.com',       '$2b$10$f.2V43IZDKoq3YtB2fu4IOhK762U0BVydmntsnvOMlrJeygLpzZuW', 'MEMBER',    NOW(),         NOW());
```
-- connect to db on dbeaver
<img width="829" height="604" alt="image" src="https://github.com/user-attachments/assets/dec19d66-c7fc-4bdc-8178-1a58c93446f3" />

## Usage

```node
npm start
```

*open link on browser

*the documentation is on the left for reference

<img width="1926" height="961" alt="image" src="https://github.com/user-attachments/assets/857e779b-0659-4afd-a54c-4ac3b71a5279" />


*example usage:
```
mutation CreateUser {
  createUser(input: {username: "john", email: "johndoe@hello.com", password: "JohnDoe01@", role: MEMBER}) {
   code
   success
   message
   errors
   user {
     userId
     username
   }
  }
}
```
<img width="1278" height="430" alt="image" src="https://github.com/user-attachments/assets/29a5e620-2c99-4d9a-9724-1965c1e84394" />

```
mutation Login {
  login(input: {email: "johndoe@hello.com", password: "JohnDoe01@"}) {
    code
    success
    message
    errors
    accessToken
    refreshToken
  }
}
```
<img width="1279" height="477" alt="image" src="https://github.com/user-attachments/assets/1bbc187b-32dd-4422-b905-f360e2aa5fdd" />

```
query User {
  user(userId: "1be478f1-f842-40fa-8f0b-ad7ecfe5db3b") {
   email
   role
   userId
   username
   createdAt 
  }
}
```
<img width="1280" height="377" alt="image" src="https://github.com/user-attachments/assets/9998ffb0-b987-421d-8072-6399c8360302" />

```
query Users {
  users(role: MEMBER) {
    userId
    username
    email
    role
    createdAt
  }
}
```
<img width="1273" height="734" alt="image" src="https://github.com/user-attachments/assets/c8ffa731-ce0e-4455-93fb-fc6fc1ca3826" />
