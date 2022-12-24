export function isValidationError(
  error: ValidationError | Error | any
): error is ValidationError {
  return (
    (error as StorageError).errorType !== undefined &&
    error.errorType === "validation"
  );
}

export function isNotFoundError(
  error: NotFoundError | Error | any
): error is NotFoundError {
  return (
    (error as StorageError).errorType !== undefined &&
    error.errorType === "notfound"
  );
}

export function isSystemError(
  error: SystemError | Error | any
): error is SystemError {
  return (
    (error as StorageError).errorType !== undefined &&
    error.errorType === "system"
  );
}

export enum ErrorTypes {
  Validation = "validation",
  NotFound = "notfound",
  System = "system",
}

type ErrorType = "validation" | "notfound" | "system";

export class ValidationError extends Error {
  constructor(msg: string | undefined = undefined) {
    super(msg);
    Object.setPrototypeOf(this, ValidationError.prototype);
  }
  errorType: ErrorType = "validation";
}

export class NotFoundError extends Error {
  constructor(msg: string | undefined = undefined) {
    super(msg);
    Object.setPrototypeOf(this, NotFoundError.prototype);
  }
  errorType: ErrorType = "notfound";
}

export class SystemError extends Error {
  constructor(msg: string | undefined = undefined) {
    super(msg);
    Object.setPrototypeOf(this, SystemError.prototype);
  }
  errorType: ErrorType = "system";
}

export type StorageError = ValidationError | NotFoundError | SystemError;
