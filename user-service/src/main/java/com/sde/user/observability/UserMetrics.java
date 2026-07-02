package com.sde.user.observability;

import io.micrometer.core.instrument.Counter;
import io.micrometer.core.instrument.MeterRegistry;
import org.springframework.stereotype.Component;

@Component
public class UserMetrics {

    private final Counter usersCreated;
    private final Counter usersGetById;
    private final Counter usersUpdated;
    private final Counter loginSuccess;
    private final Counter loginFailed;
    private final Counter logoutSuccess;
    private final Counter tokenValidated;
    private final MeterRegistry meterRegistry;

    public UserMetrics(MeterRegistry meterRegistry) {
        this.meterRegistry = meterRegistry;
        this.usersCreated = Counter.builder("users.created")
                .description("Total de usuarios creados correctamente")
                .register(meterRegistry);
        this.usersGetById = Counter.builder("users.get.by.id")
                .description("Total de consultas de usuario por ID")
                .register(meterRegistry);
        this.usersUpdated = Counter.builder("users.updated")
                .description("Total de actualizaciones de perfil de usuario")
                .register(meterRegistry);
        this.loginSuccess = Counter.builder("users.login.success")
                .description("Total de autenticaciones exitosas")
                .register(meterRegistry);
        this.loginFailed = Counter.builder("users.login.failed")
                .description("Total de autenticaciones fallidas")
                .register(meterRegistry);
        this.logoutSuccess = Counter.builder("users.logout.success")
                .description("Total de cierres de sesion exitosos")
                .register(meterRegistry);
        this.tokenValidated = Counter.builder("users.token.validated")
                .description("Total de validaciones exitosas de token")
                .register(meterRegistry);
    }

    public void userCreated() {
        usersCreated.increment();
    }

    public void userGetById() {
        usersGetById.increment();
    }

    public void userUpdated() {
        usersUpdated.increment();
    }

    public void loginSuccess() {
        loginSuccess.increment();
    }

    public void loginFailed() {
        loginFailed.increment();
    }

    public void logoutSuccess() {
        logoutSuccess.increment();
    }

    public void tokenValidated() {
        tokenValidated.increment();
    }

    public void domainError(String operation, String errorType) {
        Counter.builder("users.domain.errors")
                .description("Errores de dominio del modulo de usuarios")
                .tag("operation", operation)
                .tag("error_type", errorType)
                .register(meterRegistry)
                .increment();
    }
}
