# Echo Admin

Echo Admin provides browser-based administration for operators and token-based access for non-browser API clients.

## Language

**Administrator**:
A human operator who signs in to the administration console.
_Avoid_: User, account

**Login Session**:
A single browser/device login state for one **Administrator**.
_Avoid_: Admin Session, JWT session, global login

**API Token**:
A credential issued for non-browser API clients acting as an administrator role.
_Avoid_: Login Session, login token

**Session Revocation**:
The act of making one or more **Login Sessions** no longer valid.
_Avoid_: Token blacklist

**CSRF Token**:
A browser-readable request token required for state-changing browser API calls.
_Avoid_: Login token, session token

**Security Event**:
An administrator account change that requires **Session Revocation**.
_Avoid_: Normal logout

**System First Initialization**:
The one-time setup that makes a new Echo Admin installation usable for its first **Administrator**.
_Avoid_: Process startup, migration, seed

**Installation State**:
The persistent state that says whether **System First Initialization** has completed for an Echo Admin installation.
_Avoid_: Admin existence check, table existence check

**Root Role**:
The highest-authority administrator role created during **System First Initialization**.
_Avoid_: Custom setup role, normal role

## Relationships

- An **Administrator** may have zero or more active **Login Sessions**.
- A **Login Session** belongs to exactly one **Administrator**.
- A **Login Session** may hold the currently active role for that browser login.
- A **Login Session** does not contain a permission snapshot; authorization uses the current administrator and role state.
- A **Login Session** uses an HttpOnly cookie credential; browser JavaScript must not read the session token.
- A **CSRF Token** is not a **Login Session** and cannot authenticate a request by itself.
- An **API Token** is not a **Login Session**.
- Browser administration routes authenticate only with **Login Session** cookies.
- Machine API routes authenticate only with **API Tokens**.
- Browser login responses do not expose **Login Session** credentials in response bodies.
- Browser clients recover the current administrator view from the active **Login Session**, not from locally stored credentials.
- A **Security Event** may revoke all **Login Sessions** for one **Administrator**.
- Disabling, deleting, resetting the password for, or explicitly signing out an **Administrator** from all devices is a **Security Event**.
- When an **Administrator** changes their own password, other **Login Sessions** are revoked while the current **Login Session** remains active.
- Role and permission changes are not **Security Events** because authorization uses current administrator and role state.
- A normal logout revokes only one **Login Session**.
- Signing out from other devices revokes the other **Login Sessions** for the same **Administrator** while keeping the current **Login Session** active.
- An unavailable **Login Session** is an authentication failure; a valid **Login Session** without route permission is an authorization failure.
- **System First Initialization** happens before the first **Administrator** can use the administration console.
- **System First Initialization** creates the first **Administrator** with the **Root Role**.
- **Installation State** is the source of truth for whether **System First Initialization** is still allowed.
- **System First Initialization** may be retried until **Installation State** records completion.

## Example Dialogue

> **Dev:** "When an **Administrator** signs in on a second browser, should it replace the first login?"
> **Domain expert:** "No. Create another **Login Session**. Only explicit security events should revoke all sessions for that administrator."

## Flagged Ambiguities

- "session" was used to mean both browser login state and API credentials; resolved: browser login state is **Login Session**, while machine credentials are **API Token**.
