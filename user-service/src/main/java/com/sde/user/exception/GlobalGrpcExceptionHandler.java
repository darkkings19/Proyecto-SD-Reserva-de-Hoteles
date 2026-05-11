package com.sde.user.exception;

import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import net.devh.boot.grpc.server.advice.GrpcAdvice;
import net.devh.boot.grpc.server.advice.GrpcExceptionHandler;

/**
 * Interceptor global de errores para gRPC.
 * Atrapa las excepciones internas de negocio y las traduce automáticamente
 * a códigos de estado gRPC (Status codes) limpios para el cliente.
 */
@GrpcAdvice
public class GlobalGrpcExceptionHandler {

    @GrpcExceptionHandler(UserNotFoundException.class)
    public StatusRuntimeException handleUserNotFound(UserNotFoundException ex) {
        return Status.NOT_FOUND
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(EmailAlreadyExistsException.class)
    public StatusRuntimeException handleEmailAlreadyExists(EmailAlreadyExistsException ex) {
        return Status.ALREADY_EXISTS
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(InvalidCredentialsException.class)
    public StatusRuntimeException handleInvalidCredentials(InvalidCredentialsException ex) {
        return Status.UNAUTHENTICATED
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(IllegalArgumentException.class)
    public StatusRuntimeException handleIllegalArgument(IllegalArgumentException ex) {
        return Status.INVALID_ARGUMENT
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(Exception.class)
    public StatusRuntimeException handleGenericException(Exception ex) {
        // En un entorno productivo real aquí usaríamos un logger (ej. log.error)
        // Y nunca devolveríamos la traza entera al cliente por seguridad.
        return Status.INTERNAL
                .withDescription("Ha ocurrido un error interno en el servidor")
                .asRuntimeException();
    }
}
