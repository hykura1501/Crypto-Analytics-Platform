"""JWT validation for prediction endpoints. Requires ADMIN or VIP role."""

from fastapi import HTTPException, Security, status
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
import jwt

ALLOWED_ROLES = {"ADMIN", "VIP"}

# Optional Bearer so we can return clearer 401 messages
_security = HTTPBearer(auto_error=False)


def verify_prediction_access(credentials: HTTPAuthorizationCredentials | None = Security(_security)):
    """Validate Bearer token and require role in ADMIN or VIP."""
    from config import config

    if not credentials or not credentials.credentials:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authorization header with Bearer token required",
        )

    secret = config.api.jwt_secret
    token = credentials.credentials
    try:
        payload = jwt.decode(
            token,
            secret,
            algorithms=["HS256"],
            options={"verify_exp": True},
        )
    except jwt.ExpiredSignatureError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Token expired")
    except jwt.InvalidTokenError:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token (check JWT_SECRET matches auth-service)",
        )

    role = (payload.get("role") or "").strip() or "NORMAL"
    if role not in ALLOWED_ROLES:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="VIP or ADMIN role required to view AI predictions",
        )
    return {"user_id": payload.get("user_id"), "email": payload.get("email"), "role": role}
