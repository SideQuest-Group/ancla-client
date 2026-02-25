"""Custom exceptions for the Ancla SDK."""


class AnclaError(Exception):
    """Base exception for all Ancla SDK errors.

    Attributes:
        status_code: The HTTP status code returned by the API, if applicable.
        detail: Additional detail from the API error response, if available.
    """

    def __init__(
        self,
        message: str,
        status_code: int | None = None,
        detail: str | None = None,
    ) -> None:
        self.status_code = status_code
        self.detail = detail
        super().__init__(message)


class AuthenticationError(AnclaError):
    """Raised when the API returns 401 Unauthorized."""


class NotFoundError(AnclaError):
    """Raised when the API returns 404 Not Found."""


class ValidationError(AnclaError):
    """Raised when the API returns 422 Unprocessable Entity."""


class ServerError(AnclaError):
    """Raised when the API returns a 5xx status code."""
