package com.sde.user.exception;

public class InvalidTokenException extends DomainException {
    public InvalidTokenException(String message) {
        super(message);
    }
}
