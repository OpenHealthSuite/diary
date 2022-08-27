
export function isValidationError(error: ValidationError | Error | any): error is ValidationError {
    return (error as ValidationError).validationError !== undefined && error.validationError;
}

export class ValidationError extends Error {
    constructor(msg: string) {
        super(msg);
        Object.setPrototypeOf(this, ValidationError.prototype);
    }
    validationError = true
}

export type StorageError = ValidationError | Error
