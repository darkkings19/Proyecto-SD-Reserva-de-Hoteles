package com.sde.user.exception;

public class EmailAlreadyExistsException extends DomainException {
    public EmailAlreadyExistsException(String message) {
        super(message);
    }
}
