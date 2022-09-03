
export function isValidationError(error: ValidationError | Error | any): error is ValidationError {
    return (error as ValidationError).validationError !== undefined && error.validationError;
}

export function isNotFoundError(error: NotFoundError | Error | any): error is NotFoundError {
    return (error as NotFoundError).notFoundError !== undefined && error.notFoundError;
}

export function isSystemError(error: SystemError | Error | any): error is SystemError {
    return (error as SystemError).systemError !== undefined && error.systemError;
}

export class ValidationError extends Error {
    constructor(msg: string) {
        super(msg);
        Object.setPrototypeOf(this, ValidationError.prototype);
    }
    validationError = true
}


export class NotFoundError extends Error {
    constructor(msg: string) {
        super(msg);
        Object.setPrototypeOf(this, NotFoundError.prototype);
    }
    notFoundError = true
}

export class SystemError extends Error {
    constructor(msg: string) {
        super(msg);
        Object.setPrototypeOf(this, SystemError.prototype);
    }
    systemError = true
}

export type StorageError = ValidationError | NotFoundError | SystemError | Error
