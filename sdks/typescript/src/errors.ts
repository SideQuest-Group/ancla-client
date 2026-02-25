/** Base error class for all Ancla SDK errors. */
export class AnclaError extends Error {
  /** HTTP status code, if available. */
  public readonly status: number;
  /** Raw response body, if available. */
  public readonly body: string;

  constructor(message: string, status: number, body = "") {
    super(message);
    this.name = "AnclaError";
    this.status = status;
    this.body = body;
  }
}

/** Thrown when the API returns 401 Unauthorized. */
export class AuthenticationError extends AnclaError {
  constructor(message = "Not authenticated", body = "") {
    super(message, 401, body);
    this.name = "AuthenticationError";
  }
}

/** Thrown when the API returns 404 Not Found. */
export class NotFoundError extends AnclaError {
  constructor(message = "Not found", body = "") {
    super(message, 404, body);
    this.name = "NotFoundError";
  }
}

/** Thrown when the API returns 422 Unprocessable Entity. */
export class ValidationError extends AnclaError {
  constructor(message = "Validation error", body = "") {
    super(message, 422, body);
    this.name = "ValidationError";
  }
}

/** Thrown when the API returns a 5xx status code. */
export class ServerError extends AnclaError {
  constructor(message = "Server error", body = "") {
    super(message, 500, body);
    this.name = "ServerError";
  }
}
