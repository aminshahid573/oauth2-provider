# OAuth2 Flow Walkthroughs

This document provides practical, step-by-step guides for developers integrating their applications with this OAuth2 provider.

## Flow 1: Authorization Code + PKCE

This is the recommended flow for **web applications, mobile apps, and SPAs**.

### Step 1: Generate PKCE Verifier and Challenge

The client application must first generate a secret (`code_verifier`) and its hashed version (`code_challenge`).

-   **`code_verifier`**: A high-entropy random string (e.g., 43 characters).
-   **`code_challenge`**: `BASE64URL-ENCODE(SHA256(code_verifier))`

### Step 2: Redirect User to Authorization Endpoint

The client redirects the user's browser to the `/oauth2/authorize` endpoint with the challenge.

**Example URL:**
```
http://localhost:8080/oauth2/authorize?
response_type=code
&client_id=test-client
&redirect_uri=http://localhost:3000/callback
&scope=openid%20profile
&state=random_state_string
&code_challenge=E9Mel...
&code_challenge_method=S256
```

### Step 3: User Logs In and Consents

The user interacts with the provider's login and consent pages. On approval, the provider redirects back to the client's `redirect_uri`.

**Example Redirect:**
```
http://localhost:3000/callback?code=a_one_time_auth_code&state=random_state_string
```

### Step 4: Exchange Code for Tokens

The client's backend makes a `POST` request to the `/oauth2/token` endpoint, including the `code_verifier` from Step 1.

**Example `curl` command:**
```bash
curl -X POST http://localhost:8080/oauth2/token \
-d "grant_type=authorization_code" \
-d "code=a_one_time_auth_code" \
-d "redirect_uri=http://localhost:3000/callback" \
-d "client_id=test-client" \
-d "client_secret=test-secret" \
-d "code_verifier=the_original_random_string"
```

**Expected Response:**
```json
{
  "access_token": "...",
  "refresh_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

---

## Flow 2: Client Credentials

Used for **server-to-server** communication.

### Step 1: Request Token

The client's backend makes a single `POST` request to the `/oauth2/token` endpoint.

**Example `curl` command:**
```bash
curl -X POST http://localhost:8080/oauth2/token \
-d "grant_type=client_credentials" \
-d "client_id=m2m-client" \
-d "client_secret=m2m-secret" \
-d "scope=api:read"
```

**Expected Response:**
```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "scope": "api:read"
}
```

---

## Flow 3: Device Authorization

Used for **input-constrained devices** like Smart TVs.

### Step 1: Device Requests Codes

The device makes a `POST` request to the `/oauth2/device_authorization` endpoint.

**Example `curl` command:**
```bash
curl -X POST http://localhost:8080/oauth2/device_authorization \
-d "client_id=test-client"
```

**Expected Response:**
```json
{
  "device_code": "long_secret_for_device",
  "user_code": "ABC-DEF",
  "verification_uri": "http://localhost:8080/device",
  "expires_in": 900,
  "interval": 5
}
```
The device displays the `user_code` to the user.

### Step 2: User Authorizes on a Secondary Device

The user goes to the `verification_uri` in their phone/laptop browser, logs in, enters the `user_code`, and grants consent.

### Step 3: Device Polls for Tokens

While the user is completing Step 2, the device repeatedly makes a `POST` request to the `/oauth2/token` endpoint every 5 seconds (`interval`).

**Example `curl` command:**
```bash
curl -X POST http://localhost:8080/oauth2/token \
-d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
-d "device_code=long_secret_for_device" \
-d "client_id=test-client"
```

-   **While Pending:** The server responds with a `400 Bad Request` and `{"error":"authorization_pending"}`.
-   **After Approval:** The server responds with a `200 OK` and the final tokens.
-  

---
| [![Previous](https://img.shields.io/badge/←_Previous-1f6feb?style=for-the-badge&logo=none&logoColor=white&labelColor=1f6feb&color=1f6feb)](API.md) <br> <sub>API.md</sub> | [![Next](https://img.shields.io/badge/Next_→-1f6feb?style=for-the-badge&logo=none&logoColor=white&labelColor=1f6feb&color=1f6feb)](DEPLOYMENT.md) <br> <sub>DEPLOYMENT.md</sub> |
|----------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------|