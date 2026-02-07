---
title: "Authentication Flow"
description: "JWT + Refresh Token"
---

sequenceDiagram
    participant Client as Client
    participant Auth as Auth Service
    participant DB as Database
    participant Redis as Redis

    Note over Client,Redis: Login Flow
    
    Client->>Auth: POST /auth/login<br/>{email, password}
    Auth->>DB: SELECT user WHERE email=?
    Auth->>Auth: bcrypt.Compare(password, hash)
    
    alt Valid credentials
        Auth->>Auth: Generate JWT Access (15min)
        Auth->>Auth: Generate Refresh Token (7d)
        Auth->>DB: INSERT refresh_tokens
        Auth-->>Client: {accessToken, refreshToken}
    else Invalid
        Auth-->>Client: 401 Unauthorized
    end
    
    Note over Client,Redis: Refresh Token Flow
    
    Client->>Auth: POST /auth/refresh<br/>{refreshToken}
    Auth->>DB: SELECT refresh_token
    Auth->>Redis: Check blacklist
    
    alt Token valid
        Auth->>Auth: Generate new tokens
        Auth->>DB: DELETE old token
        Auth->>DB: INSERT new token
        Auth->>Redis: Blacklist old token
        Auth-->>Client: {new accessToken, refreshToken}
    else Invalid
        Auth-->>Client: 401 Token expired
    end
    
    Note over Client,Redis: Logout
    
    Client->>Auth: POST /auth/logout
    Auth->>DB: DELETE refresh_token
    Auth->>Redis: Blacklist token
    Auth-->>Client: 200 OK
