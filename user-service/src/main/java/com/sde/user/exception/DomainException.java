package com.sde.user.exception;

/**
 * Base de todas las excepciones de dominio del servicio.
 */
public abstract class DomainException extends RuntimeException {
    public DomainException(String message) {
        super(message);
    }
}
