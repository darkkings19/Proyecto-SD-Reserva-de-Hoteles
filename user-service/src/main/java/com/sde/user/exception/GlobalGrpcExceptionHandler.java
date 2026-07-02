package com.sde.user.exception;

import com.sde.user.observability.UserMetrics;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import net.devh.boot.grpc.server.advice.GrpcAdvice;
import net.devh.boot.grpc.server.advice.GrpcExceptionHandler;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

@GrpcAdvice
public class GlobalGrpcExceptionHandler {

    private static final Logger log = LoggerFactory.getLogger(GlobalGrpcExceptionHandler.class);

    private final UserMetrics userMetrics;

    public GlobalGrpcExceptionHandler(UserMetrics userMetrics) {
        this.userMetrics = userMetrics;
    }

    @GrpcExceptionHandler(UserNotFoundException.class)
    public StatusRuntimeException handleUserNotFound(UserNotFoundException ex) {
        log.warn("Error gRPC NOT_FOUND en user-service: {}", ex.getMessage());
        return Status.NOT_FOUND
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(EmailAlreadyExistsException.class)
    public StatusRuntimeException handleEmailAlreadyExists(EmailAlreadyExistsException ex) {
        log.warn("Error gRPC ALREADY_EXISTS en user-service: {}", ex.getMessage());
        return Status.ALREADY_EXISTS
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(InvalidCredentialsException.class)
    public StatusRuntimeException handleInvalidCredentials(InvalidCredentialsException ex) {
        log.warn("Error gRPC UNAUTHENTICATED en user-service: {}", ex.getMessage());
        return Status.UNAUTHENTICATED
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(InvalidTokenException.class)
    public StatusRuntimeException handleInvalidToken(InvalidTokenException ex) {
        log.warn("Error gRPC UNAUTHENTICATED por token invalido en user-service: {}", ex.getMessage());
        return Status.UNAUTHENTICATED
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(IllegalArgumentException.class)
    public StatusRuntimeException handleIllegalArgument(IllegalArgumentException ex) {
        userMetrics.domainError("grpc_request", "invalid_argument");
        log.warn("Error gRPC INVALID_ARGUMENT en user-service: {}", ex.getMessage());
        return Status.INVALID_ARGUMENT
                .withDescription(ex.getMessage())
                .asRuntimeException();
    }

    @GrpcExceptionHandler(Exception.class)
    public StatusRuntimeException handleGenericException(Exception ex) {
        userMetrics.domainError("grpc_request", "internal_error");
        log.error("Error interno no controlado en user-service", ex);
        return Status.INTERNAL
                .withDescription("Ha ocurrido un error interno en el servidor")
                .asRuntimeException();
    }
}
